var v2 = (function () {
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
    function destroy_each(iterations, detaching) {
        for (let i = 0; i < iterations.length; i += 1) {
            if (iterations[i])
                iterations[i].d(detaching);
        }
    }
    function element(name) {
        return document.createElement(name);
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
    function to_number(value) {
        return value === '' ? undefined : +value;
    }
    function children(element) {
        return Array.from(element.childNodes);
    }
    function set_input_value(input, value) {
        if (value != null || input.value) {
            input.value = value;
        }
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
    function createEventDispatcher() {
        const component = get_current_component();
        return (type, detail) => {
            const callbacks = component.$$.callbacks[type];
            if (callbacks) {
                // TODO are there situations where events could be dispatched
                // in a server (non-DOM) environment?
                const event = custom_event(type, detail);
                callbacks.slice().forEach(fn => {
                    fn.call(component, event);
                });
            }
        };
    }
    // TODO figure out if we still want to support
    // shorthand events, or if we want to implement
    // a real bubbling mechanism
    function bubble(component, event) {
        const callbacks = component.$$.callbacks[event.type];
        if (callbacks) {
            callbacks.slice().forEach(fn => fn(event));
        }
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
    let outros;
    function group_outros() {
        outros = {
            r: 0,
            c: [],
            p: outros // parent group
        };
    }
    function check_outros() {
        if (!outros.r) {
            run_all(outros.c);
        }
        outros = outros.p;
    }
    function transition_in(block, local) {
        if (block && block.i) {
            outroing.delete(block);
            block.i(local);
        }
    }
    function transition_out(block, local, detach, callback) {
        if (block && block.o) {
            if (outroing.has(block))
                return;
            outroing.add(block);
            outros.c.push(() => {
                outroing.delete(block);
                if (callback) {
                    if (detach)
                        block.d(1);
                    callback();
                }
            });
            block.o(local);
        }
    }

    const globals = (typeof window !== 'undefined'
        ? window
        : typeof globalThis !== 'undefined'
            ? globalThis
            : global);
    function create_component(block) {
        block && block.c();
    }
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
        document.dispatchEvent(custom_event(type, Object.assign({ version: '3.22.2' }, detail)));
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
    function prop_dev(node, property, value) {
        node[property] = value;
        dispatch_dev("SvelteDOMSetProperty", { node, property, value });
    }
    function set_data_dev(text, data) {
        data = '' + data;
        if (text.data === data)
            return;
        dispatch_dev("SvelteDOMSetData", { node: text, data });
        text.data = data;
    }
    function validate_each_argument(arg) {
        if (typeof arg !== 'string' && !(arg && typeof arg === 'object' && 'length' in arg)) {
            let msg = '{#each} only iterates over array-like objects.';
            if (typeof Symbol === 'function' && arg && Symbol.iterator in arg) {
                msg += ' You can use a spread to convert this iterable into an array.';
            }
            throw new Error(msg);
        }
    }
    function validate_slots(name, slot, keys) {
        for (const slot_key of Object.keys(slot)) {
            if (!~keys.indexOf(slot_key)) {
                console.warn(`<${name}> received an unexpected slot "${slot_key}".`);
            }
        }
    }

    /* src/components/interact/menu.wc.svelte generated by Svelte v3.22.2 */

    const file = "src/components/interact/menu.wc.svelte";

    // (170:0) {#if slideshow == '1'}
    function create_if_block_1(ctx) {
    	let button;
    	let dispose;

    	const block = {
    		c: function create() {
    			button = element("button");
    			button.textContent = "Slideshow";
    			attr_dev(button, "class", "br3");
    			add_location(button, file, 170, 2, 108992);
    		},
    		m: function mount(target, anchor, remount) {
    			insert_dev(target, button, anchor);
    			if (remount) dispose();
    			dispose = listen_dev(button, "click", showSlideshow, false, false, false);
    		},
    		p: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(button);
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1.name,
    		type: "if",
    		source: "(170:0) {#if slideshow == '1'}",
    		ctx
    	});

    	return block;
    }

    // (174:0) {#if totalrecall == '1'}
    function create_if_block(ctx) {
    	let button;
    	let dispose;

    	const block = {
    		c: function create() {
    			button = element("button");
    			button.textContent = "Total Recall";
    			attr_dev(button, "class", "br3");
    			add_location(button, file, 174, 2, 109090);
    		},
    		m: function mount(target, anchor, remount) {
    			insert_dev(target, button, anchor);
    			if (remount) dispose();
    			dispose = listen_dev(button, "click", showTotalRecall, false, false, false);
    		},
    		p: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(button);
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block.name,
    		type: "if",
    		source: "(174:0) {#if totalrecall == '1'}",
    		ctx
    	});

    	return block;
    }

    function create_fragment(ctx) {
    	let t;
    	let if_block1_anchor;
    	let if_block0 = /*slideshow*/ ctx[0] == "1" && create_if_block_1(ctx);
    	let if_block1 = /*totalrecall*/ ctx[1] == "1" && create_if_block(ctx);

    	const block = {
    		c: function create() {
    			if (if_block0) if_block0.c();
    			t = space();
    			if (if_block1) if_block1.c();
    			if_block1_anchor = empty();
    			this.c = noop;
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			if (if_block0) if_block0.m(target, anchor);
    			insert_dev(target, t, anchor);
    			if (if_block1) if_block1.m(target, anchor);
    			insert_dev(target, if_block1_anchor, anchor);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*slideshow*/ ctx[0] == "1") {
    				if (if_block0) {
    					if_block0.p(ctx, dirty);
    				} else {
    					if_block0 = create_if_block_1(ctx);
    					if_block0.c();
    					if_block0.m(t.parentNode, t);
    				}
    			} else if (if_block0) {
    				if_block0.d(1);
    				if_block0 = null;
    			}

    			if (/*totalrecall*/ ctx[1] == "1") {
    				if (if_block1) {
    					if_block1.p(ctx, dirty);
    				} else {
    					if_block1 = create_if_block(ctx);
    					if_block1.c();
    					if_block1.m(if_block1_anchor.parentNode, if_block1_anchor);
    				}
    			} else if (if_block1) {
    				if_block1.d(1);
    				if_block1 = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (if_block0) if_block0.d(detaching);
    			if (detaching) detach_dev(t);
    			if (if_block1) if_block1.d(detaching);
    			if (detaching) detach_dev(if_block1_anchor);
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

    function showSlideshow(event) {
    	window.location.hash = "#/play/slideshow";
    }

    function showTotalRecall(event) {
    	window.location.hash = "#/play/total_recall";
    }

    function instance($$self, $$props, $$invalidate) {
    	let { slideshow = "0" } = $$props;
    	let { totalrecall = "0" } = $$props;
    	const writable_props = ["slideshow", "totalrecall"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	$$self.$set = $$props => {
    		if ("slideshow" in $$props) $$invalidate(0, slideshow = $$props.slideshow);
    		if ("totalrecall" in $$props) $$invalidate(1, totalrecall = $$props.totalrecall);
    	};

    	$$self.$capture_state = () => ({
    		slideshow,
    		totalrecall,
    		showSlideshow,
    		showTotalRecall
    	});

    	$$self.$inject_state = $$props => {
    		if ("slideshow" in $$props) $$invalidate(0, slideshow = $$props.slideshow);
    		if ("totalrecall" in $$props) $$invalidate(1, totalrecall = $$props.totalrecall);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [slideshow, totalrecall];
    }

    class Menu_wc extends SvelteElement {
    	constructor(options) {
    		super();
    		this.shadowRoot.innerHTML = `<style>button{font-family:inherit;font-size:100%;line-height:1.15;margin:0}button{overflow:visible}button{text-transform:none}button{-webkit-appearance:button}button::-moz-focus-inner{border-style:none;padding:0}button:-moz-focusring{outline:1px dotted ButtonText}.br3{border-radius:.5rem}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){}@media screen and (min-width:60em){}</style>`;
    		init(this, { target: this.shadowRoot }, instance, create_fragment, safe_not_equal, { slideshow: 0, totalrecall: 1 });

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
    		return ["slideshow", "totalrecall"];
    	}

    	get slideshow() {
    		return this.$$.ctx[0];
    	}

    	set slideshow(slideshow) {
    		this.$set({ slideshow });
    		flush();
    	}

    	get totalrecall() {
    		return this.$$.ctx[1];
    	}

    	set totalrecall(totalrecall) {
    		this.$set({ totalrecall });
    		flush();
    	}
    }

    const subscriber_queue = [];
    /**
     * Creates a `Readable` store that allows reading by subscription.
     * @param value initial value
     * @param {StartStopNotifier}start start and stop notifications for subscriptions
     */
    function readable(value, start) {
        return {
            subscribe: writable(value, start).subscribe,
        };
    }
    /**
     * Create a `Writable` store that allows both updating and reading by subscription.
     * @param {*=}value initial value
     * @param {StartStopNotifier=}start start and stop notifications for subscriptions
     */
    function writable(value, start = noop) {
        let stop;
        const subscribers = [];
        function set(new_value) {
            if (safe_not_equal(value, new_value)) {
                value = new_value;
                if (stop) { // store is ready
                    const run_queue = !subscriber_queue.length;
                    for (let i = 0; i < subscribers.length; i += 1) {
                        const s = subscribers[i];
                        s[1]();
                        subscriber_queue.push(s, value);
                    }
                    if (run_queue) {
                        for (let i = 0; i < subscriber_queue.length; i += 2) {
                            subscriber_queue[i][0](subscriber_queue[i + 1]);
                        }
                        subscriber_queue.length = 0;
                    }
                }
            }
        }
        function update(fn) {
            set(fn(value));
        }
        function subscribe(run, invalidate = noop) {
            const subscriber = [run, invalidate];
            subscribers.push(subscriber);
            if (subscribers.length === 1) {
                stop = start(set) || noop;
            }
            run(value);
            return () => {
                const index = subscribers.indexOf(subscriber);
                if (index !== -1) {
                    subscribers.splice(index, 1);
                }
                if (subscribers.length === 0) {
                    stop();
                    stop = null;
                }
            };
        }
        return { set, update, subscribe };
    }
    function derived(stores, fn, initial_value) {
        const single = !Array.isArray(stores);
        const stores_array = single
            ? [stores]
            : stores;
        const auto = fn.length < 2;
        return readable(initial_value, (set) => {
            let inited = false;
            const values = [];
            let pending = 0;
            let cleanup = noop;
            const sync = () => {
                if (pending) {
                    return;
                }
                cleanup();
                const result = fn(single ? values[0] : values, set);
                if (auto) {
                    set(result);
                }
                else {
                    cleanup = is_function(result) ? result : noop;
                }
            };
            const unsubscribers = stores_array.map((store, i) => subscribe(store, (value) => {
                values[i] = value;
                pending &= ~(1 << i);
                if (inited) {
                    sync();
                }
            }, () => {
                pending |= (1 << i);
            }));
            inited = true;
            sync();
            return function stop() {
                run_all(unsubscribers);
                cleanup();
            };
        });
    }

    function regexparam (str, loose) {
    	if (str instanceof RegExp) return { keys:false, pattern:str };
    	var c, o, tmp, ext, keys=[], pattern='', arr = str.split('/');
    	arr[0] || arr.shift();

    	while (tmp = arr.shift()) {
    		c = tmp[0];
    		if (c === '*') {
    			keys.push('wild');
    			pattern += '/(.*)';
    		} else if (c === ':') {
    			o = tmp.indexOf('?', 1);
    			ext = tmp.indexOf('.', 1);
    			keys.push( tmp.substring(1, !!~o ? o : !!~ext ? ext : tmp.length) );
    			pattern += !!~o && !~ext ? '(?:/([^/]+?))?' : '/([^/]+?)';
    			if (!!~ext) pattern += (!!~o ? '?' : '') + '\\' + tmp.substring(ext);
    		} else {
    			pattern += '/' + tmp;
    		}
    	}

    	return {
    		keys: keys,
    		pattern: new RegExp('^' + pattern + (loose ? '(?=$|\/)' : '\/?$'), 'i')
    	};
    }

    /* node_modules/svelte-spa-router/Router.svelte generated by Svelte v3.22.2 */

    const { Error: Error_1, Object: Object_1, console: console_1 } = globals;

    // (209:0) {:else}
    function create_else_block(ctx) {
    	let switch_instance_anchor;
    	let current;
    	var switch_value = /*component*/ ctx[0];

    	function switch_props(ctx) {
    		return { $$inline: true };
    	}

    	if (switch_value) {
    		var switch_instance = new switch_value(switch_props());
    		switch_instance.$on("routeEvent", /*routeEvent_handler_1*/ ctx[10]);
    	}

    	const block = {
    		c: function create() {
    			if (switch_instance) create_component(switch_instance.$$.fragment);
    			switch_instance_anchor = empty();
    		},
    		m: function mount(target, anchor) {
    			if (switch_instance) {
    				mount_component(switch_instance, target, anchor);
    			}

    			insert_dev(target, switch_instance_anchor, anchor);
    			current = true;
    		},
    		p: function update(ctx, dirty) {
    			if (switch_value !== (switch_value = /*component*/ ctx[0])) {
    				if (switch_instance) {
    					group_outros();
    					const old_component = switch_instance;

    					transition_out(old_component.$$.fragment, 1, 0, () => {
    						destroy_component(old_component, 1);
    					});

    					check_outros();
    				}

    				if (switch_value) {
    					switch_instance = new switch_value(switch_props());
    					switch_instance.$on("routeEvent", /*routeEvent_handler_1*/ ctx[10]);
    					create_component(switch_instance.$$.fragment);
    					transition_in(switch_instance.$$.fragment, 1);
    					mount_component(switch_instance, switch_instance_anchor.parentNode, switch_instance_anchor);
    				} else {
    					switch_instance = null;
    				}
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			if (switch_instance) transition_in(switch_instance.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			if (switch_instance) transition_out(switch_instance.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(switch_instance_anchor);
    			if (switch_instance) destroy_component(switch_instance, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_else_block.name,
    		type: "else",
    		source: "(209:0) {:else}",
    		ctx
    	});

    	return block;
    }

    // (207:0) {#if componentParams}
    function create_if_block$1(ctx) {
    	let switch_instance_anchor;
    	let current;
    	var switch_value = /*component*/ ctx[0];

    	function switch_props(ctx) {
    		return {
    			props: { params: /*componentParams*/ ctx[1] },
    			$$inline: true
    		};
    	}

    	if (switch_value) {
    		var switch_instance = new switch_value(switch_props(ctx));
    		switch_instance.$on("routeEvent", /*routeEvent_handler*/ ctx[9]);
    	}

    	const block = {
    		c: function create() {
    			if (switch_instance) create_component(switch_instance.$$.fragment);
    			switch_instance_anchor = empty();
    		},
    		m: function mount(target, anchor) {
    			if (switch_instance) {
    				mount_component(switch_instance, target, anchor);
    			}

    			insert_dev(target, switch_instance_anchor, anchor);
    			current = true;
    		},
    		p: function update(ctx, dirty) {
    			const switch_instance_changes = {};
    			if (dirty & /*componentParams*/ 2) switch_instance_changes.params = /*componentParams*/ ctx[1];

    			if (switch_value !== (switch_value = /*component*/ ctx[0])) {
    				if (switch_instance) {
    					group_outros();
    					const old_component = switch_instance;

    					transition_out(old_component.$$.fragment, 1, 0, () => {
    						destroy_component(old_component, 1);
    					});

    					check_outros();
    				}

    				if (switch_value) {
    					switch_instance = new switch_value(switch_props(ctx));
    					switch_instance.$on("routeEvent", /*routeEvent_handler*/ ctx[9]);
    					create_component(switch_instance.$$.fragment);
    					transition_in(switch_instance.$$.fragment, 1);
    					mount_component(switch_instance, switch_instance_anchor.parentNode, switch_instance_anchor);
    				} else {
    					switch_instance = null;
    				}
    			} else if (switch_value) {
    				switch_instance.$set(switch_instance_changes);
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			if (switch_instance) transition_in(switch_instance.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			if (switch_instance) transition_out(switch_instance.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(switch_instance_anchor);
    			if (switch_instance) destroy_component(switch_instance, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$1.name,
    		type: "if",
    		source: "(207:0) {#if componentParams}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$1(ctx) {
    	let current_block_type_index;
    	let if_block;
    	let if_block_anchor;
    	let current;
    	const if_block_creators = [create_if_block$1, create_else_block];
    	const if_blocks = [];

    	function select_block_type(ctx, dirty) {
    		if (/*componentParams*/ ctx[1]) return 0;
    		return 1;
    	}

    	current_block_type_index = select_block_type(ctx);
    	if_block = if_blocks[current_block_type_index] = if_block_creators[current_block_type_index](ctx);

    	const block = {
    		c: function create() {
    			if_block.c();
    			if_block_anchor = empty();
    			this.c = noop;
    		},
    		l: function claim(nodes) {
    			throw new Error_1("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			if_blocks[current_block_type_index].m(target, anchor);
    			insert_dev(target, if_block_anchor, anchor);
    			current = true;
    		},
    		p: function update(ctx, [dirty]) {
    			let previous_block_index = current_block_type_index;
    			current_block_type_index = select_block_type(ctx);

    			if (current_block_type_index === previous_block_index) {
    				if_blocks[current_block_type_index].p(ctx, dirty);
    			} else {
    				group_outros();

    				transition_out(if_blocks[previous_block_index], 1, 1, () => {
    					if_blocks[previous_block_index] = null;
    				});

    				check_outros();
    				if_block = if_blocks[current_block_type_index];

    				if (!if_block) {
    					if_block = if_blocks[current_block_type_index] = if_block_creators[current_block_type_index](ctx);
    					if_block.c();
    				}

    				transition_in(if_block, 1);
    				if_block.m(if_block_anchor.parentNode, if_block_anchor);
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(if_block);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(if_block);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if_blocks[current_block_type_index].d(detaching);
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

    function wrap(route, userData, ...conditions) {
    	// Check if we don't have userData
    	if (userData && typeof userData == "function") {
    		conditions = conditions && conditions.length ? conditions : [];
    		conditions.unshift(userData);
    		userData = undefined;
    	}

    	// Parameter route and each item of conditions must be functions
    	if (!route || typeof route != "function") {
    		throw Error("Invalid parameter route");
    	}

    	if (conditions && conditions.length) {
    		for (let i = 0; i < conditions.length; i++) {
    			if (!conditions[i] || typeof conditions[i] != "function") {
    				throw Error("Invalid parameter conditions[" + i + "]");
    			}
    		}
    	}

    	// Returns an object that contains all the functions to execute too
    	const obj = { route, userData };

    	if (conditions && conditions.length) {
    		obj.conditions = conditions;
    	}

    	// The _sveltesparouter flag is to confirm the object was created by this router
    	Object.defineProperty(obj, "_sveltesparouter", { value: true });

    	return obj;
    }

    /**
     * @typedef {Object} Location
     * @property {string} location - Location (page/view), for example `/book`
     * @property {string} [querystring] - Querystring from the hash, as a string not parsed
     */
    /**
     * Returns the current location from the hash.
     *
     * @returns {Location} Location object
     * @private
     */
    function getLocation() {
    	const hashPosition = window.location.href.indexOf("#/");

    	let location = hashPosition > -1
    	? window.location.href.substr(hashPosition + 1)
    	: "/";

    	// Check if there's a querystring
    	const qsPosition = location.indexOf("?");

    	let querystring = "";

    	if (qsPosition > -1) {
    		querystring = location.substr(qsPosition + 1);
    		location = location.substr(0, qsPosition);
    	}

    	return { location, querystring };
    }

    const loc = readable(getLocation(), // eslint-disable-next-line prefer-arrow-callback
    function start(set) {
    	const update = () => {
    		set(getLocation());
    	};

    	window.addEventListener("hashchange", update, false);

    	return function stop() {
    		window.removeEventListener("hashchange", update, false);
    	};
    });

    const location = derived(loc, $loc => $loc.location);
    const querystring = derived(loc, $loc => $loc.querystring);

    function push(location) {
    	if (!location || location.length < 1 || location.charAt(0) != "/" && location.indexOf("#/") !== 0) {
    		throw Error("Invalid parameter location");
    	}

    	// Execute this code when the current call stack is complete
    	return nextTickPromise(() => {
    		window.location.hash = (location.charAt(0) == "#" ? "" : "#") + location;
    	});
    }

    function pop() {
    	// Execute this code when the current call stack is complete
    	return nextTickPromise(() => {
    		window.history.back();
    	});
    }

    function replace(location) {
    	if (!location || location.length < 1 || location.charAt(0) != "/" && location.indexOf("#/") !== 0) {
    		throw Error("Invalid parameter location");
    	}

    	// Execute this code when the current call stack is complete
    	return nextTickPromise(() => {
    		const dest = (location.charAt(0) == "#" ? "" : "#") + location;

    		try {
    			window.history.replaceState(undefined, undefined, dest);
    		} catch(e) {
    			// eslint-disable-next-line no-console
    			console.warn("Caught exception while replacing the current page. If you're running this in the Svelte REPL, please note that the `replace` method might not work in this environment.");
    		}

    		// The method above doesn't trigger the hashchange event, so let's do that manually
    		window.dispatchEvent(new Event("hashchange"));
    	});
    }

    function link(node) {
    	// Only apply to <a> tags
    	if (!node || !node.tagName || node.tagName.toLowerCase() != "a") {
    		throw Error("Action \"link\" can only be used with <a> tags");
    	}

    	// Destination must start with '/'
    	const href = node.getAttribute("href");

    	if (!href || href.length < 1 || href.charAt(0) != "/") {
    		throw Error("Invalid value for \"href\" attribute");
    	}

    	// Add # to every href attribute
    	node.setAttribute("href", "#" + href);
    }

    function nextTickPromise(cb) {
    	return new Promise(resolve => {
    			setTimeout(
    				() => {
    					resolve(cb());
    				},
    				0
    			);
    		});
    }

    function instance$1($$self, $$props, $$invalidate) {
    	let $loc,
    		$$unsubscribe_loc = noop;

    	validate_store(loc, "loc");
    	component_subscribe($$self, loc, $$value => $$invalidate(4, $loc = $$value));
    	$$self.$$.on_destroy.push(() => $$unsubscribe_loc());
    	let { routes = {} } = $$props;
    	let { prefix = "" } = $$props;

    	/**
     * Container for a route: path, component
     */
    	class RouteItem {
    		/**
     * Initializes the object and creates a regular expression from the path, using regexparam.
     *
     * @param {string} path - Path to the route (must start with '/' or '*')
     * @param {SvelteComponent} component - Svelte component for the route
     */
    		constructor(path, component) {
    			if (!component || typeof component != "function" && (typeof component != "object" || component._sveltesparouter !== true)) {
    				throw Error("Invalid component object");
    			}

    			// Path must be a regular or expression, or a string starting with '/' or '*'
    			if (!path || typeof path == "string" && (path.length < 1 || path.charAt(0) != "/" && path.charAt(0) != "*") || typeof path == "object" && !(path instanceof RegExp)) {
    				throw Error("Invalid value for \"path\" argument");
    			}

    			const { pattern, keys } = regexparam(path);
    			this.path = path;

    			// Check if the component is wrapped and we have conditions
    			if (typeof component == "object" && component._sveltesparouter === true) {
    				this.component = component.route;
    				this.conditions = component.conditions || [];
    				this.userData = component.userData;
    			} else {
    				this.component = component;
    				this.conditions = [];
    				this.userData = undefined;
    			}

    			this._pattern = pattern;
    			this._keys = keys;
    		}

    		/**
     * Checks if `path` matches the current route.
     * If there's a match, will return the list of parameters from the URL (if any).
     * In case of no match, the method will return `null`.
     *
     * @param {string} path - Path to test
     * @returns {null|Object.<string, string>} List of paramters from the URL if there's a match, or `null` otherwise.
     */
    		match(path) {
    			// If there's a prefix, remove it before we run the matching
    			if (prefix && path.startsWith(prefix)) {
    				path = path.substr(prefix.length) || "/";
    			}

    			// Check if the pattern matches
    			const matches = this._pattern.exec(path);

    			if (matches === null) {
    				return null;
    			}

    			// If the input was a regular expression, this._keys would be false, so return matches as is
    			if (this._keys === false) {
    				return matches;
    			}

    			const out = {};
    			let i = 0;

    			while (i < this._keys.length) {
    				out[this._keys[i]] = matches[++i] || null;
    			}

    			return out;
    		}

    		/**
     * Dictionary with route details passed to the pre-conditions functions, as well as the `routeLoaded` and `conditionsFailed` events
     * @typedef {Object} RouteDetail
     * @property {SvelteComponent} component - Svelte component
     * @property {string} name - Name of the Svelte component
     * @property {string} location - Location path
     * @property {string} querystring - Querystring from the hash
     * @property {Object} [userData] - Custom data passed by the user
     */
    		/**
     * Executes all conditions (if any) to control whether the route can be shown. Conditions are executed in the order they are defined, and if a condition fails, the following ones aren't executed.
     * 
     * @param {RouteDetail} detail - Route detail
     * @returns {bool} Returns true if all the conditions succeeded
     */
    		checkConditions(detail) {
    			for (let i = 0; i < this.conditions.length; i++) {
    				if (!this.conditions[i](detail)) {
    					return false;
    				}
    			}

    			return true;
    		}
    	}

    	// Set up all routes
    	const routesList = [];

    	if (routes instanceof Map) {
    		// If it's a map, iterate on it right away
    		routes.forEach((route, path) => {
    			routesList.push(new RouteItem(path, route));
    		});
    	} else {
    		// We have an object, so iterate on its own properties
    		Object.keys(routes).forEach(path => {
    			routesList.push(new RouteItem(path, routes[path]));
    		});
    	}

    	// Props for the component to render
    	let component = null;

    	let componentParams = null;

    	// Event dispatcher from Svelte
    	const dispatch = createEventDispatcher();

    	// Just like dispatch, but executes on the next iteration of the event loop
    	const dispatchNextTick = (name, detail) => {
    		// Execute this code when the current call stack is complete
    		setTimeout(
    			() => {
    				dispatch(name, detail);
    			},
    			0
    		);
    	};

    	const writable_props = ["routes", "prefix"];

    	Object_1.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	function routeEvent_handler(event) {
    		bubble($$self, event);
    	}

    	function routeEvent_handler_1(event) {
    		bubble($$self, event);
    	}

    	$$self.$set = $$props => {
    		if ("routes" in $$props) $$invalidate(2, routes = $$props.routes);
    		if ("prefix" in $$props) $$invalidate(3, prefix = $$props.prefix);
    	};

    	$$self.$capture_state = () => ({
    		readable,
    		derived,
    		wrap,
    		getLocation,
    		loc,
    		location,
    		querystring,
    		push,
    		pop,
    		replace,
    		link,
    		nextTickPromise,
    		createEventDispatcher,
    		regexparam,
    		routes,
    		prefix,
    		RouteItem,
    		routesList,
    		component,
    		componentParams,
    		dispatch,
    		dispatchNextTick,
    		$loc
    	});

    	$$self.$inject_state = $$props => {
    		if ("routes" in $$props) $$invalidate(2, routes = $$props.routes);
    		if ("prefix" in $$props) $$invalidate(3, prefix = $$props.prefix);
    		if ("component" in $$props) $$invalidate(0, component = $$props.component);
    		if ("componentParams" in $$props) $$invalidate(1, componentParams = $$props.componentParams);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*component, $loc*/ 17) {
    			// Handle hash change events
    			// Listen to changes in the $loc store and update the page
    			 {
    				// Find a route matching the location
    				$$invalidate(0, component = null);

    				let i = 0;

    				while (!component && i < routesList.length) {
    					const match = routesList[i].match($loc.location);

    					if (match) {
    						const detail = {
    							component: routesList[i].component,
    							name: routesList[i].component.name,
    							location: $loc.location,
    							querystring: $loc.querystring,
    							userData: routesList[i].userData
    						};

    						// Check if the route can be loaded - if all conditions succeed
    						if (!routesList[i].checkConditions(detail)) {
    							// Trigger an event to notify the user
    							dispatchNextTick("conditionsFailed", detail);

    							break;
    						}

    						$$invalidate(0, component = routesList[i].component);

    						// Set componentParams onloy if we have a match, to avoid a warning similar to `<Component> was created with unknown prop 'params'`
    						// Of course, this assumes that developers always add a "params" prop when they are expecting parameters
    						if (match && typeof match == "object" && Object.keys(match).length) {
    							$$invalidate(1, componentParams = match);
    						} else {
    							$$invalidate(1, componentParams = null);
    						}

    						dispatchNextTick("routeLoaded", detail);
    					}

    					i++;
    				}
    			}
    		}
    	};

    	return [
    		component,
    		componentParams,
    		routes,
    		prefix,
    		$loc,
    		RouteItem,
    		routesList,
    		dispatch,
    		dispatchNextTick,
    		routeEvent_handler,
    		routeEvent_handler_1
    	];
    }

    class Router extends SvelteElement {
    	constructor(options) {
    		super();
    		init(this, { target: this.shadowRoot }, instance$1, create_fragment$1, safe_not_equal, { routes: 2, prefix: 3 });

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
    		return ["routes", "prefix"];
    	}

    	get routes() {
    		return this.$$.ctx[2];
    	}

    	set routes(routes) {
    		this.$set({ routes });
    		flush();
    	}

    	get prefix() {
    		return this.$$.ctx[3];
    	}

    	set prefix(prefix) {
    		this.$set({ prefix });
    		flush();
    	}
    }

    /* src/components/interact/total_recall/recall.svelte generated by Svelte v3.22.2 */

    const { console: console_1$1 } = globals;
    const file$1 = "src/components/interact/total_recall/recall.svelte";

    function get_each_context(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[19] = list[i];
    	child_ctx[21] = i;
    	return child_ctx;
    }

    function get_each_context_1(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[19] = list[i];
    	child_ctx[21] = i;
    	return child_ctx;
    }

    function get_each_context_2(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[19] = list[i];
    	child_ctx[21] = i;
    	return child_ctx;
    }

    // (271:0) {#if state === 'playing'}
    function create_if_block_3(ctx) {
    	let t0;
    	let div;
    	let button0;
    	let t2;
    	let button1;
    	let t4;
    	let button2;
    	let t6;
    	let p;
    	let t7;
    	let dispose;
    	let each_value_2 = /*playData*/ ctx[1];
    	validate_each_argument(each_value_2);
    	let each_blocks = [];

    	for (let i = 0; i < each_value_2.length; i += 1) {
    		each_blocks[i] = create_each_block_2(get_each_context_2(ctx, each_value_2, i));
    	}

    	let if_block = /*hasChecked*/ ctx[3] && create_if_block_4(ctx);

    	const block = {
    		c: function create() {
    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			t0 = space();
    			div = element("div");
    			button0 = element("button");
    			button0.textContent = "check";
    			t2 = space();
    			button1 = element("button");
    			button1.textContent = "I give up, show me";
    			t4 = space();
    			button2 = element("button");
    			button2.textContent = "restart";
    			t6 = space();
    			p = element("p");
    			t7 = text("How many do you remember?\n    ");
    			if (if_block) if_block.c();
    			add_location(button0, file$1, 282, 4, 111676);
    			add_location(button1, file$1, 283, 4, 111720);
    			add_location(button2, file$1, 284, 4, 111778);
    			add_location(div, file$1, 281, 2, 111666);
    			add_location(p, file$1, 286, 2, 111833);
    		},
    		m: function mount(target, anchor, remount) {
    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(target, anchor);
    			}

    			insert_dev(target, t0, anchor);
    			insert_dev(target, div, anchor);
    			append_dev(div, button0);
    			append_dev(div, t2);
    			append_dev(div, button1);
    			append_dev(div, t4);
    			append_dev(div, button2);
    			insert_dev(target, t6, anchor);
    			insert_dev(target, p, anchor);
    			append_dev(p, t7);
    			if (if_block) if_block.m(p, null);
    			if (remount) run_all(dispose);

    			dispose = [
    				listen_dev(button0, "click", /*check*/ ctx[8], false, false, false),
    				listen_dev(button1, "click", /*showMe*/ ctx[11], false, false, false),
    				listen_dev(button2, "click", /*restart*/ ctx[10], false, false, false)
    			];
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*feedback, guesses, playData*/ 70) {
    				each_value_2 = /*playData*/ ctx[1];
    				validate_each_argument(each_value_2);
    				let i;

    				for (i = 0; i < each_value_2.length; i += 1) {
    					const child_ctx = get_each_context_2(ctx, each_value_2, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block_2(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(t0.parentNode, t0);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value_2.length;
    			}

    			if (/*hasChecked*/ ctx[3]) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block_4(ctx);
    					if_block.c();
    					if_block.m(p, null);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}
    		},
    		d: function destroy(detaching) {
    			destroy_each(each_blocks, detaching);
    			if (detaching) detach_dev(t0);
    			if (detaching) detach_dev(div);
    			if (detaching) detach_dev(t6);
    			if (detaching) detach_dev(p);
    			if (if_block) if_block.d();
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_3.name,
    		type: "if",
    		source: "(271:0) {#if state === 'playing'}",
    		ctx
    	});

    	return block;
    }

    // (272:2) {#each playData as item, index}
    function create_each_block_2(ctx) {
    	let div;
    	let input;
    	let input_class_value;
    	let input_disabled_value;
    	let dispose;

    	function input_input_handler() {
    		/*input_input_handler*/ ctx[17].call(input, /*index*/ ctx[21]);
    	}

    	const block = {
    		c: function create() {
    			div = element("div");
    			input = element("input");
    			attr_dev(input, "class", input_class_value = /*feedback*/ ctx[6][/*index*/ ctx[21]]);
    			input.disabled = input_disabled_value = /*feedback*/ ctx[6][/*index*/ ctx[21]] === "found";
    			attr_dev(input, "type", "text");
    			attr_dev(input, "placeholder", "");
    			add_location(input, file$1, 273, 6, 111475);
    			add_location(div, file$1, 272, 4, 111463);
    		},
    		m: function mount(target, anchor, remount) {
    			insert_dev(target, div, anchor);
    			append_dev(div, input);
    			set_input_value(input, /*guesses*/ ctx[2][/*index*/ ctx[21]]);
    			if (remount) dispose();
    			dispose = listen_dev(input, "input", input_input_handler);
    		},
    		p: function update(new_ctx, dirty) {
    			ctx = new_ctx;

    			if (dirty & /*feedback*/ 64 && input_class_value !== (input_class_value = /*feedback*/ ctx[6][/*index*/ ctx[21]])) {
    				attr_dev(input, "class", input_class_value);
    			}

    			if (dirty & /*feedback*/ 64 && input_disabled_value !== (input_disabled_value = /*feedback*/ ctx[6][/*index*/ ctx[21]] === "found")) {
    				prop_dev(input, "disabled", input_disabled_value);
    			}

    			if (dirty & /*guesses*/ 4 && input.value !== /*guesses*/ ctx[2][/*index*/ ctx[21]]) {
    				set_input_value(input, /*guesses*/ ctx[2][/*index*/ ctx[21]]);
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block_2.name,
    		type: "each",
    		source: "(272:2) {#each playData as item, index}",
    		ctx
    	});

    	return block;
    }

    // (289:4) {#if hasChecked}
    function create_if_block_4(ctx) {
    	let t0;
    	let t1;

    	const block = {
    		c: function create() {
    			t0 = text(/*leftToFind*/ ctx[4]);
    			t1 = text(" left");
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, t0, anchor);
    			insert_dev(target, t1, anchor);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*leftToFind*/ 16) set_data_dev(t0, /*leftToFind*/ ctx[4]);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(t0);
    			if (detaching) detach_dev(t1);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_4.name,
    		type: "if",
    		source: "(289:4) {#if hasChecked}",
    		ctx
    	});

    	return block;
    }

    // (293:0) {#if state === 'finished'}
    function create_if_block_1$1(ctx) {
    	let t0;
    	let p0;
    	let t2;
    	let t3;
    	let p1;
    	let t4;
    	let t5;
    	let t6;
    	let t7;
    	let div;
    	let button0;
    	let t9;
    	let button1;
    	let dispose;
    	let each_value_1 = /*playData*/ ctx[1];
    	validate_each_argument(each_value_1);
    	let each_blocks = [];

    	for (let i = 0; i < each_value_1.length; i += 1) {
    		each_blocks[i] = create_each_block_1(get_each_context_1(ctx, each_value_1, i));
    	}

    	let if_block = /*perfect*/ ctx[5] && create_if_block_2(ctx);

    	const block = {
    		c: function create() {
    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			t0 = space();
    			p0 = element("p");
    			p0.textContent = "Well done!";
    			t2 = space();
    			if (if_block) if_block.c();
    			t3 = space();
    			p1 = element("p");
    			t4 = text("You took ");
    			t5 = text(/*attempts*/ ctx[7]);
    			t6 = text(" attempt(s)");
    			t7 = space();
    			div = element("div");
    			button0 = element("button");
    			button0.textContent = "play again";
    			t9 = space();
    			button1 = element("button");
    			button1.textContent = "restart";
    			add_location(p0, file$1, 303, 2, 112192);
    			add_location(p1, file$1, 307, 2, 112263);
    			add_location(button0, file$1, 310, 4, 112314);
    			add_location(button1, file$1, 311, 4, 112367);
    			add_location(div, file$1, 309, 2, 112304);
    		},
    		m: function mount(target, anchor, remount) {
    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(target, anchor);
    			}

    			insert_dev(target, t0, anchor);
    			insert_dev(target, p0, anchor);
    			insert_dev(target, t2, anchor);
    			if (if_block) if_block.m(target, anchor);
    			insert_dev(target, t3, anchor);
    			insert_dev(target, p1, anchor);
    			append_dev(p1, t4);
    			append_dev(p1, t5);
    			append_dev(p1, t6);
    			insert_dev(target, t7, anchor);
    			insert_dev(target, div, anchor);
    			append_dev(div, button0);
    			append_dev(div, t9);
    			append_dev(div, button1);
    			if (remount) run_all(dispose);

    			dispose = [
    				listen_dev(button0, "click", /*playAgain*/ ctx[9], false, false, false),
    				listen_dev(button1, "click", /*restart*/ ctx[10], false, false, false)
    			];
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*feedback, guesses, playData*/ 70) {
    				each_value_1 = /*playData*/ ctx[1];
    				validate_each_argument(each_value_1);
    				let i;

    				for (i = 0; i < each_value_1.length; i += 1) {
    					const child_ctx = get_each_context_1(ctx, each_value_1, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block_1(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(t0.parentNode, t0);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value_1.length;
    			}

    			if (/*perfect*/ ctx[5]) {
    				if (if_block) ; else {
    					if_block = create_if_block_2(ctx);
    					if_block.c();
    					if_block.m(t3.parentNode, t3);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}

    			if (dirty & /*attempts*/ 128) set_data_dev(t5, /*attempts*/ ctx[7]);
    		},
    		d: function destroy(detaching) {
    			destroy_each(each_blocks, detaching);
    			if (detaching) detach_dev(t0);
    			if (detaching) detach_dev(p0);
    			if (detaching) detach_dev(t2);
    			if (if_block) if_block.d(detaching);
    			if (detaching) detach_dev(t3);
    			if (detaching) detach_dev(p1);
    			if (detaching) detach_dev(t7);
    			if (detaching) detach_dev(div);
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1$1.name,
    		type: "if",
    		source: "(293:0) {#if state === 'finished'}",
    		ctx
    	});

    	return block;
    }

    // (294:2) {#each playData as item, index}
    function create_each_block_1(ctx) {
    	let div;
    	let input;
    	let input_class_value;
    	let input_disabled_value;
    	let dispose;

    	function input_input_handler_1() {
    		/*input_input_handler_1*/ ctx[18].call(input, /*index*/ ctx[21]);
    	}

    	const block = {
    		c: function create() {
    			div = element("div");
    			input = element("input");
    			attr_dev(input, "class", input_class_value = /*feedback*/ ctx[6][/*index*/ ctx[21]]);
    			input.disabled = input_disabled_value = /*feedback*/ ctx[6][/*index*/ ctx[21]] === "found";
    			attr_dev(input, "type", "text");
    			attr_dev(input, "placeholder", "");
    			add_location(input, file$1, 295, 6, 112001);
    			add_location(div, file$1, 294, 4, 111989);
    		},
    		m: function mount(target, anchor, remount) {
    			insert_dev(target, div, anchor);
    			append_dev(div, input);
    			set_input_value(input, /*guesses*/ ctx[2][/*index*/ ctx[21]]);
    			if (remount) dispose();
    			dispose = listen_dev(input, "input", input_input_handler_1);
    		},
    		p: function update(new_ctx, dirty) {
    			ctx = new_ctx;

    			if (dirty & /*feedback*/ 64 && input_class_value !== (input_class_value = /*feedback*/ ctx[6][/*index*/ ctx[21]])) {
    				attr_dev(input, "class", input_class_value);
    			}

    			if (dirty & /*feedback*/ 64 && input_disabled_value !== (input_disabled_value = /*feedback*/ ctx[6][/*index*/ ctx[21]] === "found")) {
    				prop_dev(input, "disabled", input_disabled_value);
    			}

    			if (dirty & /*guesses*/ 4 && input.value !== /*guesses*/ ctx[2][/*index*/ ctx[21]]) {
    				set_input_value(input, /*guesses*/ ctx[2][/*index*/ ctx[21]]);
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block_1.name,
    		type: "each",
    		source: "(294:2) {#each playData as item, index}",
    		ctx
    	});

    	return block;
    }

    // (305:2) {#if perfect}
    function create_if_block_2(ctx) {
    	let p;

    	const block = {
    		c: function create() {
    			p = element("p");
    			p.textContent = "Perfect recall!";
    			add_location(p, file$1, 305, 4, 112230);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, p, anchor);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(p);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_2.name,
    		type: "if",
    		source: "(305:2) {#if perfect}",
    		ctx
    	});

    	return block;
    }

    // (316:0) {#if state === 'show-me'}
    function create_if_block$2(ctx) {
    	let t0;
    	let div;
    	let button0;
    	let t2;
    	let button1;
    	let dispose;
    	let each_value = /*playData*/ ctx[1];
    	validate_each_argument(each_value);
    	let each_blocks = [];

    	for (let i = 0; i < each_value.length; i += 1) {
    		each_blocks[i] = create_each_block(get_each_context(ctx, each_value, i));
    	}

    	const block = {
    		c: function create() {
    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			t0 = space();
    			div = element("div");
    			button0 = element("button");
    			button0.textContent = "play again";
    			t2 = space();
    			button1 = element("button");
    			button1.textContent = "restart";
    			add_location(button0, file$1, 327, 4, 112656);
    			add_location(button1, file$1, 328, 4, 112709);
    			add_location(div, file$1, 326, 2, 112646);
    		},
    		m: function mount(target, anchor, remount) {
    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(target, anchor);
    			}

    			insert_dev(target, t0, anchor);
    			insert_dev(target, div, anchor);
    			append_dev(div, button0);
    			append_dev(div, t2);
    			append_dev(div, button1);
    			if (remount) run_all(dispose);

    			dispose = [
    				listen_dev(button0, "click", /*playAgain*/ ctx[9], false, false, false),
    				listen_dev(button1, "click", /*restart*/ ctx[10], false, false, false)
    			];
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*playData*/ 2) {
    				each_value = /*playData*/ ctx[1];
    				validate_each_argument(each_value);
    				let i;

    				for (i = 0; i < each_value.length; i += 1) {
    					const child_ctx = get_each_context(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(t0.parentNode, t0);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value.length;
    			}
    		},
    		d: function destroy(detaching) {
    			destroy_each(each_blocks, detaching);
    			if (detaching) detach_dev(t0);
    			if (detaching) detach_dev(div);
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$2.name,
    		type: "if",
    		source: "(316:0) {#if state === 'show-me'}",
    		ctx
    	});

    	return block;
    }

    // (317:2) {#each playData as item, index}
    function create_each_block(ctx) {
    	let div;
    	let input;
    	let input_value_value;

    	const block = {
    		c: function create() {
    			div = element("div");
    			input = element("input");
    			attr_dev(input, "class", "found");
    			input.disabled = "true";
    			attr_dev(input, "type", "text");
    			attr_dev(input, "placeholder", "");
    			input.value = input_value_value = /*item*/ ctx[19];
    			add_location(input, file$1, 318, 6, 112503);
    			add_location(div, file$1, 317, 4, 112491);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div, anchor);
    			append_dev(div, input);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*playData*/ 2 && input_value_value !== (input_value_value = /*item*/ ctx[19]) && input.value !== input_value_value) {
    				prop_dev(input, "value", input_value_value);
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block.name,
    		type: "each",
    		source: "(317:2) {#each playData as item, index}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$2(ctx) {
    	let t0;
    	let t1;
    	let if_block2_anchor;
    	let if_block0 = /*state*/ ctx[0] === "playing" && create_if_block_3(ctx);
    	let if_block1 = /*state*/ ctx[0] === "finished" && create_if_block_1$1(ctx);
    	let if_block2 = /*state*/ ctx[0] === "show-me" && create_if_block$2(ctx);

    	const block = {
    		c: function create() {
    			if (if_block0) if_block0.c();
    			t0 = space();
    			if (if_block1) if_block1.c();
    			t1 = space();
    			if (if_block2) if_block2.c();
    			if_block2_anchor = empty();
    			this.c = noop;
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			if (if_block0) if_block0.m(target, anchor);
    			insert_dev(target, t0, anchor);
    			if (if_block1) if_block1.m(target, anchor);
    			insert_dev(target, t1, anchor);
    			if (if_block2) if_block2.m(target, anchor);
    			insert_dev(target, if_block2_anchor, anchor);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*state*/ ctx[0] === "playing") {
    				if (if_block0) {
    					if_block0.p(ctx, dirty);
    				} else {
    					if_block0 = create_if_block_3(ctx);
    					if_block0.c();
    					if_block0.m(t0.parentNode, t0);
    				}
    			} else if (if_block0) {
    				if_block0.d(1);
    				if_block0 = null;
    			}

    			if (/*state*/ ctx[0] === "finished") {
    				if (if_block1) {
    					if_block1.p(ctx, dirty);
    				} else {
    					if_block1 = create_if_block_1$1(ctx);
    					if_block1.c();
    					if_block1.m(t1.parentNode, t1);
    				}
    			} else if (if_block1) {
    				if_block1.d(1);
    				if_block1 = null;
    			}

    			if (/*state*/ ctx[0] === "show-me") {
    				if (if_block2) {
    					if_block2.p(ctx, dirty);
    				} else {
    					if_block2 = create_if_block$2(ctx);
    					if_block2.c();
    					if_block2.m(if_block2_anchor.parentNode, if_block2_anchor);
    				}
    			} else if (if_block2) {
    				if_block2.d(1);
    				if_block2 = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (if_block0) if_block0.d(detaching);
    			if (detaching) detach_dev(t0);
    			if (if_block1) if_block1.d(detaching);
    			if (detaching) detach_dev(t1);
    			if (if_block2) if_block2.d(detaching);
    			if (detaching) detach_dev(if_block2_anchor);
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

    function clean(input) {
    	// TODO have the UI allow for more options
    	return input.toLowerCase();
    }

    function instance$2($$self, $$props, $$invalidate) {
    	const dispatch = createEventDispatcher();
    	let { data = [] } = $$props;
    	let state = "playing";
    	let playing = false;

    	// clean the inputs
    	let found = [];

    	let playData = [];
    	let guesses = [];
    	let hasChecked = false;

    	playData = data.map(item => {
    		return clean(item);
    	});

    	let leftToFind = playData.length;
    	playing = true;
    	let perfect = false;
    	let feedback = Array(playData.length).fill("");
    	let results = [];
    	let attempts = 0;

    	function check() {
    		$$invalidate(7, attempts = attempts + 1);
    		$$invalidate(3, hasChecked = true);
    		console.log(guesses);

    		results = guesses.map(input => {
    			return clean(input);
    		});

    		// Get the unique results
    		let uniques = Array.from(new Set(results));

    		uniques = uniques.filter(item => playData.includes(item));

    		let lookUp = uniques.map(item => {
    			return { data: item, position: -1 };
    		});

    		results.forEach((input, position) => {
    			lookUp = lookUp.map(item => {
    				// skip if already found
    				if (item.position !== -1) {
    					return item;
    				}

    				if (item.data !== input) {
    					return item;
    				}

    				item.position = position;
    				return item;
    			});
    		});

    		// Set all to not found
    		$$invalidate(6, feedback = Array(playData.length).fill("notfound"));

    		lookUp = lookUp.map(item => {
    			if (item.position === -1) {
    				return item;
    			}

    			$$invalidate(6, feedback[item.position] = "found", feedback);
    			return item;
    		});

    		$$invalidate(4, leftToFind = playData.length - uniques.length);

    		if (leftToFind === 0) {
    			$$invalidate(0, state = "finished");

    			if (attempts === 1) {
    				$$invalidate(5, perfect = JSON.stringify(results) === JSON.stringify(playData));
    			}

    			console.log("actual", JSON.stringify(playData));
    			console.log("guesses", JSON.stringify(results));
    		}
    	}

    	function playAgain() {
    		dispatch("finished", { perfect, attempts, playAgain: true });
    	}

    	function restart() {
    		dispatch("finished", { perfect, attempts, playAgain: false });
    	}

    	function showMe() {
    		$$invalidate(0, state = "show-me");
    	}

    	const writable_props = ["data"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$1.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	function input_input_handler(index) {
    		guesses[index] = this.value;
    		$$invalidate(2, guesses);
    	}

    	function input_input_handler_1(index) {
    		guesses[index] = this.value;
    		$$invalidate(2, guesses);
    	}

    	$$self.$set = $$props => {
    		if ("data" in $$props) $$invalidate(12, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		createEventDispatcher,
    		dispatch,
    		data,
    		state,
    		playing,
    		found,
    		playData,
    		guesses,
    		hasChecked,
    		leftToFind,
    		perfect,
    		feedback,
    		results,
    		attempts,
    		check,
    		playAgain,
    		restart,
    		showMe,
    		clean
    	});

    	$$self.$inject_state = $$props => {
    		if ("data" in $$props) $$invalidate(12, data = $$props.data);
    		if ("state" in $$props) $$invalidate(0, state = $$props.state);
    		if ("playing" in $$props) playing = $$props.playing;
    		if ("found" in $$props) found = $$props.found;
    		if ("playData" in $$props) $$invalidate(1, playData = $$props.playData);
    		if ("guesses" in $$props) $$invalidate(2, guesses = $$props.guesses);
    		if ("hasChecked" in $$props) $$invalidate(3, hasChecked = $$props.hasChecked);
    		if ("leftToFind" in $$props) $$invalidate(4, leftToFind = $$props.leftToFind);
    		if ("perfect" in $$props) $$invalidate(5, perfect = $$props.perfect);
    		if ("feedback" in $$props) $$invalidate(6, feedback = $$props.feedback);
    		if ("results" in $$props) results = $$props.results;
    		if ("attempts" in $$props) $$invalidate(7, attempts = $$props.attempts);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [
    		state,
    		playData,
    		guesses,
    		hasChecked,
    		leftToFind,
    		perfect,
    		feedback,
    		attempts,
    		check,
    		playAgain,
    		restart,
    		showMe,
    		data,
    		playing,
    		results,
    		dispatch,
    		found,
    		input_input_handler,
    		input_input_handler_1
    	];
    }

    class Recall extends SvelteElement {
    	constructor(options) {
    		super();
    		this.shadowRoot.innerHTML = `<style>button,input{font-family:inherit;font-size:100%;line-height:1.15;margin:0}button,input{overflow:visible}button{text-transform:none}button{-webkit-appearance:button}button::-moz-focus-inner{border-style:none;padding:0}button:-moz-focusring{outline:1px dotted ButtonText}div,p{box-sizing:border-box}.ba{border-style:solid;border-width:1px}.bt{border-top-style:solid;border-top-width:1px}.br{border-right-style:solid;border-right-width:1px}.bb{border-bottom-style:solid;border-bottom-width:1px}.bl{border-left-style:solid;border-left-width:1px}.bn{border-style:none;border-width:0}.b--black{border-color:#000}.b--moon-gray{border-color:#ccc}.b--black-30{border-color:rgba(0,0,0,.3)}.b--black-20{border-color:rgba(0,0,0,.2)}.b--black-10{border-color:rgba(0,0,0,.1)}.b--black-05{border-color:rgba(0,0,0,.05)}.b--red{border-color:#ff4136}.b--yellow{border-color:gold}.b--washed-yellow{border-color:#fffceb}.b--transparent{border-color:transparent}.br1{border-radius:.125rem}.br2{border-radius:.25rem}.br3{border-radius:.5rem}.b--dotted{border-style:dotted}.bw1{border-width:.125rem}.bw2{border-width:.25rem}.bw3{border-width:.5rem}.bt-0{border-top-width:0}.br-0{border-right-width:0}.bl-0{border-left-width:0}.pre{overflow-x:auto;overflow-y:hidden;overflow:scroll}.di{display:inline}.db{display:block}.dib{display:inline-block}.dt{display:table}.dtc{display:table-cell}.flex{display:flex}.flex-column{flex-direction:column}.items-end{align-items:flex-end}.items-center{align-items:center}.justify-center{justify-content:center}.fl{float:left}.fl,.fr{_display:inline}.fr{float:right}.athelas{font-family:athelas,georgia,serif}.fs-normal{font-style:normal}.b{font-weight:700}.fw3{font-weight:300}.fw4{font-weight:400}.fw5{font-weight:500}.fw6{font-weight:600}.input-reset{-webkit-appearance:none;-moz-appearance:none}.input-reset::-moz-focus-inner{border:0;padding:0}.h1{height:1rem}.h2{height:2rem}.h3{height:4rem}.tracked{letter-spacing:.1em}.lh-title{line-height:1.25}.lh-copy{line-height:1.5}.link{text-decoration:none}.link,.link:active,.link:focus,.link:hover,.link:link,.link:visited{transition:color .15s ease-in}.link:focus{outline:1px dotted currentColor}.list{list-style-type:none}.mw-100{max-width:100%}.w1{width:1rem}.w-25{width:25%}.w-75{width:75%}.w-100{width:100%}.black-90{color:rgba(0,0,0,.9)}.black-80{color:rgba(0,0,0,.8)}.black-70{color:rgba(0,0,0,.7)}.black-60{color:rgba(0,0,0,.6)}.black-40{color:rgba(0,0,0,.4)}.black{color:#000}.dark-gray{color:#333}.white{color:#fff}.dark-pink{color:#d5008f}.navy{color:#001b44}.bg-white{background-color:#fff}.bg-transparent{background-color:transparent}.bg-light-red{background-color:#ff725c}.bg-washed-yellow{background-color:#fffceb}.bg-washed-red{background-color:#ffdfdf}.hover-red:focus,.hover-red:hover{color:#ff4136}.hover-blue:focus,.hover-blue:hover{color:#357edd}.pa0{padding:0}.pa1{padding:.25rem}.pa2{padding:.5rem}.pa3{padding:1rem}.pa4{padding:2rem}.pl0{padding-left:0}.pl4{padding-left:2rem}.pb2{padding-bottom:.5rem}.pt2{padding-top:.5rem}.pt5{padding-top:4rem}.pv0{padding-top:0;padding-bottom:0}.pv1{padding-top:.25rem;padding-bottom:.25rem}.pv2{padding-top:.5rem;padding-bottom:.5rem}.pv3{padding-top:1rem;padding-bottom:1rem}.pv5{padding-top:4rem;padding-bottom:4rem}.ph0{padding-left:0;padding-right:0}.ph1{padding-left:.25rem;padding-right:.25rem}.ph3{padding-left:1rem;padding-right:1rem}.ph4{padding-left:2rem;padding-right:2rem}.ml0{margin-left:0}.ml3{margin-left:1rem}.mr1{margin-right:.25rem}.mr2{margin-right:.5rem}.mb0{margin-bottom:0}.mb2{margin-bottom:.5rem}.mb3{margin-bottom:1rem}.mb5{margin-bottom:4rem}.mt0{margin-top:0}.mt2{margin-top:.5rem}.mt3{margin-top:1rem}.mt4{margin-top:2rem}.mv0{margin-top:0;margin-bottom:0}.mv2{margin-top:.5rem;margin-bottom:.5rem}.mv3{margin-top:1rem;margin-bottom:1rem}.mh0{margin-left:0;margin-right:0}.mh1{margin-left:.25rem;margin-right:.25rem}.underline{text-decoration:underline}.tc{text-align:center}.ttu{text-transform:uppercase}.f2{font-size:2.25rem}.f3{font-size:1.5rem}.f4{font-size:1.25rem}.f5{font-size:1rem}.f6{font-size:.875rem}.measure{max-width:30em}.center{margin-left:auto}.center{margin-right:auto}.pre{white-space:pre}.v-mid{vertical-align:middle}.dim{opacity:1}.dim,.dim:focus,.dim:hover{transition:opacity .15s ease-in}.dim:focus,.dim:hover{opacity:.5}.dim:active{opacity:.8;transition:opacity .15s ease-out}@media screen and (min-width:30em){.pa2-ns{padding:.5rem}.pa5-ns{padding:4rem}.ph1-ns{padding-left:.25rem;padding-right:.25rem}.ph5-ns{padding-left:4rem;padding-right:4rem}.mr6-ns{margin-right:8rem}.f2-ns{font-size:2.25rem}.f5-ns{font-size:1rem}}@media screen and (min-width:30em) and (max-width:60em){.mr3-m{margin-right:1rem}.f4-m{font-size:1.25rem}}@media screen and (min-width:60em){.ph4-l{padding-left:2rem;padding-right:2rem}.mr2-l{margin-right:.5rem}.mr4-l{margin-right:2rem}.mr5-l{margin-right:4rem}.f3-l{font-size:1.5rem}}.found{border:4px solid #19a974;border-radius:2px}</style>`;
    		init(this, { target: this.shadowRoot }, instance$2, create_fragment$2, safe_not_equal, { data: 12 });

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
    		return ["data"];
    	}

    	get data() {
    		return this.$$.ctx[12];
    	}

    	set data(data) {
    		this.$set({ data });
    		flush();
    	}
    }

    /* src/components/interact/total_recall/view.svelte generated by Svelte v3.22.2 */
    const file$2 = "src/components/interact/total_recall/view.svelte";

    function create_fragment$3(ctx) {
    	let blockquote;
    	let p;
    	let t;

    	const block = {
    		c: function create() {
    			blockquote = element("blockquote");
    			p = element("p");
    			t = text(/*show*/ ctx[0]);
    			this.c = noop;
    			attr_dev(p, "class", "dark-pink f5 f4-m f3-l lh-copy measure mt0");
    			add_location(p, file$2, 184, 2, 109264);
    			attr_dev(blockquote, "class", "athelas ml0 mt4 pl4 black-90 bl bw2 b--black");
    			add_location(blockquote, file$2, 183, 0, 109196);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, blockquote, anchor);
    			append_dev(blockquote, p);
    			append_dev(p, t);
    		},
    		p: function update(ctx, [dirty]) {
    			if (dirty & /*show*/ 1) set_data_dev(t, /*show*/ ctx[0]);
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(blockquote);
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
    	const dispatch = createEventDispatcher();
    	let { data = [] } = $$props;
    	let { speed = 1000 } = $$props;
    	let index = 0;
    	const size = data.length - 1;

    	const cancel = () => {
    		clearInterval(timeout);
    	};

    	const timeout = setInterval(
    		() => {
    			$$invalidate(0, show = data[index]);
    			$$invalidate(3, index = index + 1);

    			if (index <= size) {
    				return;
    			}

    			cancel();
    			dispatch("finished");
    		},
    		speed
    	);

    	const writable_props = ["data", "speed"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	$$self.$set = $$props => {
    		if ("data" in $$props) $$invalidate(1, data = $$props.data);
    		if ("speed" in $$props) $$invalidate(2, speed = $$props.speed);
    	};

    	$$self.$capture_state = () => ({
    		createEventDispatcher,
    		dispatch,
    		data,
    		speed,
    		index,
    		size,
    		cancel,
    		timeout,
    		show
    	});

    	$$self.$inject_state = $$props => {
    		if ("data" in $$props) $$invalidate(1, data = $$props.data);
    		if ("speed" in $$props) $$invalidate(2, speed = $$props.speed);
    		if ("index" in $$props) $$invalidate(3, index = $$props.index);
    		if ("show" in $$props) $$invalidate(0, show = $$props.show);
    	};

    	let show;

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, index*/ 10) {
    			 $$invalidate(0, show = data[index]);
    		}
    	};

    	return [show, data, speed];
    }

    class View extends SvelteElement {
    	constructor(options) {
    		super();
    		this.shadowRoot.innerHTML = `<style>blockquote,p{box-sizing:border-box}.bl{border-left-style:solid;border-left-width:1px}.b--black{border-color:#000}.bw2{border-width:.25rem}.athelas{font-family:athelas,georgia,serif}.lh-copy{line-height:1.5}.black-90{color:rgba(0,0,0,.9)}.dark-pink{color:#d5008f}.pl4{padding-left:2rem}.ml0{margin-left:0}.mt0{margin-top:0}.mt4{margin-top:2rem}.f5{font-size:1rem}.measure{max-width:30em}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){.f4-m{font-size:1.25rem}}@media screen and (min-width:60em){.f3-l{font-size:1.5rem}}</style>`;
    		init(this, { target: this.shadowRoot }, instance$3, create_fragment$3, safe_not_equal, { data: 1, speed: 2 });

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
    		return ["data", "speed"];
    	}

    	get data() {
    		return this.$$.ctx[1];
    	}

    	set data(data) {
    		this.$set({ data });
    		flush();
    	}

    	get speed() {
    		return this.$$.ctx[2];
    	}

    	set speed(speed) {
    		this.$set({ speed });
    		flush();
    	}
    }

    /* src/components/interact/total_recall/v1.svelte generated by Svelte v3.22.2 */
    const file$3 = "src/components/interact/total_recall/v1.svelte";

    // (240:4) {#if state === 'not-playing'}
    function create_if_block_2$1(ctx) {
    	let h1;
    	let t1;
    	let p0;
    	let t3;
    	let p1;
    	let t5;
    	let p2;
    	let t7;
    	let p3;
    	let span0;
    	let t9;
    	let input0;
    	let t10;
    	let p4;
    	let span1;
    	let t12;
    	let input1;
    	let input1_max_value;
    	let t13;
    	let button;
    	let dispose;

    	const block = {
    		c: function create() {
    			h1 = element("h1");
    			h1.textContent = "Rules";
    			t1 = space();
    			p0 = element("p");
    			p0.textContent = "Can you remember all the words?";
    			t3 = space();
    			p1 = element("p");
    			p1.textContent = "Can you remember the order to make it perfect?";
    			t5 = space();
    			p2 = element("p");
    			p2.textContent = "How many times do you need to check?";
    			t7 = space();
    			p3 = element("p");
    			span0 = element("span");
    			span0.textContent = "How many to recall?";
    			t9 = space();
    			input0 = element("input");
    			t10 = space();
    			p4 = element("p");
    			span1 = element("span");
    			span1.textContent = "How fast? (seconds)";
    			t12 = space();
    			input1 = element("input");
    			t13 = space();
    			button = element("button");
    			button.textContent = "Are you ready to play?";
    			add_location(h1, file$3, 240, 6, 110253);
    			add_location(p0, file$3, 241, 6, 110274);
    			add_location(p1, file$3, 242, 6, 110319);
    			add_location(p2, file$3, 243, 6, 110379);
    			add_location(span0, file$3, 246, 8, 110442);
    			attr_dev(input0, "type", "number");
    			attr_dev(input0, "max", /*maxSize*/ ctx[4]);
    			attr_dev(input0, "min", "1");
    			add_location(input0, file$3, 247, 8, 110483);
    			add_location(p3, file$3, 245, 6, 110430);
    			add_location(span1, file$3, 251, 8, 110581);
    			attr_dev(input1, "type", "number");
    			attr_dev(input1, "max", input1_max_value = 5);
    			attr_dev(input1, "min", "1");
    			add_location(input1, file$3, 252, 8, 110622);
    			add_location(p4, file$3, 250, 6, 110569);
    			attr_dev(button, "class", "br3");
    			add_location(button, file$3, 254, 6, 110698);
    		},
    		m: function mount(target, anchor, remount) {
    			insert_dev(target, h1, anchor);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, p0, anchor);
    			insert_dev(target, t3, anchor);
    			insert_dev(target, p1, anchor);
    			insert_dev(target, t5, anchor);
    			insert_dev(target, p2, anchor);
    			insert_dev(target, t7, anchor);
    			insert_dev(target, p3, anchor);
    			append_dev(p3, span0);
    			append_dev(p3, t9);
    			append_dev(p3, input0);
    			set_input_value(input0, /*gameSize*/ ctx[0]);
    			insert_dev(target, t10, anchor);
    			insert_dev(target, p4, anchor);
    			append_dev(p4, span1);
    			append_dev(p4, t12);
    			append_dev(p4, input1);
    			set_input_value(input1, /*speed*/ ctx[1]);
    			insert_dev(target, t13, anchor);
    			insert_dev(target, button, anchor);
    			if (remount) run_all(dispose);

    			dispose = [
    				listen_dev(input0, "input", /*input0_input_handler*/ ctx[13]),
    				listen_dev(input1, "input", /*input1_input_handler*/ ctx[14]),
    				listen_dev(button, "click", /*play*/ ctx[6], false, false, false)
    			];
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*maxSize*/ 16) {
    				attr_dev(input0, "max", /*maxSize*/ ctx[4]);
    			}

    			if (dirty & /*gameSize*/ 1 && to_number(input0.value) !== /*gameSize*/ ctx[0]) {
    				set_input_value(input0, /*gameSize*/ ctx[0]);
    			}

    			if (dirty & /*speed*/ 2 && to_number(input1.value) !== /*speed*/ ctx[1]) {
    				set_input_value(input1, /*speed*/ ctx[1]);
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(h1);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(p0);
    			if (detaching) detach_dev(t3);
    			if (detaching) detach_dev(p1);
    			if (detaching) detach_dev(t5);
    			if (detaching) detach_dev(p2);
    			if (detaching) detach_dev(t7);
    			if (detaching) detach_dev(p3);
    			if (detaching) detach_dev(t10);
    			if (detaching) detach_dev(p4);
    			if (detaching) detach_dev(t13);
    			if (detaching) detach_dev(button);
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_2$1.name,
    		type: "if",
    		source: "(240:4) {#if state === 'not-playing'}",
    		ctx
    	});

    	return block;
    }

    // (258:4) {#if state === 'playing'}
    function create_if_block_1$2(ctx) {
    	let current;

    	const view = new View({
    			props: {
    				data: /*playData*/ ctx[2],
    				speed: /*speed*/ ctx[1] * 1000
    			},
    			$$inline: true
    		});

    	view.$on("finished", /*handleFinished*/ ctx[8]);

    	const block = {
    		c: function create() {
    			create_component(view.$$.fragment);
    		},
    		m: function mount(target, anchor) {
    			mount_component(view, target, anchor);
    			current = true;
    		},
    		p: function update(ctx, dirty) {
    			const view_changes = {};
    			if (dirty & /*playData*/ 4) view_changes.data = /*playData*/ ctx[2];
    			if (dirty & /*speed*/ 2) view_changes.speed = /*speed*/ ctx[1] * 1000;
    			view.$set(view_changes);
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(view.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(view.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			destroy_component(view, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1$2.name,
    		type: "if",
    		source: "(258:4) {#if state === 'playing'}",
    		ctx
    	});

    	return block;
    }

    // (262:4) {#if state === 'recall'}
    function create_if_block$3(ctx) {
    	let current;

    	const recall = new Recall({
    			props: { data: /*playData*/ ctx[2] },
    			$$inline: true
    		});

    	recall.$on("finished", /*finished*/ ctx[7]);

    	const block = {
    		c: function create() {
    			create_component(recall.$$.fragment);
    		},
    		m: function mount(target, anchor) {
    			mount_component(recall, target, anchor);
    			current = true;
    		},
    		p: function update(ctx, dirty) {
    			const recall_changes = {};
    			if (dirty & /*playData*/ 4) recall_changes.data = /*playData*/ ctx[2];
    			recall.$set(recall_changes);
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(recall.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(recall.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			destroy_component(recall, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$3.name,
    		type: "if",
    		source: "(262:4) {#if state === 'recall'}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$4(ctx) {
    	let article;
    	let header;
    	let h1;
    	let t1;
    	let button;
    	let t3;
    	let div;
    	let t4;
    	let t5;
    	let current;
    	let dispose;
    	let if_block0 = /*state*/ ctx[3] === "not-playing" && create_if_block_2$1(ctx);
    	let if_block1 = /*state*/ ctx[3] === "playing" && create_if_block_1$2(ctx);
    	let if_block2 = /*state*/ ctx[3] === "recall" && create_if_block$3(ctx);

    	const block = {
    		c: function create() {
    			article = element("article");
    			header = element("header");
    			h1 = element("h1");
    			h1.textContent = "Total Recall";
    			t1 = space();
    			button = element("button");
    			button.textContent = "Close";
    			t3 = space();
    			div = element("div");
    			if (if_block0) if_block0.c();
    			t4 = space();
    			if (if_block1) if_block1.c();
    			t5 = space();
    			if (if_block2) if_block2.c();
    			this.c = noop;
    			attr_dev(h1, "class", "f2 measure");
    			add_location(h1, file$3, 233, 4, 110088);
    			attr_dev(button, "class", "br3");
    			add_location(button, file$3, 234, 4, 110133);
    			add_location(header, file$3, 232, 2, 110075);
    			add_location(div, file$3, 237, 2, 110206);
    			add_location(article, file$3, 231, 0, 110063);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor, remount) {
    			insert_dev(target, article, anchor);
    			append_dev(article, header);
    			append_dev(header, h1);
    			append_dev(header, t1);
    			append_dev(header, button);
    			append_dev(article, t3);
    			append_dev(article, div);
    			if (if_block0) if_block0.m(div, null);
    			append_dev(div, t4);
    			if (if_block1) if_block1.m(div, null);
    			append_dev(div, t5);
    			if (if_block2) if_block2.m(div, null);
    			current = true;
    			if (remount) dispose();
    			dispose = listen_dev(button, "click", /*handleClose*/ ctx[5], false, false, false);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*state*/ ctx[3] === "not-playing") {
    				if (if_block0) {
    					if_block0.p(ctx, dirty);
    				} else {
    					if_block0 = create_if_block_2$1(ctx);
    					if_block0.c();
    					if_block0.m(div, t4);
    				}
    			} else if (if_block0) {
    				if_block0.d(1);
    				if_block0 = null;
    			}

    			if (/*state*/ ctx[3] === "playing") {
    				if (if_block1) {
    					if_block1.p(ctx, dirty);

    					if (dirty & /*state*/ 8) {
    						transition_in(if_block1, 1);
    					}
    				} else {
    					if_block1 = create_if_block_1$2(ctx);
    					if_block1.c();
    					transition_in(if_block1, 1);
    					if_block1.m(div, t5);
    				}
    			} else if (if_block1) {
    				group_outros();

    				transition_out(if_block1, 1, 1, () => {
    					if_block1 = null;
    				});

    				check_outros();
    			}

    			if (/*state*/ ctx[3] === "recall") {
    				if (if_block2) {
    					if_block2.p(ctx, dirty);

    					if (dirty & /*state*/ 8) {
    						transition_in(if_block2, 1);
    					}
    				} else {
    					if_block2 = create_if_block$3(ctx);
    					if_block2.c();
    					transition_in(if_block2, 1);
    					if_block2.m(div, null);
    				}
    			} else if (if_block2) {
    				group_outros();

    				transition_out(if_block2, 1, 1, () => {
    					if_block2 = null;
    				});

    				check_outros();
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(if_block1);
    			transition_in(if_block2);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(if_block1);
    			transition_out(if_block2);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(article);
    			if (if_block0) if_block0.d();
    			if (if_block1) if_block1.d();
    			if (if_block2) if_block2.d();
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$4.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$4($$self, $$props, $$invalidate) {
    	let { listElement } = $$props;
    	let { playElement } = $$props;
    	let { data = [] } = $$props;
    	let { gameSize = 7 } = $$props;
    	let { speed = 1 } = $$props;
    	playElement.style.display = "";
    	listElement.style.display = "none";

    	function handleClose(event) {
    		$$invalidate(10, playElement.style.display = "none", playElement);
    		$$invalidate(9, listElement.style.display = "", listElement);
    		push("/");
    	}

    	let playData = [];

    	// This needs to pick the data
    	let state = "not-playing";

    	const shuffle = arr => arr.map(a => [Math.random(), a]).sort((a, b) => a[0] - b[0]).map(a => a[1]);

    	function play() {
    		// reduce to 7
    		// shuffle
    		let temp = shuffle(data);

    		$$invalidate(2, playData = temp.slice(0, gameSize));
    		$$invalidate(3, state = "playing");
    	}

    	function finished(event) {
    		if (event.detail.playAgain) {
    			play();
    			return;
    		}

    		$$invalidate(3, state = "not-playing");
    	}

    	function handleFinished() {
    		$$invalidate(3, state = "recall");
    	}

    	const writable_props = ["listElement", "playElement", "data", "gameSize", "speed"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	function input0_input_handler() {
    		gameSize = to_number(this.value);
    		$$invalidate(0, gameSize);
    	}

    	function input1_input_handler() {
    		speed = to_number(this.value);
    		$$invalidate(1, speed);
    	}

    	$$self.$set = $$props => {
    		if ("listElement" in $$props) $$invalidate(9, listElement = $$props.listElement);
    		if ("playElement" in $$props) $$invalidate(10, playElement = $$props.playElement);
    		if ("data" in $$props) $$invalidate(11, data = $$props.data);
    		if ("gameSize" in $$props) $$invalidate(0, gameSize = $$props.gameSize);
    		if ("speed" in $$props) $$invalidate(1, speed = $$props.speed);
    	};

    	$$self.$capture_state = () => ({
    		Recall,
    		View,
    		push,
    		listElement,
    		playElement,
    		data,
    		gameSize,
    		speed,
    		handleClose,
    		playData,
    		state,
    		shuffle,
    		play,
    		finished,
    		handleFinished,
    		maxSize
    	});

    	$$self.$inject_state = $$props => {
    		if ("listElement" in $$props) $$invalidate(9, listElement = $$props.listElement);
    		if ("playElement" in $$props) $$invalidate(10, playElement = $$props.playElement);
    		if ("data" in $$props) $$invalidate(11, data = $$props.data);
    		if ("gameSize" in $$props) $$invalidate(0, gameSize = $$props.gameSize);
    		if ("speed" in $$props) $$invalidate(1, speed = $$props.speed);
    		if ("playData" in $$props) $$invalidate(2, playData = $$props.playData);
    		if ("state" in $$props) $$invalidate(3, state = $$props.state);
    		if ("maxSize" in $$props) $$invalidate(4, maxSize = $$props.maxSize);
    	};

    	let maxSize;

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data*/ 2048) {
    			 $$invalidate(4, maxSize = data.length);
    		}
    	};

    	return [
    		gameSize,
    		speed,
    		playData,
    		state,
    		maxSize,
    		handleClose,
    		play,
    		finished,
    		handleFinished,
    		listElement,
    		playElement,
    		data,
    		shuffle,
    		input0_input_handler,
    		input1_input_handler
    	];
    }

    class V1 extends SvelteElement {
    	constructor(options) {
    		super();
    		this.shadowRoot.innerHTML = `<style>h1{font-size:2em;margin:.67em 0}button,input{font-family:inherit;font-size:100%;line-height:1.15;margin:0}button,input{overflow:visible}button{text-transform:none}button{-webkit-appearance:button}button::-moz-focus-inner{border-style:none;padding:0}button:-moz-focusring{outline:1px dotted ButtonText}article,div,h1,header,p{box-sizing:border-box}.br3{border-radius:.5rem}.f2{font-size:2.25rem}.measure{max-width:30em}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){}@media screen and (min-width:60em){}</style>`;

    		init(this, { target: this.shadowRoot }, instance$4, create_fragment$4, safe_not_equal, {
    			listElement: 9,
    			playElement: 10,
    			data: 11,
    			gameSize: 0,
    			speed: 1
    		});

    		const { ctx } = this.$$;
    		const props = this.attributes;

    		if (/*listElement*/ ctx[9] === undefined && !("listElement" in props)) {
    			console.warn("<undefined> was created without expected prop 'listElement'");
    		}

    		if (/*playElement*/ ctx[10] === undefined && !("playElement" in props)) {
    			console.warn("<undefined> was created without expected prop 'playElement'");
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
    		return ["listElement", "playElement", "data", "gameSize", "speed"];
    	}

    	get listElement() {
    		return this.$$.ctx[9];
    	}

    	set listElement(listElement) {
    		this.$set({ listElement });
    		flush();
    	}

    	get playElement() {
    		return this.$$.ctx[10];
    	}

    	set playElement(playElement) {
    		this.$set({ playElement });
    		flush();
    	}

    	get data() {
    		return this.$$.ctx[11];
    	}

    	set data(data) {
    		this.$set({ data });
    		flush();
    	}

    	get gameSize() {
    		return this.$$.ctx[0];
    	}

    	set gameSize(gameSize) {
    		this.$set({ gameSize });
    		flush();
    	}

    	get speed() {
    		return this.$$.ctx[1];
    	}

    	set speed(speed) {
    		this.$set({ speed });
    		flush();
    	}
    }

    /* src/components/interact/routes/total_recall.svelte generated by Svelte v3.22.2 */

    function create_fragment$5(ctx) {
    	let current;

    	const totalrecall = new V1({
    			props: {
    				data: /*aList*/ ctx[0].data,
    				listElement: /*listElement*/ ctx[1],
    				playElement: /*playElement*/ ctx[2]
    			},
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			create_component(totalrecall.$$.fragment);
    			this.c = noop;
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			mount_component(totalrecall, target, anchor);
    			current = true;
    		},
    		p: noop,
    		i: function intro(local) {
    			if (current) return;
    			transition_in(totalrecall.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(totalrecall.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			destroy_component(totalrecall, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$5.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$5($$self, $$props, $$invalidate) {
    	let aList = JSON.parse(document.querySelector("#play-data").innerHTML);
    	let listElement = document.querySelector("#list-info");
    	let playElement = document.querySelector("#play");
    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	$$self.$capture_state = () => ({
    		TotalRecall: V1,
    		aList,
    		listElement,
    		playElement
    	});

    	$$self.$inject_state = $$props => {
    		if ("aList" in $$props) $$invalidate(0, aList = $$props.aList);
    		if ("listElement" in $$props) $$invalidate(1, listElement = $$props.listElement);
    		if ("playElement" in $$props) $$invalidate(2, playElement = $$props.playElement);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [aList, listElement, playElement];
    }

    class Total_recall extends SvelteElement {
    	constructor(options) {
    		super();
    		init(this, { target: this.shadowRoot }, instance$5, create_fragment$5, safe_not_equal, {});

    		if (options) {
    			if (options.target) {
    				insert_dev(options.target, this, options.anchor);
    			}
    		}
    	}
    }

    /* src/components/interact/slideshow/v1.svelte generated by Svelte v3.22.2 */

    const { console: console_1$2 } = globals;
    const file$4 = "src/components/interact/slideshow/v1.svelte";

    // (235:4) {#if loops > 0}
    function create_if_block$4(ctx) {
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
    			add_location(cite, file$4, 235, 6, 110493);
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
    		id: create_if_block$4.name,
    		type: "if",
    		source: "(235:4) {#if loops > 0}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$6(ctx) {
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
    	let if_block = /*loops*/ ctx[0] > 0 && create_if_block$4(ctx);

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
    			add_location(h1, file$4, 228, 4, 110161);
    			attr_dev(button0, "class", "br3");
    			add_location(button0, file$4, 229, 4, 110203);
    			attr_dev(button1, "class", "br3");
    			add_location(button1, file$4, 230, 4, 110260);
    			add_location(header, file$4, 227, 2, 110148);
    			attr_dev(p, "class", "dark-pink f5 f4-m f3-l lh-copy measure mt0");
    			add_location(p, file$4, 233, 4, 110402);
    			attr_dev(blockquote, "class", "athelas ml0 mt4 pl4 black-90 bl bw2 b--black");
    			add_location(blockquote, file$4, 232, 2, 110332);
    			add_location(article, file$4, 226, 0, 110136);
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
    				listen_dev(window, "keydown", /*handleKeydown*/ ctx[4], false, false, false),
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
    					if_block = create_if_block$4(ctx);
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
    		id: create_fragment$6.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$6($$self, $$props, $$invalidate) {
    	let { listElement } = $$props;
    	let { playElement } = $$props;
    	let { aList } = $$props;
    	playElement.style.display = "";
    	listElement.style.display = "none";
    	let loops = 0;
    	let index = -1;
    	let firstTime = "Welcome, to beginning, click next, or use the right arrow key..";
    	let show = firstTime;
    	let nextTimeIsLoop = 0;

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
    		$$invalidate(6, playElement.style.display = "none", playElement);
    		$$invalidate(5, listElement.style.display = "", listElement);
    		push("/");
    	}

    	function handleKeydown(event) {
    		switch (event.code) {
    			case "ArrowLeft":
    				backward();
    				break;
    			case "Space":
    			case "ArrowRight":
    				forward();
    				break;
    			default:
    				console.log(event);
    				console.log(`pressed the ${event.key} key`);
    				break;
    		}
    	}

    	const writable_props = ["listElement", "playElement", "aList"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$2.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	$$self.$set = $$props => {
    		if ("listElement" in $$props) $$invalidate(5, listElement = $$props.listElement);
    		if ("playElement" in $$props) $$invalidate(6, playElement = $$props.playElement);
    		if ("aList" in $$props) $$invalidate(7, aList = $$props.aList);
    	};

    	$$self.$capture_state = () => ({
    		push,
    		listElement,
    		playElement,
    		aList,
    		loops,
    		index,
    		firstTime,
    		show,
    		nextTimeIsLoop,
    		forward,
    		backward,
    		handleClose,
    		handleKeydown
    	});

    	$$self.$inject_state = $$props => {
    		if ("listElement" in $$props) $$invalidate(5, listElement = $$props.listElement);
    		if ("playElement" in $$props) $$invalidate(6, playElement = $$props.playElement);
    		if ("aList" in $$props) $$invalidate(7, aList = $$props.aList);
    		if ("loops" in $$props) $$invalidate(0, loops = $$props.loops);
    		if ("index" in $$props) index = $$props.index;
    		if ("firstTime" in $$props) firstTime = $$props.firstTime;
    		if ("show" in $$props) $$invalidate(1, show = $$props.show);
    		if ("nextTimeIsLoop" in $$props) nextTimeIsLoop = $$props.nextTimeIsLoop;
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [
    		loops,
    		show,
    		forward,
    		handleClose,
    		handleKeydown,
    		listElement,
    		playElement,
    		aList
    	];
    }

    class V1$1 extends SvelteElement {
    	constructor(options) {
    		super();
    		this.shadowRoot.innerHTML = `<style>h1{font-size:2em;margin:.67em 0}button{font-family:inherit;font-size:100%;line-height:1.15;margin:0}button{overflow:visible}button{text-transform:none}button{-webkit-appearance:button}button::-moz-focus-inner{border-style:none;padding:0}button:-moz-focusring{outline:1px dotted ButtonText}article,blockquote,h1,header,p{box-sizing:border-box}.bl{border-left-style:solid;border-left-width:1px}.b--black{border-color:#000}.br3{border-radius:.5rem}.bw2{border-width:.25rem}.athelas{font-family:athelas,georgia,serif}.fs-normal{font-style:normal}.tracked{letter-spacing:.1em}.lh-copy{line-height:1.5}.black-90{color:rgba(0,0,0,.9)}.dark-pink{color:#d5008f}.pl4{padding-left:2rem}.ml0{margin-left:0}.mt0{margin-top:0}.mt4{margin-top:2rem}.ttu{text-transform:uppercase}.f2{font-size:2.25rem}.f5{font-size:1rem}.f6{font-size:.875rem}.measure{max-width:30em}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){.f4-m{font-size:1.25rem}}@media screen and (min-width:60em){.f3-l{font-size:1.5rem}}</style>`;
    		init(this, { target: this.shadowRoot }, instance$6, create_fragment$6, safe_not_equal, { listElement: 5, playElement: 6, aList: 7 });
    		const { ctx } = this.$$;
    		const props = this.attributes;

    		if (/*listElement*/ ctx[5] === undefined && !("listElement" in props)) {
    			console_1$2.warn("<undefined> was created without expected prop 'listElement'");
    		}

    		if (/*playElement*/ ctx[6] === undefined && !("playElement" in props)) {
    			console_1$2.warn("<undefined> was created without expected prop 'playElement'");
    		}

    		if (/*aList*/ ctx[7] === undefined && !("aList" in props)) {
    			console_1$2.warn("<undefined> was created without expected prop 'aList'");
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
    		return ["listElement", "playElement", "aList"];
    	}

    	get listElement() {
    		return this.$$.ctx[5];
    	}

    	set listElement(listElement) {
    		this.$set({ listElement });
    		flush();
    	}

    	get playElement() {
    		return this.$$.ctx[6];
    	}

    	set playElement(playElement) {
    		this.$set({ playElement });
    		flush();
    	}

    	get aList() {
    		return this.$$.ctx[7];
    	}

    	set aList(aList) {
    		this.$set({ aList });
    		flush();
    	}
    }

    /* src/components/interact/routes/slideshow.svelte generated by Svelte v3.22.2 */

    function create_fragment$7(ctx) {
    	let current;

    	const slideshow = new V1$1({
    			props: {
    				aList: /*aList*/ ctx[0],
    				listElement: /*listElement*/ ctx[1],
    				playElement: /*playElement*/ ctx[2]
    			},
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			create_component(slideshow.$$.fragment);
    			this.c = noop;
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			mount_component(slideshow, target, anchor);
    			current = true;
    		},
    		p: noop,
    		i: function intro(local) {
    			if (current) return;
    			transition_in(slideshow.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(slideshow.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			destroy_component(slideshow, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$7.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$7($$self, $$props, $$invalidate) {
    	let aList = JSON.parse(document.querySelector("#play-data").innerHTML);
    	let listElement = document.querySelector("#list-info");
    	let playElement = document.querySelector("#play");
    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	$$self.$capture_state = () => ({
    		Slideshow: V1$1,
    		aList,
    		listElement,
    		playElement
    	});

    	$$self.$inject_state = $$props => {
    		if ("aList" in $$props) $$invalidate(0, aList = $$props.aList);
    		if ("listElement" in $$props) $$invalidate(1, listElement = $$props.listElement);
    		if ("playElement" in $$props) $$invalidate(2, playElement = $$props.playElement);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [aList, listElement, playElement];
    }

    class Slideshow_1 extends SvelteElement {
    	constructor(options) {
    		super();
    		init(this, { target: this.shadowRoot }, instance$7, create_fragment$7, safe_not_equal, {});

    		if (options) {
    			if (options.target) {
    				insert_dev(options.target, this, options.anchor);
    			}
    		}
    	}
    }

    /* src/components/interact/routes/nothing.svelte generated by Svelte v3.22.2 */

    function create_fragment$8(ctx) {
    	const block = {
    		c: function create() {
    			this.c = noop;
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: noop,
    		p: noop,
    		i: noop,
    		o: noop,
    		d: noop
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$8.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$8($$self, $$props, $$invalidate) {
    	let { params } = $$props;
    	document.querySelector("#list-info").style.display = "";
    	document.querySelector("#play").style.display = "none";
    	const writable_props = ["params"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	$$self.$set = $$props => {
    		if ("params" in $$props) $$invalidate(0, params = $$props.params);
    	};

    	$$self.$capture_state = () => ({ params });

    	$$self.$inject_state = $$props => {
    		if ("params" in $$props) $$invalidate(0, params = $$props.params);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [params];
    }

    class Nothing extends SvelteElement {
    	constructor(options) {
    		super();
    		init(this, { target: this.shadowRoot }, instance$8, create_fragment$8, safe_not_equal, { params: 0 });
    		const { ctx } = this.$$;
    		const props = this.attributes;

    		if (/*params*/ ctx[0] === undefined && !("params" in props)) {
    			console.warn("<undefined> was created without expected prop 'params'");
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
    		return ["params"];
    	}

    	get params() {
    		return this.$$.ctx[0];
    	}

    	set params(params) {
    		this.$set({ params });
    		flush();
    	}
    }

    /* src/components/interact/interact.svelte generated by Svelte v3.22.2 */

    const { console: console_1$3 } = globals;

    function create_fragment$9(ctx) {
    	let current;

    	const router = new Router({
    			props: { routes: /*routes*/ ctx[0] },
    			$$inline: true
    		});

    	router.$on("conditionsFailed", /*conditionsFailed*/ ctx[1]);
    	router.$on("routeLoaded", routeLoaded);

    	const block = {
    		c: function create() {
    			create_component(router.$$.fragment);
    			this.c = noop;
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			mount_component(router, target, anchor);
    			current = true;
    		},
    		p: noop,
    		i: function intro(local) {
    			if (current) return;
    			transition_in(router.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(router.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			destroy_component(router, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$9.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function routeLoaded(event) {
    	// eslint-disable-next-line no-console
    	console.info("Caught event routeLoaded", event.detail);
    }

    function instance$9($$self, $$props, $$invalidate) {
    	const routes = {
    		"/play/total_recall": Total_recall,
    		"/play/slideshow": Slideshow_1,
    		// Catch-all, must be last
    		"*": Nothing
    	};

    	// Handles the "conditionsFailed" event dispatched by the router when a component can't be loaded because one of its pre-condition failed
    	function conditionsFailed(event) {
    		replace("/");
    	}

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$3.warn(`<undefined> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("undefined", $$slots, []);

    	$$self.$capture_state = () => ({
    		Router,
    		TotalRecall: Total_recall,
    		Slideshow: Slideshow_1,
    		Nothing,
    		replace,
    		routes,
    		conditionsFailed,
    		routeLoaded
    	});

    	return [routes, conditionsFailed];
    }

    class Interact extends SvelteElement {
    	constructor(options) {
    		super();
    		init(this, { target: this.shadowRoot }, instance$9, create_fragment$9, safe_not_equal, {});

    		if (options) {
    			if (options.target) {
    				insert_dev(options.target, this, options.anchor);
    			}
    		}
    	}
    }

    // Webcomponent
    customElements.define('interact-menu', Menu_wc);

    // Actual app to handle the interactions
    let app;
    const el = document.querySelector("#play-screen");
    if (el) {
        app = new Interact({
            target: el,
        });
    }

    var app$1 = app;

    return app$1;

}());
//# sourceMappingURL=v2.dev.js.map
