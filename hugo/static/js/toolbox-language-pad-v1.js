(function (global, factory) {
    typeof exports === 'object' && typeof module !== 'undefined' ? module.exports = factory() :
    typeof define === 'function' && define.amd ? define(factory) :
    (global = typeof globalThis !== 'undefined' ? globalThis : global || self, global['toolbox-language-pad-v1'] = factory());
}(this, (function () { 'use strict';

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
    function is_empty(obj) {
        return Object.keys(obj).length === 0;
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
    function set_input_value(input, value) {
        input.value = value == null ? '' : value;
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
            throw new Error('Function called outside component initialization');
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
            set_current_component(null);
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

    const globals = (typeof window !== 'undefined'
        ? window
        : typeof globalThis !== 'undefined'
            ? globalThis
            : global);
    function mount_component(component, target, anchor, customElement) {
        const { fragment, on_mount, on_destroy, after_update } = component.$$;
        fragment && fragment.m(target, anchor);
        if (!customElement) {
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
        }
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
            on_disconnect: [],
            before_update: [],
            after_update: [],
            context: new Map(parent_component ? parent_component.$$.context : []),
            // everything else
            callbacks: blank_object(),
            dirty,
            skip_bound: false
        };
        let ready = false;
        $$.ctx = instance
            ? instance(component, options.props || {}, (i, ret, ...rest) => {
                const value = rest.length ? rest[0] : ret;
                if ($$.ctx && not_equal($$.ctx[i], $$.ctx[i] = value)) {
                    if (!$$.skip_bound && $$.bound[i])
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
            mount_component(component, options.target, options.anchor, options.customElement);
            flush();
        }
        set_current_component(parent_component);
    }
    /**
     * Base class for Svelte components. Used when dev=false.
     */
    class SvelteComponent {
        $destroy() {
            destroy_component(this, 1);
            this.$destroy = noop;
        }
        $on(type, callback) {
            const callbacks = (this.$$.callbacks[type] || (this.$$.callbacks[type] = []));
            callbacks.push(callback);
            return () => {
                const index = callbacks.indexOf(callback);
                if (index !== -1)
                    callbacks.splice(index, 1);
            };
        }
        $set($$props) {
            if (this.$$set && !is_empty($$props)) {
                this.$$.skip_bound = true;
                this.$$set($$props);
                this.$$.skip_bound = false;
            }
        }
    }

    function dispatch_dev(type, detail) {
        document.dispatchEvent(custom_event(type, Object.assign({ version: '3.35.0' }, detail)));
    }
    function append_dev(target, node) {
        dispatch_dev('SvelteDOMInsert', { target, node });
        append(target, node);
    }
    function insert_dev(target, node, anchor) {
        dispatch_dev('SvelteDOMInsert', { target, node, anchor });
        insert(target, node, anchor);
    }
    function detach_dev(node) {
        dispatch_dev('SvelteDOMRemove', { node });
        detach(node);
    }
    function listen_dev(node, event, handler, options, has_prevent_default, has_stop_propagation) {
        const modifiers = options === true ? ['capture'] : options ? Array.from(Object.keys(options)) : [];
        if (has_prevent_default)
            modifiers.push('preventDefault');
        if (has_stop_propagation)
            modifiers.push('stopPropagation');
        dispatch_dev('SvelteDOMAddEventListener', { node, event, handler, modifiers });
        const dispose = listen(node, event, handler, options);
        return () => {
            dispatch_dev('SvelteDOMRemoveEventListener', { node, event, handler, modifiers });
            dispose();
        };
    }
    function attr_dev(node, attribute, value) {
        attr(node, attribute, value);
        if (value == null)
            dispatch_dev('SvelteDOMRemoveAttribute', { node, attribute });
        else
            dispatch_dev('SvelteDOMSetAttribute', { node, attribute, value });
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
    /**
     * Base class for Svelte components with some minor dev-enhancements. Used when dev=true.
     */
    class SvelteComponentDev extends SvelteComponent {
        constructor(options) {
            if (!options || (!options.target && !options.$$inline)) {
                throw new Error("'target' is a required option");
            }
            super();
        }
        $destroy() {
            super.$destroy();
            this.$destroy = () => {
                console.warn('Component was already destroyed'); // eslint-disable-line no-console
            };
        }
        $capture_state() { }
        $inject_state() { }
    }

    /* src/toolbox/language-pad/v1.svelte generated by Svelte v3.35.0 */

    const { console: console_1 } = globals;
    const file = "src/toolbox/language-pad/v1.svelte";

    function get_each_context(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[13] = list[i];
    	child_ctx[14] = list;
    	child_ctx[15] = i;
    	return child_ctx;
    }

    // (69:1) {#if locked}
    function create_if_block_2(ctx) {
    	let button0;
    	let t1;
    	let button1;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			button0 = element("button");
    			button0.textContent = "Save";
    			t1 = space();
    			button1 = element("button");
    			button1.textContent = "Clear";
    			attr_dev(button0, "class", "br3 ma2 pa2 svelte-9k6d0q");
    			add_location(button0, file, 69, 2, 1388);
    			attr_dev(button1, "class", "br3 ma2 pa2 svelte-9k6d0q");
    			add_location(button1, file, 76, 2, 1491);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, button0, anchor);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, button1, anchor);

    			if (!mounted) {
    				dispose = [
    					listen_dev(button0, "click", /*click_handler_1*/ ctx[7], false, false, false),
    					listen_dev(button1, "click", /*click_handler_2*/ ctx[8], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(button0);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(button1);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_2.name,
    		type: "if",
    		source: "(69:1) {#if locked}",
    		ctx
    	});

    	return block;
    }

    // (87:2) {#if !locked}
    function create_if_block_1(ctx) {
    	let div1;
    	let div0;
    	let textarea;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			textarea = element("textarea");
    			attr_dev(textarea, "placeholder", "Paste some sentences");
    			attr_dev(textarea, "class", "pv0 mv0 svelte-9k6d0q");
    			attr_dev(textarea, "rows", "1");
    			add_location(textarea, file, 89, 5, 1804);
    			attr_dev(div0, "class", "outline w-75 pa3 mr2 svelte-9k6d0q");
    			add_location(div0, file, 88, 4, 1764);
    			attr_dev(div1, "class", "flex items-center flex-column svelte-9k6d0q");
    			add_location(div1, file, 87, 3, 1716);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div1, anchor);
    			append_dev(div1, div0);
    			append_dev(div0, textarea);

    			if (!mounted) {
    				dispose = listen_dev(textarea, "paste", /*paste_handler*/ ctx[9], false, false, false);
    				mounted = true;
    			}
    		},
    		p: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			mounted = false;
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1.name,
    		type: "if",
    		source: "(87:2) {#if !locked}",
    		ctx
    	});

    	return block;
    }

    // (142:2) {#if locked}
    function create_if_block(ctx) {
    	let div1;
    	let div0;
    	let each_value = /*rows*/ ctx[1];
    	validate_each_argument(each_value);
    	let each_blocks = [];

    	for (let i = 0; i < each_value.length; i += 1) {
    		each_blocks[i] = create_each_block(get_each_context(ctx, each_value, i));
    	}

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			attr_dev(div0, "class", "outline w-75 pa3 mr2 svelte-9k6d0q");
    			add_location(div0, file, 143, 4, 3019);
    			attr_dev(div1, "class", "flex items-center flex-column svelte-9k6d0q");
    			add_location(div1, file, 142, 3, 2971);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div1, anchor);
    			append_dev(div1, div0);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(div0, null);
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*rows*/ 2) {
    				each_value = /*rows*/ ctx[1];
    				validate_each_argument(each_value);
    				let i;

    				for (i = 0; i < each_value.length; i += 1) {
    					const child_ctx = get_each_context(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(div0, null);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value.length;
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			destroy_each(each_blocks, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block.name,
    		type: "if",
    		source: "(142:2) {#if locked}",
    		ctx
    	});

    	return block;
    }

    // (145:5) {#each rows as row, index}
    function create_each_block(ctx) {
    	let textarea;
    	let each_value = /*each_value*/ ctx[14];
    	let index = /*index*/ ctx[15];
    	let mounted;
    	let dispose;

    	function textarea_input_handler() {
    		/*textarea_input_handler*/ ctx[10].call(textarea, /*each_value*/ ctx[14], /*index*/ ctx[15]);
    	}

    	const assign_textarea = () => /*textarea_binding*/ ctx[11](textarea, each_value, index);
    	const unassign_textarea = () => /*textarea_binding*/ ctx[11](null, each_value, index);

    	const block = {
    		c: function create() {
    			textarea = element("textarea");
    			attr_dev(textarea, "class", "pv0 mv0 svelte-9k6d0q");
    			attr_dev(textarea, "learn", "");
    			attr_dev(textarea, "rows", "1");
    			textarea.disabled = !(/*index*/ ctx[15] & 1);
    			toggle_class(textarea, "learn", /*index*/ ctx[15] & 1);
    			add_location(textarea, file, 145, 6, 3092);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, textarea, anchor);
    			set_input_value(textarea, /*row*/ ctx[13].from);
    			assign_textarea();

    			if (!mounted) {
    				dispose = [
    					listen_dev(textarea, "input", textarea_input_handler),
    					listen_dev(textarea, "paste", paste_handler_1, false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(new_ctx, dirty) {
    			ctx = new_ctx;

    			if (dirty & /*rows*/ 2) {
    				set_input_value(textarea, /*row*/ ctx[13].from);
    			}

    			if (each_value !== /*each_value*/ ctx[14] || index !== /*index*/ ctx[15]) {
    				unassign_textarea();
    				each_value = /*each_value*/ ctx[14];
    				index = /*index*/ ctx[15];
    				assign_textarea();
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(textarea);
    			unassign_textarea();
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block.name,
    		type: "each",
    		source: "(145:5) {#each rows as row, index}",
    		ctx
    	});

    	return block;
    }

    function create_fragment(ctx) {
    	let main;
    	let h1;
    	let t1;
    	let button;
    	let t3;
    	let t4;
    	let article;
    	let t5;
    	let mounted;
    	let dispose;
    	let if_block0 = /*locked*/ ctx[2] && create_if_block_2(ctx);
    	let if_block1 = !/*locked*/ ctx[2] && create_if_block_1(ctx);
    	let if_block2 = /*locked*/ ctx[2] && create_if_block(ctx);

    	const block = {
    		c: function create() {
    			main = element("main");
    			h1 = element("h1");
    			h1.textContent = "Notepad";
    			t1 = space();
    			button = element("button");
    			button.textContent = "Restart";
    			t3 = space();
    			if (if_block0) if_block0.c();
    			t4 = space();
    			article = element("article");
    			if (if_block1) if_block1.c();
    			t5 = space();
    			if (if_block2) if_block2.c();
    			attr_dev(h1, "class", "svelte-9k6d0q");
    			add_location(h1, file, 47, 1, 1078);
    			attr_dev(button, "class", "br3 ma2 pa2 svelte-9k6d0q");
    			add_location(button, file, 48, 1, 1096);
    			attr_dev(article, "class", "cf svelte-9k6d0q");
    			add_location(article, file, 85, 1, 1676);
    			attr_dev(main, "class", "svelte-9k6d0q");
    			add_location(main, file, 46, 0, 1070);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, main, anchor);
    			append_dev(main, h1);
    			append_dev(main, t1);
    			append_dev(main, button);
    			append_dev(main, t3);
    			if (if_block0) if_block0.m(main, null);
    			append_dev(main, t4);
    			append_dev(main, article);
    			if (if_block1) if_block1.m(article, null);
    			append_dev(article, t5);
    			if (if_block2) if_block2.m(article, null);

    			if (!mounted) {
    				dispose = listen_dev(button, "click", /*click_handler*/ ctx[6], false, false, false);
    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*locked*/ ctx[2]) {
    				if (if_block0) {
    					if_block0.p(ctx, dirty);
    				} else {
    					if_block0 = create_if_block_2(ctx);
    					if_block0.c();
    					if_block0.m(main, t4);
    				}
    			} else if (if_block0) {
    				if_block0.d(1);
    				if_block0 = null;
    			}

    			if (!/*locked*/ ctx[2]) {
    				if (if_block1) {
    					if_block1.p(ctx, dirty);
    				} else {
    					if_block1 = create_if_block_1(ctx);
    					if_block1.c();
    					if_block1.m(article, t5);
    				}
    			} else if (if_block1) {
    				if_block1.d(1);
    				if_block1 = null;
    			}

    			if (/*locked*/ ctx[2]) {
    				if (if_block2) {
    					if_block2.p(ctx, dirty);
    				} else {
    					if_block2 = create_if_block(ctx);
    					if_block2.c();
    					if_block2.m(article, null);
    				}
    			} else if (if_block2) {
    				if_block2.d(1);
    				if_block2 = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(main);
    			if (if_block0) if_block0.d();
    			if (if_block1) if_block1.d();
    			if (if_block2) if_block2.d();
    			mounted = false;
    			dispose();
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

    function getFromStorage(key, _default) {
    	let temp = localStorage.getItem(key);
    	return temp ? JSON.parse(temp) : _default;
    }

    const paste_handler_1 = event => {
    	event.preventDefault();
    	return false;
    };

    function instance($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots("V1", slots, []);
    	let before = "";
    	let rows = [{ from: "", to: "" }];
    	let mounted = false;
    	let locked = false;
    	let sentenceLength = 100;

    	onMount(async () => {
    		$$invalidate(0, before = getFromStorage("before", ""));
    		$$invalidate(1, rows = getFromStorage("rows", [{ from: "", to: "" }]));
    		$$invalidate(2, locked = getFromStorage("locked", false));
    		$$invalidate(3, sentenceLength = getFromStorage("sentenceLength", 100));
    		mounted = true;
    	});

    	function store(key, data) {
    		if (!mounted) return;
    		localStorage.setItem(key, JSON.stringify(data));
    	}

    	const insertIntoArray = (arr, value) => {
    		return arr.reduce(
    			(result, element, index, array) => {
    				result.push(element);

    				if (index < array.length - 1) {
    					const copy = JSON.parse(JSON.stringify(value));
    					result.push(copy);
    				}

    				return result;
    			},
    			[]
    		);
    	};

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1.warn(`<V1> was created with unknown prop '${key}'`);
    	});

    	const click_handler = () => {
    		$$invalidate(1, rows = [{ from: "", to: "" }]);
    		$$invalidate(2, locked = false);
    		$$invalidate(0, before = "");
    	};

    	const click_handler_1 = () => {
    		console.log("TODO");
    	};

    	const click_handler_2 = () => {
    		$$invalidate(1, rows = rows.map((row, index) => {
    			return !(index & 1) ? row : { from: "", to: "" };
    		}));
    	};

    	const paste_handler = event => {
    		let paste = (event.clipboardData || window.clipboardData).getData("text");
    		paste = paste.trim();
    		store("before", paste);
    		paste = paste.replace(/\n\s*\n/g, "\n");
    		let parts = paste.split("\n");
    		var regex = new RegExp("[\\s\\S]{1," + sentenceLength + "}(?!\\S)", "g");

    		let parts2 = parts.flatMap(e => {
    			let parts = e.//.replace(/[\s\S]{1,100}(?!\S)/g, "$&\n")
    			replace(regex, "$&\n").split("\n");

    			return parts.map(e => {
    				return e.trimStart();
    			});
    		});

    		$$invalidate(1, rows = parts2.map(e => {
    			return { from: e, to: "" };
    		}).filter(e => !(e.from === "" && e.to === "")));

    		$$invalidate(1, rows = [...insertIntoArray(rows, { from: "", to: "" }), { from: "", to: "" }]);
    		$$invalidate(2, locked = true);
    		store("rows", rows);
    		store("locked", locked);
    	};

    	function textarea_input_handler(each_value, index) {
    		each_value[index].from = this.value;
    		$$invalidate(1, rows);
    	}

    	function textarea_binding($$value, each_value, index) {
    		binding_callbacks[$$value ? "unshift" : "push"](() => {
    			each_value[index].elFrom = $$value;
    			$$invalidate(1, rows);
    		});
    	}

    	$$self.$capture_state = () => ({
    		onMount,
    		before,
    		rows,
    		mounted,
    		locked,
    		sentenceLength,
    		getFromStorage,
    		store,
    		insertIntoArray
    	});

    	$$self.$inject_state = $$props => {
    		if ("before" in $$props) $$invalidate(0, before = $$props.before);
    		if ("rows" in $$props) $$invalidate(1, rows = $$props.rows);
    		if ("mounted" in $$props) mounted = $$props.mounted;
    		if ("locked" in $$props) $$invalidate(2, locked = $$props.locked);
    		if ("sentenceLength" in $$props) $$invalidate(3, sentenceLength = $$props.sentenceLength);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*before*/ 1) {
    			store("before", before);
    		}

    		if ($$self.$$.dirty & /*locked*/ 4) {
    			store("locked", locked);
    		}

    		if ($$self.$$.dirty & /*rows*/ 2) {
    			store("rows", rows);
    		}

    		if ($$self.$$.dirty & /*sentenceLength*/ 8) {
    			store("sentenceLength", sentenceLength);
    		}
    	};

    	return [
    		before,
    		rows,
    		locked,
    		sentenceLength,
    		store,
    		insertIntoArray,
    		click_handler,
    		click_handler_1,
    		click_handler_2,
    		paste_handler,
    		textarea_input_handler,
    		textarea_binding
    	];
    }

    class V1 extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance, create_fragment, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "V1",
    			options,
    			id: create_fragment.name
    		});
    	}
    }

    // Auto generated from rollup.config.toolbox.js

    // Actual app to handle the interactions
    let app;
    const el = document.querySelector("#main-panel");
    if (el) {
        app = new V1({
            target: el,
        });
    }

    var app$1 = app;

    return app$1;

})));
//# sourceMappingURL=toolbox-language-pad-v1.js.map
