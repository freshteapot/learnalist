(function (global, factory) {
    typeof exports === 'object' && typeof module !== 'undefined' ? module.exports = factory() :
    typeof define === 'function' && define.amd ? define(factory) :
    (global = typeof globalThis !== 'undefined' ? globalThis : global || self, global['toolbox-read-write-repeat-v1'] = factory());
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
    function action_destroyer(action_result) {
        return action_result && is_function(action_result.destroy) ? action_result.destroy : noop;
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
    function children(element) {
        return Array.from(element.childNodes);
    }
    function set_input_value(input, value) {
        input.value = value == null ? '' : value;
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
    function set_data_dev(text, data) {
        data = '' + data;
        if (text.wholeText === data)
            return;
        dispatch_dev('SvelteDOMSetData', { node: text, data });
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

    function resize({ target }) {
        target.style.height = "1px";
        target.style.height = (+target.scrollHeight) + "px";
    }

    function text_area_resize(el) {
        resize({ target: el });
        el.style.overflow = 'hidden';
        el.addEventListener('input', resize);

        return {
            destroy: () => el.removeEventListener('input', resize)
        }
    }

    /* src/toolbox/read-write-repeat/v1.svelte generated by Svelte v3.35.0 */

    const { console: console_1 } = globals;
    const file = "src/toolbox/read-write-repeat/v1.svelte";

    function get_each_context(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[11] = list[i];
    	return child_ctx;
    }

    function get_each_context_1(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[11] = list[i];
    	return child_ctx;
    }

    // (71:2) {#if state === "start"}
    function create_if_block_4(ctx) {
    	let div;
    	let h1;
    	let t1;
    	let textarea;
    	let t2;
    	let button;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			div = element("div");
    			h1 = element("h1");
    			h1.textContent = "What shall we train today?";
    			t1 = space();
    			textarea = element("textarea");
    			t2 = space();
    			button = element("button");
    			button.textContent = "Next";
    			attr_dev(h1, "class", "svelte-1vmbms4");
    			add_location(h1, file, 72, 6, 1349);
    			attr_dev(textarea, "class", "db border-box hover-black w-100 ba b--black-20 pa2 br2 mb2 svelte-1vmbms4");
    			attr_dev(textarea, "placeholder", "write something here, and then practice typing it");
    			add_location(textarea, file, 73, 6, 1391);
    			attr_dev(button, "class", "br3 svelte-1vmbms4");
    			add_location(button, file, 79, 6, 1619);
    			attr_dev(div, "class", "outline pa3 mr2 svelte-1vmbms4");
    			add_location(div, file, 71, 4, 1313);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div, anchor);
    			append_dev(div, h1);
    			append_dev(div, t1);
    			append_dev(div, textarea);
    			set_input_value(textarea, /*input*/ ctx[0]);
    			append_dev(div, t2);
    			append_dev(div, button);

    			if (!mounted) {
    				dispose = [
    					listen_dev(textarea, "input", /*textarea_input_handler*/ ctx[10]),
    					action_destroyer(text_area_resize.call(null, textarea)),
    					listen_dev(button, "click", /*nextStep*/ ctx[4], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*input*/ 1) {
    				set_input_value(textarea, /*input*/ ctx[0]);
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_4.name,
    		type: "if",
    		source: "(71:2) {#if state === \\\"start\\\"}",
    		ctx
    	});

    	return block;
    }

    // (84:2) {#if state === "play"}
    function create_if_block_1(ctx) {
    	let div;
    	let h4;
    	let t0;
    	let button;
    	let t2;
    	let t3;
    	let p;
    	let t4;
    	let mounted;
    	let dispose;
    	let if_block0 = /*mark*/ ctx[2].length > 0 && create_if_block_3(ctx);
    	let if_block1 = /*suffix*/ ctx[3].length > 0 && create_if_block_2(ctx);

    	const block = {
    		c: function create() {
    			div = element("div");
    			h4 = element("h4");
    			t0 = text("Start typing,\n        ");
    			button = element("button");
    			button.textContent = "Reset";
    			t2 = text("\n        at any time");
    			t3 = space();
    			p = element("p");
    			if (if_block0) if_block0.c();
    			t4 = space();
    			if (if_block1) if_block1.c();
    			attr_dev(button, "class", "br3 svelte-1vmbms4");
    			add_location(button, file, 87, 8, 1793);
    			attr_dev(h4, "class", "svelte-1vmbms4");
    			add_location(h4, file, 85, 6, 1758);
    			attr_dev(p, "class", "svelte-1vmbms4");
    			add_location(p, file, 90, 6, 1883);
    			attr_dev(div, "class", "outline pa3 mr2 svelte-1vmbms4");
    			add_location(div, file, 84, 4, 1722);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div, anchor);
    			append_dev(div, h4);
    			append_dev(h4, t0);
    			append_dev(h4, button);
    			append_dev(h4, t2);
    			append_dev(div, t3);
    			append_dev(div, p);
    			if (if_block0) if_block0.m(p, null);
    			append_dev(p, t4);
    			if (if_block1) if_block1.m(p, null);

    			if (!mounted) {
    				dispose = listen_dev(button, "click", /*reset*/ ctx[5], false, false, false);
    				mounted = true;
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (/*mark*/ ctx[2].length > 0) {
    				if (if_block0) {
    					if_block0.p(ctx, dirty);
    				} else {
    					if_block0 = create_if_block_3(ctx);
    					if_block0.c();
    					if_block0.m(p, t4);
    				}
    			} else if (if_block0) {
    				if_block0.d(1);
    				if_block0 = null;
    			}

    			if (/*suffix*/ ctx[3].length > 0) {
    				if (if_block1) {
    					if_block1.p(ctx, dirty);
    				} else {
    					if_block1 = create_if_block_2(ctx);
    					if_block1.c();
    					if_block1.m(p, null);
    				}
    			} else if (if_block1) {
    				if_block1.d(1);
    				if_block1 = null;
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    			if (if_block0) if_block0.d();
    			if (if_block1) if_block1.d();
    			mounted = false;
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1.name,
    		type: "if",
    		source: "(84:2) {#if state === \\\"play\\\"}",
    		ctx
    	});

    	return block;
    }

    // (92:8) {#if mark.length > 0}
    function create_if_block_3(ctx) {
    	let each_1_anchor;
    	let each_value_1 = /*mark*/ ctx[2];
    	validate_each_argument(each_value_1);
    	let each_blocks = [];

    	for (let i = 0; i < each_value_1.length; i += 1) {
    		each_blocks[i] = create_each_block_1(get_each_context_1(ctx, each_value_1, i));
    	}

    	const block = {
    		c: function create() {
    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			each_1_anchor = empty();
    		},
    		m: function mount(target, anchor) {
    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(target, anchor);
    			}

    			insert_dev(target, each_1_anchor, anchor);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*mark*/ 4) {
    				each_value_1 = /*mark*/ ctx[2];
    				validate_each_argument(each_value_1);
    				let i;

    				for (i = 0; i < each_value_1.length; i += 1) {
    					const child_ctx = get_each_context_1(ctx, each_value_1, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block_1(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(each_1_anchor.parentNode, each_1_anchor);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value_1.length;
    			}
    		},
    		d: function destroy(detaching) {
    			destroy_each(each_blocks, detaching);
    			if (detaching) detach_dev(each_1_anchor);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_3.name,
    		type: "if",
    		source: "(92:8) {#if mark.length > 0}",
    		ctx
    	});

    	return block;
    }

    // (93:10) {#each mark as w}
    function create_each_block_1(ctx) {
    	let mark_1;
    	let t_value = /*w*/ ctx[11] + "";
    	let t;

    	const block = {
    		c: function create() {
    			mark_1 = element("mark");
    			t = text(t_value);
    			attr_dev(mark_1, "class", "fl pre bg-light-green svelte-1vmbms4");
    			add_location(mark_1, file, 93, 12, 1957);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, mark_1, anchor);
    			append_dev(mark_1, t);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*mark*/ 4 && t_value !== (t_value = /*w*/ ctx[11] + "")) set_data_dev(t, t_value);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(mark_1);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block_1.name,
    		type: "each",
    		source: "(93:10) {#each mark as w}",
    		ctx
    	});

    	return block;
    }

    // (98:8) {#if suffix.length > 0}
    function create_if_block_2(ctx) {
    	let each_1_anchor;
    	let each_value = /*suffix*/ ctx[3];
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

    			each_1_anchor = empty();
    		},
    		m: function mount(target, anchor) {
    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(target, anchor);
    			}

    			insert_dev(target, each_1_anchor, anchor);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*suffix*/ 8) {
    				each_value = /*suffix*/ ctx[3];
    				validate_each_argument(each_value);
    				let i;

    				for (i = 0; i < each_value.length; i += 1) {
    					const child_ctx = get_each_context(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(each_1_anchor.parentNode, each_1_anchor);
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
    			if (detaching) detach_dev(each_1_anchor);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_2.name,
    		type: "if",
    		source: "(98:8) {#if suffix.length > 0}",
    		ctx
    	});

    	return block;
    }

    // (99:10) {#each suffix as w}
    function create_each_block(ctx) {
    	let span;
    	let t_value = /*w*/ ctx[11] + "";
    	let t;

    	const block = {
    		c: function create() {
    			span = element("span");
    			t = text(t_value);
    			attr_dev(span, "class", "pre svelte-1vmbms4");
    			add_location(span, file, 99, 12, 2111);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, span, anchor);
    			append_dev(span, t);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*suffix*/ 8 && t_value !== (t_value = /*w*/ ctx[11] + "")) set_data_dev(t, t_value);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(span);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block.name,
    		type: "each",
    		source: "(99:10) {#each suffix as w}",
    		ctx
    	});

    	return block;
    }

    // (107:2) {#if state === "finished"}
    function create_if_block(ctx) {
    	let div;
    	let h1;
    	let t1;
    	let p;
    	let t2;
    	let t3;
    	let button0;
    	let t5;
    	let button1;
    	let t7;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			div = element("div");
    			h1 = element("h1");
    			h1.textContent = "Done!";
    			t1 = space();
    			p = element("p");
    			t2 = text(/*input*/ ctx[0]);
    			t3 = space();
    			button0 = element("button");
    			button0.textContent = "Reset";
    			t5 = space();
    			button1 = element("button");
    			button1.textContent = "Repeat";
    			t7 = text("\n      (or type to start again)");
    			attr_dev(h1, "class", "svelte-1vmbms4");
    			add_location(h1, file, 108, 6, 2272);
    			attr_dev(p, "class", "svelte-1vmbms4");
    			add_location(p, file, 109, 6, 2293);
    			attr_dev(button0, "class", "br3 svelte-1vmbms4");
    			add_location(button0, file, 110, 6, 2314);
    			attr_dev(button1, "class", "br3 svelte-1vmbms4");
    			add_location(button1, file, 111, 6, 2372);
    			attr_dev(div, "class", "outline pa3 mr2 svelte-1vmbms4");
    			add_location(div, file, 107, 4, 2236);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div, anchor);
    			append_dev(div, h1);
    			append_dev(div, t1);
    			append_dev(div, p);
    			append_dev(p, t2);
    			append_dev(div, t3);
    			append_dev(div, button0);
    			append_dev(div, t5);
    			append_dev(div, button1);
    			append_dev(div, t7);

    			if (!mounted) {
    				dispose = [
    					listen_dev(button0, "click", /*reset*/ ctx[5], false, false, false),
    					listen_dev(button1, "click", /*repeat*/ ctx[6], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*input*/ 1) set_data_dev(t2, /*input*/ ctx[0]);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block.name,
    		type: "if",
    		source: "(107:2) {#if state === \\\"finished\\\"}",
    		ctx
    	});

    	return block;
    }

    function create_fragment(ctx) {
    	let div;
    	let t0;
    	let t1;
    	let mounted;
    	let dispose;
    	let if_block0 = /*state*/ ctx[1] === "start" && create_if_block_4(ctx);
    	let if_block1 = /*state*/ ctx[1] === "play" && create_if_block_1(ctx);
    	let if_block2 = /*state*/ ctx[1] === "finished" && create_if_block(ctx);

    	const block = {
    		c: function create() {
    			div = element("div");
    			if (if_block0) if_block0.c();
    			t0 = space();
    			if (if_block1) if_block1.c();
    			t1 = space();
    			if (if_block2) if_block2.c();
    			attr_dev(div, "class", "flex flex-column w-100 svelte-1vmbms4");
    			add_location(div, file, 69, 0, 1246);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div, anchor);
    			if (if_block0) if_block0.m(div, null);
    			append_dev(div, t0);
    			if (if_block1) if_block1.m(div, null);
    			append_dev(div, t1);
    			if (if_block2) if_block2.m(div, null);

    			if (!mounted) {
    				dispose = listen_dev(window, "keydown", /*handleKeydown*/ ctx[7], false, false, false);
    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*state*/ ctx[1] === "start") {
    				if (if_block0) {
    					if_block0.p(ctx, dirty);
    				} else {
    					if_block0 = create_if_block_4(ctx);
    					if_block0.c();
    					if_block0.m(div, t0);
    				}
    			} else if (if_block0) {
    				if_block0.d(1);
    				if_block0 = null;
    			}

    			if (/*state*/ ctx[1] === "play") {
    				if (if_block1) {
    					if_block1.p(ctx, dirty);
    				} else {
    					if_block1 = create_if_block_1(ctx);
    					if_block1.c();
    					if_block1.m(div, t1);
    				}
    			} else if (if_block1) {
    				if_block1.d(1);
    				if_block1 = null;
    			}

    			if (/*state*/ ctx[1] === "finished") {
    				if (if_block2) {
    					if_block2.p(ctx, dirty);
    				} else {
    					if_block2 = create_if_block(ctx);
    					if_block2.c();
    					if_block2.m(div, null);
    				}
    			} else if (if_block2) {
    				if_block2.d(1);
    				if_block2 = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
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

    function instance($$self, $$props, $$invalidate) {
    	let mark;
    	let suffix;
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots("V1", slots, []);
    	let input = "";
    	let offset = 0;
    	let chars = [];
    	let state = "start";

    	function nextStep() {
    		$$invalidate(1, state = "play");
    		$$invalidate(9, chars = input.trim().split(""));
    	}

    	function reset() {
    		$$invalidate(0, input = "");
    		$$invalidate(8, offset = 0);
    		$$invalidate(9, chars = []);
    		$$invalidate(1, state = "start");
    	}

    	function repeat() {
    		$$invalidate(8, offset = 0);
    		$$invalidate(1, state = "play");
    	}

    	function handleKeydown(event) {
    		if (state !== "play") {
    			if (state === "finished") {
    				repeat();
    			}

    			return;
    		}

    		const key = event.key;
    		console.log("keydown", key, `${chars[offset]}`, chars, event.keyCode, chars[offset].charCodeAt(0));

    		// Handle going back
    		if (key === "Backspace") {
    			$$invalidate(8, offset--, offset);

    			if (offset < 0) {
    				$$invalidate(8, offset = 0);
    			}

    			event.preventDefault();
    			return;
    		}

    		// ↵
    		if (event.keyCode == 13 && chars[offset] == "↵") {
    			console.log("match");
    		}

    		if (chars[offset] === key) {
    			$$invalidate(8, offset++, offset);

    			if (chars[offset] === undefined) {
    				$$invalidate(1, state = "finished");
    			}
    		}
    	}

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1.warn(`<V1> was created with unknown prop '${key}'`);
    	});

    	function textarea_input_handler() {
    		input = this.value;
    		$$invalidate(0, input);
    	}

    	$$self.$capture_state = () => ({
    		text_area_resize,
    		input,
    		offset,
    		chars,
    		state,
    		nextStep,
    		reset,
    		repeat,
    		handleKeydown,
    		mark,
    		suffix
    	});

    	$$self.$inject_state = $$props => {
    		if ("input" in $$props) $$invalidate(0, input = $$props.input);
    		if ("offset" in $$props) $$invalidate(8, offset = $$props.offset);
    		if ("chars" in $$props) $$invalidate(9, chars = $$props.chars);
    		if ("state" in $$props) $$invalidate(1, state = $$props.state);
    		if ("mark" in $$props) $$invalidate(2, mark = $$props.mark);
    		if ("suffix" in $$props) $$invalidate(3, suffix = $$props.suffix);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*chars, offset*/ 768) {
    			$$invalidate(2, mark = chars.slice(0, offset));
    		}

    		if ($$self.$$.dirty & /*chars, offset*/ 768) {
    			$$invalidate(3, suffix = chars.slice(offset));
    		}
    	};

    	return [
    		input,
    		state,
    		mark,
    		suffix,
    		nextStep,
    		reset,
    		repeat,
    		handleKeydown,
    		offset,
    		chars,
    		textarea_input_handler
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
//# sourceMappingURL=toolbox-read-write-repeat-v1.js.map
