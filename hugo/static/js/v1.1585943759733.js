(function (superstore) {
    'use strict';

    function noop() { }
    function add_location(element, file, line, column, char) {
        element.__svelte_meta = {
            loc: { file, line, column, char }
        };
    }
    function run(fn) {
        return fn();
    }
    function blank_object() {
        return Object.create(null);
    }
    function run_all(fns) {
        fns.forEach(run);
    }
    function is_function(thing) {
        return typeof thing === 'function';
    }
    function safe_not_equal(a, b) {
        return a != a ? b == b : a !== b || ((a && typeof a === 'object') || typeof a === 'function');
    }
    function validate_store(store, name) {
        if (store != null && typeof store.subscribe !== 'function') {
            throw new Error(`'${name}' is not a store with a 'subscribe' method`);
        }
    }
    function subscribe(store, ...callbacks) {
        if (store == null) {
            return noop;
        }
        const unsub = store.subscribe(...callbacks);
        return unsub.unsubscribe ? () => unsub.unsubscribe() : unsub;
    }
    function component_subscribe(component, store, callback) {
        component.$$.on_destroy.push(subscribe(store, callback));
    }

    function append(target, node) {
        target.appendChild(node);
    }
    function insert(target, node, anchor) {
        target.insertBefore(node, anchor || null);
    }
    function detach(node) {
        node.parentNode.removeChild(node);
    }
    function element(name) {
        return document.createElement(name);
    }
    function svg_element(name) {
        return document.createElementNS('http://www.w3.org/2000/svg', name);
    }
    function text(data) {
        return document.createTextNode(data);
    }
    function space() {
        return text(' ');
    }
    function empty() {
        return text('');
    }
    function listen(node, event, handler, options) {
        node.addEventListener(event, handler, options);
        return () => node.removeEventListener(event, handler, options);
    }
    function attr(node, attribute, value) {
        if (value == null)
            node.removeAttribute(attribute);
        else if (node.getAttribute(attribute) !== value)
            node.setAttribute(attribute, value);
    }
    function children(element) {
        return Array.from(element.childNodes);
    }
    function set_style(node, key, value, important) {
        node.style.setProperty(key, value, important ? 'important' : '');
    }
    function toggle_class(element, name, toggle) {
        element.classList[toggle ? 'add' : 'remove'](name);
    }
    function custom_event(type, detail) {
        const e = document.createEvent('CustomEvent');
        e.initCustomEvent(type, false, false, detail);
        return e;
    }

    let current_component;
    function set_current_component(component) {
        current_component = component;
    }
    function get_current_component() {
        if (!current_component)
            throw new Error(`Function called outside component initialization`);
        return current_component;
    }
    function onMount(fn) {
        get_current_component().$$.on_mount.push(fn);
    }

    const dirty_components = [];
    const binding_callbacks = [];
    const render_callbacks = [];
    const flush_callbacks = [];
    const resolved_promise = Promise.resolve();
    let update_scheduled = false;
    function schedule_update() {
        if (!update_scheduled) {
            update_scheduled = true;
            resolved_promise.then(flush);
        }
    }
    function tick() {
        schedule_update();
        return resolved_promise;
    }
    function add_render_callback(fn) {
        render_callbacks.push(fn);
    }
    let flushing = false;
    const seen_callbacks = new Set();
    function flush() {
        if (flushing)
            return;
        flushing = true;
        do {
            // first, call beforeUpdate functions
            // and update components
            for (let i = 0; i < dirty_components.length; i += 1) {
                const component = dirty_components[i];
                set_current_component(component);
                update(component.$$);
            }
            dirty_components.length = 0;
            while (binding_callbacks.length)
                binding_callbacks.pop()();
            // then, once components are updated, call
            // afterUpdate functions. This may cause
            // subsequent updates...
            for (let i = 0; i < render_callbacks.length; i += 1) {
                const callback = render_callbacks[i];
                if (!seen_callbacks.has(callback)) {
                    // ...so guard against infinite loops
                    seen_callbacks.add(callback);
                    callback();
                }
            }
            render_callbacks.length = 0;
        } while (dirty_components.length);
        while (flush_callbacks.length) {
            flush_callbacks.pop()();
        }
        update_scheduled = false;
        flushing = false;
        seen_callbacks.clear();
    }
    function update($$) {
        if ($$.fragment !== null) {
            $$.update();
            run_all($$.before_update);
            const dirty = $$.dirty;
            $$.dirty = [-1];
            $$.fragment && $$.fragment.p($$.ctx, dirty);
            $$.after_update.forEach(add_render_callback);
        }
    }
    const outroing = new Set();
    function transition_in(block, local) {
        if (block && block.i) {
            outroing.delete(block);
            block.i(local);
        }
    }

    const globals = (typeof window !== 'undefined' ? window : global);
    function mount_component(component, target, anchor) {
        const { fragment, on_mount, on_destroy, after_update } = component.$$;
        fragment && fragment.m(target, anchor);
        // onMount happens before the initial afterUpdate
        add_render_callback(() => {
            const new_on_destroy = on_mount.map(run).filter(is_function);
            if (on_destroy) {
                on_destroy.push(...new_on_destroy);
            }
            else {
                // Edge case - component was destroyed immediately,
                // most likely as a result of a binding initialising
                run_all(new_on_destroy);
            }
            component.$$.on_mount = [];
        });
        after_update.forEach(add_render_callback);
    }
    function destroy_component(component, detaching) {
        const $$ = component.$$;
        if ($$.fragment !== null) {
            run_all($$.on_destroy);
            $$.fragment && $$.fragment.d(detaching);
            // TODO null out other refs, including component.$$ (but need to
            // preserve final state?)
            $$.on_destroy = $$.fragment = null;
            $$.ctx = [];
        }
    }
    function make_dirty(component, i) {
        if (component.$$.dirty[0] === -1) {
            dirty_components.push(component);
            schedule_update();
            component.$$.dirty.fill(0);
        }
        component.$$.dirty[(i / 31) | 0] |= (1 << (i % 31));
    }
    function init(component, options, instance, create_fragment, not_equal, props, dirty = [-1]) {
        const parent_component = current_component;
        set_current_component(component);
        const prop_values = options.props || {};
        const $$ = component.$$ = {
            fragment: null,
            ctx: null,
            // state
            props,
            update: noop,
            not_equal,
            bound: blank_object(),
            // lifecycle
            on_mount: [],
            on_destroy: [],
            before_update: [],
            after_update: [],
            context: new Map(parent_component ? parent_component.$$.context : []),
            // everything else
            callbacks: blank_object(),
            dirty
        };
        let ready = false;
        $$.ctx = instance
            ? instance(component, prop_values, (i, ret, ...rest) => {
                const value = rest.length ? rest[0] : ret;
                if ($$.ctx && not_equal($$.ctx[i], $$.ctx[i] = value)) {
                    if ($$.bound[i])
                        $$.bound[i](value);
                    if (ready)
                        make_dirty(component, i);
                }
                return ret;
            })
            : [];
        $$.update();
        ready = true;
        run_all($$.before_update);
        // `false` as a special case of no DOM component
        $$.fragment = create_fragment ? create_fragment($$.ctx) : false;
        if (options.target) {
            if (options.hydrate) {
                const nodes = children(options.target);
                // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                $$.fragment && $$.fragment.l(nodes);
                nodes.forEach(detach);
            }
            else {
                // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                $$.fragment && $$.fragment.c();
            }
            if (options.intro)
                transition_in(component.$$.fragment);
            mount_component(component, options.target, options.anchor);
            flush();
        }
        set_current_component(parent_component);
    }
    let SvelteElement;
    if (typeof HTMLElement === 'function') {
        SvelteElement = class extends HTMLElement {
            constructor() {
                super();
                this.attachShadow({ mode: 'open' });
            }
            connectedCallback() {
                // @ts-ignore todo: improve typings
                for (const key in this.$$.slotted) {
                    // @ts-ignore todo: improve typings
                    this.appendChild(this.$$.slotted[key]);
                }
            }
            attributeChangedCallback(attr, _oldValue, newValue) {
                this[attr] = newValue;
            }
            $destroy() {
                destroy_component(this, 1);
                this.$destroy = noop;
            }
            $on(type, callback) {
                // TODO should this delegate to addEventListener?
                const callbacks = (this.$$.callbacks[type] || (this.$$.callbacks[type] = []));
                callbacks.push(callback);
                return () => {
                    const index = callbacks.indexOf(callback);
                    if (index !== -1)
                        callbacks.splice(index, 1);
                };
            }
            $set() {
                // overridden by instance, if it has props
            }
        };
    }

    function dispatch_dev(type, detail) {
        document.dispatchEvent(custom_event(type, Object.assign({ version: '3.20.1' }, detail)));
    }
    function append_dev(target, node) {
        dispatch_dev("SvelteDOMInsert", { target, node });
        append(target, node);
    }
    function insert_dev(target, node, anchor) {
        dispatch_dev("SvelteDOMInsert", { target, node, anchor });
        insert(target, node, anchor);
    }
    function detach_dev(node) {
        dispatch_dev("SvelteDOMRemove", { node });
        detach(node);
    }
    function listen_dev(node, event, handler, options, has_prevent_default, has_stop_propagation) {
        const modifiers = options === true ? ["capture"] : options ? Array.from(Object.keys(options)) : [];
        if (has_prevent_default)
            modifiers.push('preventDefault');
        if (has_stop_propagation)
            modifiers.push('stopPropagation');
        dispatch_dev("SvelteDOMAddEventListener", { node, event, handler, modifiers });
        const dispose = listen(node, event, handler, options);
        return () => {
            dispatch_dev("SvelteDOMRemoveEventListener", { node, event, handler, modifiers });
            dispose();
        };
    }
    function attr_dev(node, attribute, value) {
        attr(node, attribute, value);
        if (value == null)
            dispatch_dev("SvelteDOMRemoveAttribute", { node, attribute });
        else
            dispatch_dev("SvelteDOMSetAttribute", { node, attribute, value });
    }
    function set_data_dev(text, data) {
        data = '' + data;
        if (text.data === data)
            return;
        dispatch_dev("SvelteDOMSetData", { node: text, data });
        text.data = data;
    }
    function validate_slots(name, slot, keys) {
        for (const slot_key of Object.keys(slot)) {
            if (!~keys.indexOf(slot_key)) {
                console.warn(`<${name}> received an unexpected slot "${slot_key}".`);
            }
        }
    }

    /* src/components/login_header.svelte generated by Svelte v3.20.1 */
    const file = "src/components/login_header.svelte";

    // (185:49) 
    function create_if_block_1(ctx) {
    	let a;
    	let t;

    	const block = {
    		c: function create() {
    			a = element("a");
    			t = text("Login");
    			attr_dev(a, "title", "Click to login");
    			attr_dev(a, "href", /*loginurl*/ ctx[0]);
    			attr_dev(a, "class", "f6 fw6 hover-red link black-70 mr2 mr3-m mr4-l dib");
    			add_location(a, file, 185, 4, 108775);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, a, anchor);
    			append_dev(a, t);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*loginurl*/ 1) {
    				attr_dev(a, "href", /*loginurl*/ ctx[0]);
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(a);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1.name,
    		type: "if",
    		source: "(185:49) ",
    		ctx
    	});

    	return block;
    }

    // (166:2) {#if loggedIn()}
    function create_if_block(ctx) {
    	let a0;
    	let t1;
    	let a1;
    	let t3;
    	let a2;

    	const block = {
    		c: function create() {
    			a0 = element("a");
    			a0.textContent = "Create";
    			t1 = space();
    			a1 = element("a");
    			a1.textContent = "My Lists";
    			t3 = space();
    			a2 = element("a");
    			a2.textContent = "Logout";
    			attr_dev(a0, "title", "Create");
    			attr_dev(a0, "href", "/create.html");
    			attr_dev(a0, "class", "f6 fw6 hover-blue link black-70 ml0 mr2-l di");
    			add_location(a0, file, 166, 4, 108312);
    			attr_dev(a1, "title", "Lists created by you");
    			attr_dev(a1, "href", "/lists-by-me.html");
    			attr_dev(a1, "class", "f6 fw6 hover-blue link black-70 di");
    			add_location(a1, file, 172, 4, 108448);
    			attr_dev(a2, "title", "Logout");
    			attr_dev(a2, "href", "/logout.html");
    			attr_dev(a2, "class", "f6 fw6 hover-blue link black-70 di ml3");
    			add_location(a2, file, 178, 4, 108595);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, a0, anchor);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, a1, anchor);
    			insert_dev(target, t3, anchor);
    			insert_dev(target, a2, anchor);
    		},
    		p: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(a0);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(a1);
    			if (detaching) detach_dev(t3);
    			if (detaching) detach_dev(a2);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block.name,
    		type: "if",
    		source: "(166:2) {#if loggedIn()}",
    		ctx
    	});

    	return block;
    }

    function create_fragment(ctx) {
    	let div;
    	let show_if;

    	function select_block_type(ctx, dirty) {
    		if (show_if == null) show_if = !!superstore.loggedIn();
    		if (show_if) return create_if_block;
    		if (window.location.pathname != /*loginurl*/ ctx[0]) return create_if_block_1;
    	}

    	let current_block_type = select_block_type(ctx);
    	let if_block = current_block_type && current_block_type(ctx);

    	const block = {
    		c: function create() {
    			div = element("div");
    			if (if_block) if_block.c();
    			this.c = noop;
    			attr_dev(div, "class", "fr mt0");
    			add_location(div, file, 164, 0, 108268);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div, anchor);
    			if (if_block) if_block.m(div, null);
    		},
    		p: function update(ctx, [dirty]) {
    			if (current_block_type === (current_block_type = select_block_type(ctx)) && if_block) {
    				if_block.p(ctx, dirty);
    			} else {
    				if (if_block) if_block.d(1);
    				if_block = current_block_type && current_block_type(ctx);

    				if (if_block) {
    					if_block.c();
    					if_block.m(div, null);
    				}
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);

    			if (if_block) {
    				if_block.d();
    			}
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance($$self, $$props, $$invalidate) {
    	let { loginurl = "/login.html" } = $$props;
    	const writable_props = ["loginurl"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	$$self.$set = $$props => {
    		if ("loginurl" in $$props) $$invalidate(0, loginurl = $$props.loginurl);
    	};

    	$$self.$capture_state = () => ({ loggedIn: superstore.loggedIn, loginurl });

    	$$self.$inject_state = $$props => {
    		if ("loginurl" in $$props) $$invalidate(0, loginurl = $$props.loginurl);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [loginurl];
    }

    class Login_header extends SvelteElement {
    	constructor(options) {
    		super();
    		this.shadowRoot.innerHTML = `<style>a{background-color:transparent}a,div{box-sizing:border-box}.di{display:inline}.dib{display:inline-block}.fr{_display:inline}.fr{float:right}.fw6{font-weight:600}.link{text-decoration:none}.link,.link:active,.link:focus,.link:hover,.link:link,.link:visited{transition:color .15s ease-in}.link:focus{outline:1px dotted currentColor}.black-70{color:rgba(0,0,0,.7)}.hover-red:focus,.hover-red:hover{color:#ff4136}.hover-blue:focus,.hover-blue:hover{color:#357edd}.ml0{margin-left:0}.ml3{margin-left:1rem}.mr2{margin-right:.5rem}.mt0{margin-top:0}.f6{font-size:.875rem}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){.mr3-m{margin-right:1rem}}@media screen and (min-width:60em){.mr2-l{margin-right:.5rem}.mr4-l{margin-right:2rem}}</style>`;
    		init(this, { target: this.shadowRoot }, instance, create_fragment, safe_not_equal, { loginurl: 0 });

    		if (options) {
    			if (options.target) {
    				insert_dev(options.target, this, options.anchor);
    			}

    			if (options.props) {
    				this.$set(options.props);
    				flush();
    			}
    		}
    	}

    	static get observedAttributes() {
    		return ["loginurl"];
    	}

    	get loginurl() {
    		return this.$$.ctx[0];
    	}

    	set loginurl(loginurl) {
    		this.$set({ loginurl });
    		flush();
    	}
    }

    /* src/components/banner/banner.svelte generated by Svelte v3.20.1 */

    const { console: console_1 } = globals;
    const file$1 = "src/components/banner/banner.svelte";

    // (193:0) {#if show}
    function create_if_block$1(ctx) {
    	let div;
    	let svg;
    	let title;
    	let t0;
    	let path;
    	let path_d_value;
    	let t1;
    	let span;
    	let t2;
    	let dispose;

    	const block = {
    		c: function create() {
    			div = element("div");
    			svg = svg_element("svg");
    			title = svg_element("title");
    			t0 = text("info icon");
    			path = svg_element("path");
    			t1 = space();
    			span = element("span");
    			t2 = text(/*message*/ ctx[2]);
    			add_location(title, file$1, 203, 6, 109572);
    			attr_dev(path, "d", path_d_value = /*getIcon*/ ctx[4](/*$notifications*/ ctx[1].level));
    			add_location(path, file$1, 204, 6, 109603);
    			attr_dev(svg, "class", "w1");
    			attr_dev(svg, "data-icon", "info");
    			attr_dev(svg, "viewBox", "0 0 24 24");
    			set_style(svg, "fill", "currentcolor");
    			set_style(svg, "width", "2em");
    			set_style(svg, "height", "2em");
    			add_location(svg, file$1, 198, 4, 109441);
    			attr_dev(span, "class", "lh-title ml3");
    			add_location(span, file$1, 206, 4, 109661);
    			attr_dev(div, "class", "flex items-center justify-center pa3 navy");
    			toggle_class(div, "info", /*level*/ ctx[0] === "info");
    			toggle_class(div, "error", /*level*/ ctx[0] === "error");
    			add_location(div, file$1, 193, 2, 109284);
    		},
    		m: function mount(target, anchor, remount) {
    			insert_dev(target, div, anchor);
    			append_dev(div, svg);
    			append_dev(svg, title);
    			append_dev(title, t0);
    			append_dev(svg, path);
    			append_dev(div, t1);
    			append_dev(div, span);
    			append_dev(span, t2);
    			if (remount) dispose();
    			dispose = listen_dev(div, "click", dismiss, false, false, false);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*$notifications*/ 2 && path_d_value !== (path_d_value = /*getIcon*/ ctx[4](/*$notifications*/ ctx[1].level))) {
    				attr_dev(path, "d", path_d_value);
    			}

    			if (dirty & /*message*/ 4) set_data_dev(t2, /*message*/ ctx[2]);

    			if (dirty & /*level*/ 1) {
    				toggle_class(div, "info", /*level*/ ctx[0] === "info");
    			}

    			if (dirty & /*level*/ 1) {
    				toggle_class(div, "error", /*level*/ ctx[0] === "error");
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$1.name,
    		type: "if",
    		source: "(193:0) {#if show}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$1(ctx) {
    	let if_block_anchor;
    	let if_block = /*show*/ ctx[3] && create_if_block$1(ctx);

    	const block = {
    		c: function create() {
    			if (if_block) if_block.c();
    			if_block_anchor = empty();
    			this.c = noop;
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			if (if_block) if_block.m(target, anchor);
    			insert_dev(target, if_block_anchor, anchor);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*show*/ ctx[3]) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$1(ctx);
    					if_block.c();
    					if_block.m(if_block_anchor.parentNode, if_block_anchor);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (if_block) if_block.d(detaching);
    			if (detaching) detach_dev(if_block_anchor);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$1.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function handleClick() {
    	console.log("HI");
    }

    function dismiss() {
    	superstore.notifications.clear();
    }

    function instance$1($$self, $$props, $$invalidate) {
    	let $notifications;
    	validate_store(superstore.notifications, "notifications");
    	component_subscribe($$self, superstore.notifications, $$value => $$invalidate(1, $notifications = $$value));
    	let infoIcon = `M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z`;
    	let errorIcon = `M11 15h2v2h-2zm0-8h2v6h-2zm.99-5C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z`;

    	function getIcon(level) {
    		if (level == "") {
    			return "";
    		}

    		return level == "info" ? infoIcon : errorIcon;
    	}

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	$$self.$capture_state = () => ({
    		notifications: superstore.notifications,
    		infoIcon,
    		errorIcon,
    		handleClick,
    		dismiss,
    		getIcon,
    		level,
    		$notifications,
    		message,
    		show
    	});

    	$$self.$inject_state = $$props => {
    		if ("infoIcon" in $$props) infoIcon = $$props.infoIcon;
    		if ("errorIcon" in $$props) errorIcon = $$props.errorIcon;
    		if ("level" in $$props) $$invalidate(0, level = $$props.level);
    		if ("message" in $$props) $$invalidate(2, message = $$props.message);
    		if ("show" in $$props) $$invalidate(3, show = $$props.show);
    	};

    	let level;
    	let message;
    	let show;

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*$notifications*/ 2) {
    			 $$invalidate(0, level = $notifications.level);
    		}

    		if ($$self.$$.dirty & /*$notifications*/ 2) {
    			 $$invalidate(2, message = $notifications.message);
    		}

    		if ($$self.$$.dirty & /*level*/ 1) {
    			 $$invalidate(3, show = level != "" ? true : false);
    		}
    	};

    	return [level, $notifications, message, show, getIcon];
    }

    class Banner extends SvelteElement {
    	constructor(options) {
    		super();
    		this.shadowRoot.innerHTML = `<style>div{box-sizing:border-box}.flex{display:flex}.items-center{align-items:center}.justify-center{justify-content:center}.lh-title{line-height:1.25}.w1{width:1rem}.navy{color:#001b44}.pa3{padding:1rem}.ml3{margin-left:1rem}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){}@media screen and (min-width:60em){}.error{background-color:#ffdfdf}.info{background-color:#96ccff}</style>`;
    		init(this, { target: this.shadowRoot }, instance$1, create_fragment$1, safe_not_equal, {});

    		if (options) {
    			if (options.target) {
    				insert_dev(options.target, this, options.anchor);
    			}
    		}
    	}
    }

    /* src/components/v1/Slideshow.svelte generated by Svelte v3.20.1 */

    const { console: console_1$1 } = globals;
    const file$2 = "src/components/v1/Slideshow.svelte";

    // (259:4) {#if loops > 0}
    function create_if_block$2(ctx) {
    	let cite;
    	let t0;
    	let t1;
    	let t2;

    	const block = {
    		c: function create() {
    			cite = element("cite");
    			t0 = text("- ");
    			t1 = text(/*loops*/ ctx[0]);
    			t2 = text(" (Looped over the list)");
    			attr_dev(cite, "class", "f6 ttu tracked fs-normal");
    			add_location(cite, file$2, 259, 6, 110416);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, cite, anchor);
    			append_dev(cite, t0);
    			append_dev(cite, t1);
    			append_dev(cite, t2);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*loops*/ 1) set_data_dev(t1, /*loops*/ ctx[0]);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(cite);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$2.name,
    		type: "if",
    		source: "(259:4) {#if loops > 0}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$2(ctx) {
    	let article;
    	let header;
    	let h1;
    	let t1;
    	let button0;
    	let t3;
    	let button1;
    	let t5;
    	let blockquote;
    	let p;
    	let t6;
    	let t7;
    	let dispose;
    	let if_block = /*loops*/ ctx[0] > 0 && create_if_block$2(ctx);

    	const block = {
    		c: function create() {
    			article = element("article");
    			header = element("header");
    			h1 = element("h1");
    			h1.textContent = "Slideshow";
    			t1 = space();
    			button0 = element("button");
    			button0.textContent = "Next";
    			t3 = space();
    			button1 = element("button");
    			button1.textContent = "Close";
    			t5 = space();
    			blockquote = element("blockquote");
    			p = element("p");
    			t6 = text(/*show*/ ctx[1]);
    			t7 = space();
    			if (if_block) if_block.c();
    			this.c = noop;
    			attr_dev(h1, "class", "f2 measure");
    			add_location(h1, file$2, 252, 4, 110094);
    			attr_dev(button0, "class", "br3");
    			add_location(button0, file$2, 253, 4, 110136);
    			attr_dev(button1, "class", "br3");
    			add_location(button1, file$2, 254, 4, 110193);
    			add_location(header, file$2, 251, 2, 110081);
    			attr_dev(p, "class", "f5 f4-m f3-l lh-copy measure mt0");
    			add_location(p, file$2, 257, 4, 110335);
    			attr_dev(blockquote, "class", "athelas ml0 mt4 pl4 black-90 bl bw2 b--black");
    			add_location(blockquote, file$2, 256, 2, 110265);
    			add_location(article, file$2, 250, 0, 110069);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor, remount) {
    			insert_dev(target, article, anchor);
    			append_dev(article, header);
    			append_dev(header, h1);
    			append_dev(header, t1);
    			append_dev(header, button0);
    			append_dev(header, t3);
    			append_dev(header, button1);
    			append_dev(article, t5);
    			append_dev(article, blockquote);
    			append_dev(blockquote, p);
    			append_dev(p, t6);
    			append_dev(blockquote, t7);
    			if (if_block) if_block.m(blockquote, null);
    			if (remount) run_all(dispose);

    			dispose = [
    				listen_dev(button0, "click", /*forward*/ ctx[2], false, false, false),
    				listen_dev(button1, "click", /*handleClose*/ ctx[3], false, false, false)
    			];
    		},
    		p: function update(ctx, [dirty]) {
    			if (dirty & /*show*/ 2) set_data_dev(t6, /*show*/ ctx[1]);

    			if (/*loops*/ ctx[0] > 0) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$2(ctx);
    					if_block.c();
    					if_block.m(blockquote, null);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(article);
    			if (if_block) if_block.d();
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$2.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$2($$self, $$props, $$invalidate) {
    	let listElement;

    	// {DomElement}
    	let playElement;

    	let { listdata } = $$props;
    	let aList = {};

    	onMount(async () => {
    		await tick();
    		aList = JSON.parse(listdata);
    	});

    	let loops = 0;
    	let index = -1;
    	let firstTime = "Welcome, to beginning, click next, or use the right arrow key..";
    	let show = firstTime;
    	let nextTimeIsLoop = 0;

    	function start(_listElement, _playElement) {
    		$$invalidate(1, show = firstTime);
    		$$invalidate(0, loops = 0);
    		index = -1;
    		nextTimeIsLoop = 0;
    		playElement = _playElement;
    		listElement = _listElement;
    		playElement.style.display = "";
    		listElement.style.display = "none";
    		window.addEventListener("keydown", handleKeydown);
    	}

    	function forward(event) {
    		index += 1;

    		if (!aList.data[index]) {
    			index = 0;
    			nextTimeIsLoop = 1;
    		}

    		if (nextTimeIsLoop) {
    			$$invalidate(0, loops += 1);
    			nextTimeIsLoop = 0;
    		}

    		$$invalidate(1, show = aList.data[index]);
    	}

    	function backward() {
    		index -= 1;

    		if (index >= 0) {
    			$$invalidate(1, show = aList.data[index]);
    		} else {
    			$$invalidate(1, show = firstTime);
    			index = -1;
    		}
    	}

    	function handleClose(event) {
    		window.removeEventListener("keydown", handleKeydown);
    		playElement.style.display = "none";
    		listElement.style.display = "";
    	}

    	function handleKeydown(event) {
    		switch (event.code) {
    			case "ArrowLeft":
    				backward();
    				break;
    			case "Space":
    			case "ArrowRight":
    				console.log("right");
    				forward();
    				break;
    			default:
    				console.log(event);
    				console.log(`pressed the ${event.key} key`);
    				break;
    		}
    	}

    	const writable_props = ["listdata"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$1.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	$$self.$set = $$props => {
    		if ("listdata" in $$props) $$invalidate(4, listdata = $$props.listdata);
    	};

    	$$self.$capture_state = () => ({
    		tick,
    		onMount,
    		listElement,
    		playElement,
    		listdata,
    		aList,
    		loops,
    		index,
    		firstTime,
    		show,
    		nextTimeIsLoop,
    		start,
    		forward,
    		backward,
    		handleClose,
    		handleKeydown
    	});

    	$$self.$inject_state = $$props => {
    		if ("listElement" in $$props) listElement = $$props.listElement;
    		if ("playElement" in $$props) playElement = $$props.playElement;
    		if ("listdata" in $$props) $$invalidate(4, listdata = $$props.listdata);
    		if ("aList" in $$props) aList = $$props.aList;
    		if ("loops" in $$props) $$invalidate(0, loops = $$props.loops);
    		if ("index" in $$props) index = $$props.index;
    		if ("firstTime" in $$props) firstTime = $$props.firstTime;
    		if ("show" in $$props) $$invalidate(1, show = $$props.show);
    		if ("nextTimeIsLoop" in $$props) nextTimeIsLoop = $$props.nextTimeIsLoop;
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [loops, show, forward, handleClose, listdata, start];
    }

    class Slideshow extends SvelteElement {
    	constructor(options) {
    		super();
    		this.shadowRoot.innerHTML = `<style>h1{font-size:2em;margin:.67em 0}button{font-family:inherit;font-size:100%;line-height:1.15;margin:0}button{overflow:visible}button{text-transform:none}button{-webkit-appearance:button}button::-moz-focus-inner{border-style:none;padding:0}button:-moz-focusring{outline:1px dotted ButtonText}article,blockquote,h1,header,p{box-sizing:border-box}.bl{border-left-style:solid;border-left-width:1px}.b--black{border-color:#000}.br3{border-radius:.5rem}.bw2{border-width:.25rem}.athelas{font-family:athelas,georgia,serif}.fs-normal{font-style:normal}.tracked{letter-spacing:.1em}.lh-copy{line-height:1.5}.black-90{color:rgba(0,0,0,.9)}.pl4{padding-left:2rem}.ml0{margin-left:0}.mt0{margin-top:0}.mt4{margin-top:2rem}.ttu{text-transform:uppercase}.f2{font-size:2.25rem}.f5{font-size:1rem}.f6{font-size:.875rem}.measure{max-width:30em}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){.f4-m{font-size:1.25rem}}@media screen and (min-width:60em){.f3-l{font-size:1.5rem}}</style>`;
    		init(this, { target: this.shadowRoot }, instance$2, create_fragment$2, safe_not_equal, { listdata: 4, start: 5 });
    		const { ctx } = this.$$;
    		const props = this.attributes;

    		if (/*listdata*/ ctx[4] === undefined && !("listdata" in props)) {
    			console_1$1.warn("<undefined> was created without expected prop 'listdata'");
    		}

    		if (options) {
    			if (options.target) {
    				insert_dev(options.target, this, options.anchor);
    			}

    			if (options.props) {
    				this.$set(options.props);
    				flush();
    			}
    		}
    	}

    	static get observedAttributes() {
    		return ["listdata", "start"];
    	}

    	get listdata() {
    		return this.$$.ctx[4];
    	}

    	set listdata(listdata) {
    		this.$set({ listdata });
    		flush();
    	}

    	get start() {
    		return this.$$.ctx[5];
    	}

    	set start(value) {
    		throw new Error("<undefined>: Cannot set read-only property 'start'");
    	}
    }

    /* src/components/v1/Menu.svelte generated by Svelte v3.20.1 */

    const file$3 = "src/components/v1/Menu.svelte";

    // (174:0) {#if slideshow == '1'}
    function create_if_block$3(ctx) {
    	let button;
    	let dispose;

    	const block = {
    		c: function create() {
    			button = element("button");
    			button.textContent = "Slideshow";
    			attr_dev(button, "class", "br3");
    			add_location(button, file$3, 174, 2, 108507);
    		},
    		m: function mount(target, anchor, remount) {
    			insert_dev(target, button, anchor);
    			if (remount) dispose();
    			dispose = listen_dev(button, "click", /*showSlideshow*/ ctx[1], false, false, false);
    		},
    		p: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(button);
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$3.name,
    		type: "if",
    		source: "(174:0) {#if slideshow == '1'}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$3(ctx) {
    	let if_block_anchor;
    	let if_block = /*slideshow*/ ctx[0] == "1" && create_if_block$3(ctx);

    	const block = {
    		c: function create() {
    			if (if_block) if_block.c();
    			if_block_anchor = empty();
    			this.c = noop;
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			if (if_block) if_block.m(target, anchor);
    			insert_dev(target, if_block_anchor, anchor);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*slideshow*/ ctx[0] == "1") {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$3(ctx);
    					if_block.c();
    					if_block.m(if_block_anchor.parentNode, if_block_anchor);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (if_block) if_block.d(detaching);
    			if (detaching) detach_dev(if_block_anchor);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$3.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$3($$self, $$props, $$invalidate) {
    	let { slideshow = "1" } = $$props;
    	let { playScreen = "#play" } = $$props;
    	let { infoScreen = "#list-info" } = $$props;

    	function showSlideshow(event) {
    		document.querySelector("v1-slideshow").start(document.querySelector(infoScreen), document.querySelector(playScreen));
    	}

    	const writable_props = ["slideshow", "playScreen", "infoScreen"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	$$self.$set = $$props => {
    		if ("slideshow" in $$props) $$invalidate(0, slideshow = $$props.slideshow);
    		if ("playScreen" in $$props) $$invalidate(2, playScreen = $$props.playScreen);
    		if ("infoScreen" in $$props) $$invalidate(3, infoScreen = $$props.infoScreen);
    	};

    	$$self.$capture_state = () => ({
    		slideshow,
    		playScreen,
    		infoScreen,
    		showSlideshow
    	});

    	$$self.$inject_state = $$props => {
    		if ("slideshow" in $$props) $$invalidate(0, slideshow = $$props.slideshow);
    		if ("playScreen" in $$props) $$invalidate(2, playScreen = $$props.playScreen);
    		if ("infoScreen" in $$props) $$invalidate(3, infoScreen = $$props.infoScreen);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [slideshow, showSlideshow, playScreen, infoScreen];
    }

    class Menu extends SvelteElement {
    	constructor(options) {
    		super();
    		this.shadowRoot.innerHTML = `<style>button{font-family:inherit;font-size:100%;line-height:1.15;margin:0}button{overflow:visible}button{text-transform:none}button{-webkit-appearance:button}button::-moz-focus-inner{border-style:none;padding:0}button:-moz-focusring{outline:1px dotted ButtonText}.br3{border-radius:.5rem}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){}@media screen and (min-width:60em){}</style>`;

    		init(this, { target: this.shadowRoot }, instance$3, create_fragment$3, safe_not_equal, {
    			slideshow: 0,
    			playScreen: 2,
    			infoScreen: 3
    		});

    		if (options) {
    			if (options.target) {
    				insert_dev(options.target, this, options.anchor);
    			}

    			if (options.props) {
    				this.$set(options.props);
    				flush();
    			}
    		}
    	}

    	static get observedAttributes() {
    		return ["slideshow", "playScreen", "infoScreen"];
    	}

    	get slideshow() {
    		return this.$$.ctx[0];
    	}

    	set slideshow(slideshow) {
    		this.$set({ slideshow });
    		flush();
    	}

    	get playScreen() {
    		return this.$$.ctx[2];
    	}

    	set playScreen(playScreen) {
    		this.$set({ playScreen });
    		flush();
    	}

    	get infoScreen() {
    		return this.$$.ctx[3];
    	}

    	set infoScreen(infoScreen) {
    		this.$set({ infoScreen });
    		flush();
    	}
    }

    customElements.define('login-header', Login_header);
    customElements.define('notification-center', Banner);
    customElements.define('v1-menu', Menu);
    customElements.define('v1-slideshow', Slideshow);

}(superstore));
//# sourceMappingURL=v1.1585943759733.js.map
