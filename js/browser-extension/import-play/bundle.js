(function () {
    'use strict';

    function noop() { }
    function assign(tar, src) {
        // @ts-ignore
        for (const k in src)
            tar[k] = src[k];
        return tar;
    }
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
    function create_slot(definition, ctx, $$scope, fn) {
        if (definition) {
            const slot_ctx = get_slot_context(definition, ctx, $$scope, fn);
            return definition[0](slot_ctx);
        }
    }
    function get_slot_context(definition, ctx, $$scope, fn) {
        return definition[1] && fn
            ? assign($$scope.ctx.slice(), definition[1](fn(ctx)))
            : $$scope.ctx;
    }
    function get_slot_changes(definition, $$scope, dirty, fn) {
        if (definition[2] && fn) {
            const lets = definition[2](fn(dirty));
            if ($$scope.dirty === undefined) {
                return lets;
            }
            if (typeof lets === 'object') {
                const merged = [];
                const len = Math.max($$scope.dirty.length, lets.length);
                for (let i = 0; i < len; i += 1) {
                    merged[i] = $$scope.dirty[i] | lets[i];
                }
                return merged;
            }
            return $$scope.dirty | lets;
        }
        return $$scope.dirty;
    }
    function update_slot(slot, slot_definition, ctx, $$scope, dirty, get_slot_changes_fn, get_slot_context_fn) {
        const slot_changes = get_slot_changes(slot_definition, $$scope, dirty, get_slot_changes_fn);
        if (slot_changes) {
            const slot_context = get_slot_context(slot_definition, ctx, $$scope, get_slot_context_fn);
            slot.p(slot_context, slot_changes);
        }
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
    function prevent_default(fn) {
        return function (event) {
            event.preventDefault();
            // @ts-ignore
            return fn.call(this, event);
        };
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
    function get_current_component() {
        if (!current_component)
            throw new Error(`Function called outside component initialization`);
        return current_component;
    }
    function onMount(fn) {
        get_current_component().$$.on_mount.push(fn);
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
        $set() {
            // overridden by instance, if it has props
        }
    }

    function dispatch_dev(type, detail) {
        document.dispatchEvent(custom_event(type, Object.assign({ version: '3.24.0' }, detail)));
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
        if (text.wholeText === data)
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
    class SvelteComponentDev extends SvelteComponent {
        constructor(options) {
            if (!options || (!options.target && !options.$$inline)) {
                throw new Error(`'target' is a required option`);
            }
            super();
        }
        $destroy() {
            super.$destroy();
            this.$destroy = () => {
                console.warn(`Component was already destroyed`); // eslint-disable-line no-console
            };
        }
        $capture_state() { }
        $inject_state() { }
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

    /* node_modules/svelte-spa-router/Router.svelte generated by Svelte v3.24.0 */

    const { Error: Error_1, Object: Object_1, console: console_1 } = globals;

    // (219:0) {:else}
    function create_else_block(ctx) {
    	let switch_instance;
    	let switch_instance_anchor;
    	let current;
    	var switch_value = /*component*/ ctx[0];

    	function switch_props(ctx) {
    		return { $$inline: true };
    	}

    	if (switch_value) {
    		switch_instance = new switch_value(switch_props());
    		switch_instance.$on("routeEvent", /*routeEvent_handler_1*/ ctx[5]);
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
    					switch_instance.$on("routeEvent", /*routeEvent_handler_1*/ ctx[5]);
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
    		source: "(219:0) {:else}",
    		ctx
    	});

    	return block;
    }

    // (217:0) {#if componentParams}
    function create_if_block(ctx) {
    	let switch_instance;
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
    		switch_instance = new switch_value(switch_props(ctx));
    		switch_instance.$on("routeEvent", /*routeEvent_handler*/ ctx[4]);
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
    					switch_instance.$on("routeEvent", /*routeEvent_handler*/ ctx[4]);
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
    		id: create_if_block.name,
    		type: "if",
    		source: "(217:0) {#if componentParams}",
    		ctx
    	});

    	return block;
    }

    function create_fragment(ctx) {
    	let current_block_type_index;
    	let if_block;
    	let if_block_anchor;
    	let current;
    	const if_block_creators = [create_if_block, create_else_block];
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
    		id: create_fragment.name,
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

    const loc = readable(null, // eslint-disable-next-line prefer-arrow-callback
    function start(set) {
    	set(getLocation());

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
    	return tick().then(() => {
    		window.location.hash = (location.charAt(0) == "#" ? "" : "#") + location;
    	});
    }

    function pop() {
    	// Execute this code when the current call stack is complete
    	return tick().then(() => {
    		window.history.back();
    	});
    }

    function replace(location) {
    	if (!location || location.length < 1 || location.charAt(0) != "/" && location.indexOf("#/") !== 0) {
    		throw Error("Invalid parameter location");
    	}

    	// Execute this code when the current call stack is complete
    	return tick().then(() => {
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

    function link(node, hrefVar) {
    	// Only apply to <a> tags
    	if (!node || !node.tagName || node.tagName.toLowerCase() != "a") {
    		throw Error("Action \"link\" can only be used with <a> tags");
    	}

    	updateLink(node, hrefVar || node.getAttribute("href"));

    	return {
    		update(updated) {
    			updateLink(node, updated);
    		}
    	};
    }

    // Internal function used by the link function
    function updateLink(node, href) {
    	// Destination must start with '/'
    	if (!href || href.length < 1 || href.charAt(0) != "/") {
    		throw Error("Invalid value for \"href\" attribute");
    	}

    	// Add # to the href attribute
    	node.setAttribute("href", "#" + href);
    }

    function nextTickPromise(cb) {
    	// eslint-disable-next-line no-console
    	console.warn("nextTickPromise from 'svelte-spa-router' is deprecated and will be removed in version 3; use the 'tick' method from the Svelte runtime instead");

    	return tick().then(cb);
    }

    function instance($$self, $$props, $$invalidate) {
    	let $loc,
    		$$unsubscribe_loc = noop;

    	validate_store(loc, "loc");
    	component_subscribe($$self, loc, $$value => $$invalidate(6, $loc = $$value));
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
    		tick().then(() => {
    			dispatch(name, detail);
    		});
    	};

    	const writable_props = ["routes", "prefix"];

    	Object_1.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1.warn(`<Router> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Router", $$slots, []);

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
    		tick,
    		wrap,
    		getLocation,
    		loc,
    		location,
    		querystring,
    		push,
    		pop,
    		replace,
    		link,
    		updateLink,
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
    		if ($$self.$$.dirty & /*component, $loc*/ 65) {
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
    		routeEvent_handler,
    		routeEvent_handler_1
    	];
    }

    class Router extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance, create_fragment, safe_not_equal, { routes: 2, prefix: 3 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Router",
    			options,
    			id: create_fragment.name
    		});
    	}

    	get routes() {
    		throw new Error_1("<Router>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set routes(value) {
    		throw new Error_1("<Router>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get prefix() {
    		throw new Error_1("<Router>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set prefix(value) {
    		throw new Error_1("<Router>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src/components/login_modal.svelte generated by Svelte v3.24.0 */
    const file = "src/components/login_modal.svelte";

    function create_fragment$1(ctx) {
    	let div0;
    	let t0;
    	let div1;
    	let t1;
    	let button0;
    	let t3;
    	let button1;
    	let current;
    	let mounted;
    	let dispose;
    	const default_slot_template = /*$$slots*/ ctx[2].default;
    	const default_slot = create_slot(default_slot_template, ctx, /*$$scope*/ ctx[1], null);

    	const block = {
    		c: function create() {
    			div0 = element("div");
    			t0 = space();
    			div1 = element("div");
    			if (default_slot) default_slot.c();
    			t1 = space();
    			button0 = element("button");
    			button0.textContent = "Login";
    			t3 = space();
    			button1 = element("button");
    			button1.textContent = "cancel";
    			attr_dev(div0, "class", "modal-background svelte-1dsbof8");
    			add_location(div0, file, 49, 0, 992);
    			attr_dev(button0, "class", "br3");
    			add_location(button0, file, 52, 2, 1113);
    			attr_dev(button1, "class", "br3");
    			add_location(button1, file, 53, 2, 1173);
    			attr_dev(div1, "class", "modal svelte-1dsbof8");
    			attr_dev(div1, "role", "dialog");
    			attr_dev(div1, "aria-modal", "true");
    			add_location(div1, file, 50, 0, 1048);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div0, anchor);
    			insert_dev(target, t0, anchor);
    			insert_dev(target, div1, anchor);

    			if (default_slot) {
    				default_slot.m(div1, null);
    			}

    			append_dev(div1, t1);
    			append_dev(div1, button0);
    			append_dev(div1, t3);
    			append_dev(div1, button1);
    			current = true;

    			if (!mounted) {
    				dispose = [
    					listen_dev(div0, "click", /*handleClose*/ ctx[0], false, false, false),
    					listen_dev(button0, "click", handleLogin, false, false, false),
    					listen_dev(button1, "click", /*handleClose*/ ctx[0], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (default_slot) {
    				if (default_slot.p && dirty & /*$$scope*/ 2) {
    					update_slot(default_slot, default_slot_template, ctx, /*$$scope*/ ctx[1], dirty, null, null);
    				}
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(default_slot, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(default_slot, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div0);
    			if (detaching) detach_dev(t0);
    			if (detaching) detach_dev(div1);
    			if (default_slot) default_slot.d(detaching);
    			mounted = false;
    			run_all(dispose);
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

    function handleLogin() {
    	const searchParams = new URLSearchParams();
    	const redirectUrl = window.location.href.replace(window.location.origin, "");
    	searchParams.set("redirect", redirectUrl);
    	window.location = `/login.html?${searchParams.toString()}`;
    }

    function instance$1($$self, $$props, $$invalidate) {
    	const dispatch = createEventDispatcher();
    	const close = () => dispatch("close");

    	function handleClose() {
    		dispatch("close");
    	}

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Login_modal> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Login_modal", $$slots, ['default']);

    	$$self.$set = $$props => {
    		if ("$$scope" in $$props) $$invalidate(1, $$scope = $$props.$$scope);
    	};

    	$$self.$capture_state = () => ({
    		createEventDispatcher,
    		dispatch,
    		close,
    		handleClose,
    		handleLogin
    	});

    	return [handleClose, $$scope, $$slots];
    }

    class Login_modal extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$1, create_fragment$1, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Login_modal",
    			options,
    			id: create_fragment$1.name
    		});
    	}
    }

    function get(key, defaultResult) {
        if (!localStorage.hasOwnProperty(key)) {
            return defaultResult;
        }

        return JSON.parse(localStorage.getItem(key))
    }

    function save(key, data) {
        localStorage.setItem(key, JSON.stringify(data));
    }

    function rm(key) {
        localStorage.removeItem(key);
    }

    function getApiServer() {
        const apiServer = document.querySelector('meta[name="api.server"]');
        return apiServer ? apiServer.content : "https://learnalist.net";
    }

    const KeySettingsServer = "settings.server";
    const KeySettingsInstallDefaults = "settings.install.defaults";
    const KeyUserAuthentication = "app.user.authentication";
    const KeyNotifications = "app.notifications";
    const KeyEditorMyEditedLists = "my.edited.lists";
    const KeyEditorMyLists = "my.lists";

    function clear() {
      localStorage.clear();
      save(KeySettingsInstallDefaults, true);
      save(KeySettingsServer, getApiServer());
      save(KeyEditorMyEditedLists, []);
      save(KeyEditorMyLists, []);
    }

    const clearConfiguration = clear;
    const saveConfiguration = save;
    const removeConfiguration = rm;
    const getConfiguration = get;

    function copyObject(item) {
        return JSON.parse(JSON.stringify(item))
    }

    const data = {
        level: "",
        message: "",
        sticky: false,
    };

    const emptyData = JSON.parse(JSON.stringify(data));
    let liveData = JSON.parse(JSON.stringify(data));


    const storedData = getConfiguration(KeyNotifications, null);

    if (storedData !== null) {
        liveData = storedData;
    }

    const { subscribe: subscribe$1, update: update$1, set } = writable(liveData);

    function wrapper() {
        return {
            subscribe: subscribe$1,

            add: (level, message, sticky) => {
                if (sticky == undefined) {
                    sticky = false;
                }

                update$1(notification => {
                    notification.level = level;
                    notification.message = message;
                    notification.sticky = sticky;
                    saveConfiguration(KeyNotifications, notification);
                    return notification;
                });
            },

            clear: () => {
                removeConfiguration(KeyNotifications);
                set(copyObject(emptyData));
            }
        };
    }

    const notifications = wrapper();

    /*! *****************************************************************************
    Copyright (c) Microsoft Corporation.

    Permission to use, copy, modify, and/or distribute this software for any
    purpose with or without fee is hereby granted.

    THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
    REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
    AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
    INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
    LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
    OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
    PERFORMANCE OF THIS SOFTWARE.
    ***************************************************************************** */

    function __awaiter(thisArg, _arguments, P, generator) {
        function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
        return new (P || (P = Promise))(function (resolve, reject) {
            function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
            function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
            function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
            step((generator = generator.apply(thisArg, _arguments || [])).next());
        });
    }

    const BASE_PATH = "https://learnalist.net/api/v1".replace(/\/+$/, "");
    const isBlob = (value) => typeof Blob !== 'undefined' && value instanceof Blob;
    class BaseAPI {
        constructor(configuration = new Configuration()) {
            this.configuration = configuration;
            this.fetchApi = (url, init) => __awaiter(this, void 0, void 0, function* () {
                let fetchParams = { url, init };
                for (const middleware of this.middleware) {
                    if (middleware.pre) {
                        fetchParams = (yield middleware.pre(Object.assign({ fetch: this.fetchApi }, fetchParams))) || fetchParams;
                    }
                }
                let response = yield this.configuration.fetchApi(fetchParams.url, fetchParams.init);
                for (const middleware of this.middleware) {
                    if (middleware.post) {
                        response = (yield middleware.post({
                            fetch: this.fetchApi,
                            url,
                            init,
                            response: response.clone(),
                        })) || response;
                    }
                }
                return response;
            });
            this.middleware = configuration.middleware;
        }
        withMiddleware(...middlewares) {
            const next = this.clone();
            next.middleware = next.middleware.concat(...middlewares);
            return next;
        }
        withPreMiddleware(...preMiddlewares) {
            const middlewares = preMiddlewares.map((pre) => ({ pre }));
            return this.withMiddleware(...middlewares);
        }
        withPostMiddleware(...postMiddlewares) {
            const middlewares = postMiddlewares.map((post) => ({ post }));
            return this.withMiddleware(...middlewares);
        }
        request(context) {
            return __awaiter(this, void 0, void 0, function* () {
                const { url, init } = this.createFetchParams(context);
                const response = yield this.fetchApi(url, init);
                if (response.status >= 200 && response.status < 300) {
                    return response;
                }
                throw response;
            });
        }
        createFetchParams(context) {
            let url = this.configuration.basePath + context.path;
            if (context.query !== undefined && Object.keys(context.query).length !== 0) {
                url += '?' + this.configuration.queryParamsStringify(context.query);
            }
            const body = ((typeof FormData !== "undefined" && context.body instanceof FormData) || context.body instanceof URLSearchParams || isBlob(context.body))
                ? context.body
                : JSON.stringify(context.body);
            const headers = Object.assign({}, this.configuration.headers, context.headers);
            const init = {
                method: context.method,
                headers: headers,
                body,
                credentials: this.configuration.credentials
            };
            return { url, init };
        }
        clone() {
            const constructor = this.constructor;
            const next = new constructor(this.configuration);
            next.middleware = this.middleware.slice();
            return next;
        }
    }
    class RequiredError extends Error {
        constructor(field, msg) {
            super(msg);
            this.field = field;
            this.name = "RequiredError";
        }
    }
    class Configuration {
        constructor(configuration = {}) {
            this.configuration = configuration;
        }
        get basePath() {
            return this.configuration.basePath != null ? this.configuration.basePath : BASE_PATH;
        }
        get fetchApi() {
            return this.configuration.fetchApi || window.fetch.bind(window);
        }
        get middleware() {
            return this.configuration.middleware || [];
        }
        get queryParamsStringify() {
            return this.configuration.queryParamsStringify || querystring$1;
        }
        get username() {
            return this.configuration.username;
        }
        get password() {
            return this.configuration.password;
        }
        get apiKey() {
            const apiKey = this.configuration.apiKey;
            if (apiKey) {
                return typeof apiKey === 'function' ? apiKey : () => apiKey;
            }
            return undefined;
        }
        get accessToken() {
            const accessToken = this.configuration.accessToken;
            if (accessToken) {
                return typeof accessToken === 'function' ? accessToken : () => accessToken;
            }
            return undefined;
        }
        get headers() {
            return this.configuration.headers;
        }
        get credentials() {
            return this.configuration.credentials;
        }
    }
    function exists(json, key) {
        const value = json[key];
        return value !== null && value !== undefined;
    }
    function querystring$1(params, prefix = '') {
        return Object.keys(params)
            .map((key) => {
            const fullKey = prefix + (prefix.length ? `[${key}]` : key);
            const value = params[key];
            if (value instanceof Array) {
                const multiValue = value.map(singleValue => encodeURIComponent(String(singleValue)))
                    .join(`&${encodeURIComponent(fullKey)}=`);
                return `${encodeURIComponent(fullKey)}=${multiValue}`;
            }
            if (value instanceof Object) {
                return querystring$1(value, fullKey);
            }
            return `${encodeURIComponent(fullKey)}=${encodeURIComponent(String(value))}`;
        })
            .filter(part => part.length > 0)
            .join('&');
    }
    class JSONApiResponse {
        constructor(raw, transformer = (jsonValue) => jsonValue) {
            this.raw = raw;
            this.transformer = transformer;
        }
        value() {
            return __awaiter(this, void 0, void 0, function* () {
                return this.transformer(yield this.raw.json());
            });
        }
    }
    class VoidApiResponse {
        constructor(raw) {
            this.raw = raw;
        }
        value() {
            return __awaiter(this, void 0, void 0, function* () {
                return undefined;
            });
        }
    }

    function AlistFromJSON(json) {
        return AlistFromJSONTyped(json);
    }
    function AlistFromJSONTyped(json, ignoreDiscriminator) {
        if ((json === undefined) || (json === null)) {
            return json;
        }
        return {
            'uuid': json['uuid'],
            'info': AlistInfoFromJSON(json['info']),
            'data': AnyTypeFromJSON(json['data']),
        };
    }
    function AlistToJSON(value) {
        if (value === undefined) {
            return undefined;
        }
        if (value === null) {
            return null;
        }
        return {
            'uuid': value.uuid,
            'info': AlistInfoToJSON(value.info),
            'data': AnyTypeToJSON(value.data),
        };
    }

    function AlistInfoFromJSON(json) {
        return AlistInfoFromJSONTyped(json);
    }
    function AlistInfoFromJSONTyped(json, ignoreDiscriminator) {
        if ((json === undefined) || (json === null)) {
            return json;
        }
        return {
            'title': json['title'],
            'type': json['type'],
            'labels': !exists(json, 'labels') ? undefined : json['labels'],
            'shared_with': !exists(json, 'shared_with') ? undefined : json['shared_with'],
            'interact': !exists(json, 'interact') ? undefined : AlistInteractFromJSON(json['interact']),
        };
    }
    function AlistInfoToJSON(value) {
        if (value === undefined) {
            return undefined;
        }
        if (value === null) {
            return null;
        }
        return {
            'title': value.title,
            'type': value.type,
            'labels': value.labels,
            'shared_with': value.shared_with,
            'interact': AlistInteractToJSON(value.interact),
        };
    }

    function AlistInputFromJSON(json) {
        return AlistInputFromJSONTyped(json);
    }
    function AlistInputFromJSONTyped(json, ignoreDiscriminator) {
        if ((json === undefined) || (json === null)) {
            return json;
        }
        return {
            'info': AlistInfoFromJSON(json['info']),
            'data': AnyTypeFromJSON(json['data']),
        };
    }
    function AlistInputToJSON(value) {
        if (value === undefined) {
            return undefined;
        }
        if (value === null) {
            return null;
        }
        return {
            'info': AlistInfoToJSON(value.info),
            'data': AnyTypeToJSON(value.data),
        };
    }

    function AlistInteractFromJSON(json) {
        return AlistInteractFromJSONTyped(json);
    }
    function AlistInteractFromJSONTyped(json, ignoreDiscriminator) {
        if ((json === undefined) || (json === null)) {
            return json;
        }
        return {
            'slideshow': !exists(json, 'slideshow') ? undefined : json['slideshow'],
            'totalrecall': !exists(json, 'totalrecall') ? undefined : json['totalrecall'],
        };
    }
    function AlistInteractToJSON(value) {
        if (value === undefined) {
            return undefined;
        }
        if (value === null) {
            return null;
        }
        return {
            'slideshow': value.slideshow,
            'totalrecall': value.totalrecall,
        };
    }

    function AnyTypeFromJSON(json) {
        return AnyTypeFromJSONTyped(json);
    }
    function AnyTypeFromJSONTyped(json, ignoreDiscriminator) {
        return json;
    }
    function AnyTypeToJSON(value) {
        return value;
    }

    function HttpResponseMessageFromJSON(json) {
        return HttpResponseMessageFromJSONTyped(json);
    }
    function HttpResponseMessageFromJSONTyped(json, ignoreDiscriminator) {
        if ((json === undefined) || (json === null)) {
            return json;
        }
        return {
            'message': !exists(json, 'message') ? undefined : json['message'],
        };
    }

    function HttpUserLoginRequestToJSON(value) {
        return value;
    }

    function HttpUserLoginResponseFromJSON(json) {
        return HttpUserLoginResponseFromJSONTyped(json);
    }
    function HttpUserLoginResponseFromJSONTyped(json, ignoreDiscriminator) {
        if ((json === undefined) || (json === null)) {
            return json;
        }
        return {
            'token': json['token'],
            'user_uuid': json['user_uuid'],
        };
    }

    function HttpUserRegisterInputToJSON(value) {
        if (value === undefined) {
            return undefined;
        }
        if (value === null) {
            return null;
        }
        return {
            'username': value.username,
            'password': value.password,
        };
    }

    function HttpUserRegisterResponseFromJSON(json) {
        return HttpUserRegisterResponseFromJSONTyped(json);
    }
    function HttpUserRegisterResponseFromJSONTyped(json, ignoreDiscriminator) {
        if ((json === undefined) || (json === null)) {
            return json;
        }
        return {
            'username': !exists(json, 'username') ? undefined : json['username'],
            'uuid': !exists(json, 'uuid') ? undefined : json['uuid'],
        };
    }

    function SpacedRepetitionEntryViewedToJSON(value) {
        if (value === undefined) {
            return undefined;
        }
        if (value === null) {
            return null;
        }
        return {
            'uuid': value.uuid,
            'action': value.action,
        };
    }

    function VersionFromJSON(json) {
        return VersionFromJSONTyped(json);
    }
    function VersionFromJSONTyped(json, ignoreDiscriminator) {
        if ((json === undefined) || (json === null)) {
            return json;
        }
        return {
            'gitHash': !exists(json, 'gitHash') ? undefined : json['gitHash'],
            'gitDate': !exists(json, 'gitDate') ? undefined : json['gitDate'],
            'version': !exists(json, 'version') ? undefined : json['version'],
            'url': !exists(json, 'url') ? undefined : json['url'],
        };
    }

    class AListApi extends BaseAPI {
        addListRaw(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                if (requestParameters.alistInput === null || requestParameters.alistInput === undefined) {
                    throw new RequiredError('alistInput', 'Required parameter requestParameters.alistInput was null or undefined when calling addList.');
                }
                const queryParameters = {};
                const headerParameters = {};
                headerParameters['Content-Type'] = 'application/json';
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/alist`,
                    method: 'POST',
                    headers: headerParameters,
                    query: queryParameters,
                    body: AlistInputToJSON(requestParameters.alistInput),
                });
                return new JSONApiResponse(response, (jsonValue) => AlistFromJSON(jsonValue));
            });
        }
        addList(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.addListRaw(requestParameters);
                return yield response.value();
            });
        }
        deleteListByUuidRaw(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                if (requestParameters.uuid === null || requestParameters.uuid === undefined) {
                    throw new RequiredError('uuid', 'Required parameter requestParameters.uuid was null or undefined when calling deleteListByUuid.');
                }
                const queryParameters = {};
                const headerParameters = {};
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/alist/{uuid}`.replace(`{${"uuid"}}`, encodeURIComponent(String(requestParameters.uuid))),
                    method: 'DELETE',
                    headers: headerParameters,
                    query: queryParameters,
                });
                return new JSONApiResponse(response, (jsonValue) => HttpResponseMessageFromJSON(jsonValue));
            });
        }
        deleteListByUuid(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.deleteListByUuidRaw(requestParameters);
                return yield response.value();
            });
        }
        getListByUuidRaw(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                if (requestParameters.uuid === null || requestParameters.uuid === undefined) {
                    throw new RequiredError('uuid', 'Required parameter requestParameters.uuid was null or undefined when calling getListByUuid.');
                }
                const queryParameters = {};
                const headerParameters = {};
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/alist/{uuid}`.replace(`{${"uuid"}}`, encodeURIComponent(String(requestParameters.uuid))),
                    method: 'GET',
                    headers: headerParameters,
                    query: queryParameters,
                });
                return new JSONApiResponse(response, (jsonValue) => AlistFromJSON(jsonValue));
            });
        }
        getListByUuid(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.getListByUuidRaw(requestParameters);
                return yield response.value();
            });
        }
        getListsByMeRaw(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                const queryParameters = {};
                if (requestParameters.labels !== undefined) {
                    queryParameters['labels'] = requestParameters.labels;
                }
                if (requestParameters.listType !== undefined) {
                    queryParameters['list_type'] = requestParameters.listType;
                }
                const headerParameters = {};
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/alist/by/me`,
                    method: 'GET',
                    headers: headerParameters,
                    query: queryParameters,
                });
                return new JSONApiResponse(response, (jsonValue) => jsonValue.map(AlistFromJSON));
            });
        }
        getListsByMe(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.getListsByMeRaw(requestParameters);
                return yield response.value();
            });
        }
        updateListByUuidRaw(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                if (requestParameters.uuid === null || requestParameters.uuid === undefined) {
                    throw new RequiredError('uuid', 'Required parameter requestParameters.uuid was null or undefined when calling updateListByUuid.');
                }
                if (requestParameters.alist === null || requestParameters.alist === undefined) {
                    throw new RequiredError('alist', 'Required parameter requestParameters.alist was null or undefined when calling updateListByUuid.');
                }
                const queryParameters = {};
                const headerParameters = {};
                headerParameters['Content-Type'] = 'application/json';
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/alist/{uuid}`.replace(`{${"uuid"}}`, encodeURIComponent(String(requestParameters.uuid))),
                    method: 'PUT',
                    headers: headerParameters,
                    query: queryParameters,
                    body: AlistToJSON(requestParameters.alist),
                });
                return new JSONApiResponse(response, (jsonValue) => AlistFromJSON(jsonValue));
            });
        }
        updateListByUuid(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.updateListByUuidRaw(requestParameters);
                return yield response.value();
            });
        }
    }

    class DefaultApi extends BaseAPI {
        getServerVersionRaw() {
            return __awaiter(this, void 0, void 0, function* () {
                const queryParameters = {};
                const headerParameters = {};
                const response = yield this.request({
                    path: `/version`,
                    method: 'GET',
                    headers: headerParameters,
                    query: queryParameters,
                });
                return new JSONApiResponse(response, (jsonValue) => VersionFromJSON(jsonValue));
            });
        }
        getServerVersion() {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.getServerVersionRaw();
                return yield response.value();
            });
        }
    }

    class SpacedRepetitionApi extends BaseAPI {
        addSpacedRepetitionEntryRaw(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                if (requestParameters.body === null || requestParameters.body === undefined) {
                    throw new RequiredError('body', 'Required parameter requestParameters.body was null or undefined when calling addSpacedRepetitionEntry.');
                }
                const queryParameters = {};
                const headerParameters = {};
                headerParameters['Content-Type'] = 'application/json';
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/spaced-repetition/`,
                    method: 'POST',
                    headers: headerParameters,
                    query: queryParameters,
                    body: requestParameters.body,
                });
                return new JSONApiResponse(response);
            });
        }
        addSpacedRepetitionEntry(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.addSpacedRepetitionEntryRaw(requestParameters);
                return yield response.value();
            });
        }
        deleteSpacedRepetitionEntryRaw(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                if (requestParameters.uuid === null || requestParameters.uuid === undefined) {
                    throw new RequiredError('uuid', 'Required parameter requestParameters.uuid was null or undefined when calling deleteSpacedRepetitionEntry.');
                }
                const queryParameters = {};
                const headerParameters = {};
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/spaced-repetition/{uuid}`.replace(`{${"uuid"}}`, encodeURIComponent(String(requestParameters.uuid))),
                    method: 'DELETE',
                    headers: headerParameters,
                    query: queryParameters,
                });
                return new VoidApiResponse(response);
            });
        }
        deleteSpacedRepetitionEntry(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                yield this.deleteSpacedRepetitionEntryRaw(requestParameters);
            });
        }
        getNextSpacedRepetitionEntryRaw() {
            return __awaiter(this, void 0, void 0, function* () {
                const queryParameters = {};
                const headerParameters = {};
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/spaced-repetition/next`,
                    method: 'GET',
                    headers: headerParameters,
                    query: queryParameters,
                });
                return new JSONApiResponse(response);
            });
        }
        getNextSpacedRepetitionEntry() {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.getNextSpacedRepetitionEntryRaw();
                return yield response.value();
            });
        }
        getSpacedRepetitionEntriesRaw() {
            return __awaiter(this, void 0, void 0, function* () {
                const queryParameters = {};
                const headerParameters = {};
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/spaced-repetition/all`,
                    method: 'GET',
                    headers: headerParameters,
                    query: queryParameters,
                });
                return new JSONApiResponse(response);
            });
        }
        getSpacedRepetitionEntries() {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.getSpacedRepetitionEntriesRaw();
                return yield response.value();
            });
        }
        updateSpacedRepetitionEntryRaw(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                if (requestParameters.spacedRepetitionEntryViewed === null || requestParameters.spacedRepetitionEntryViewed === undefined) {
                    throw new RequiredError('spacedRepetitionEntryViewed', 'Required parameter requestParameters.spacedRepetitionEntryViewed was null or undefined when calling updateSpacedRepetitionEntry.');
                }
                const queryParameters = {};
                const headerParameters = {};
                headerParameters['Content-Type'] = 'application/json';
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/spaced-repetition/viewed`,
                    method: 'POST',
                    headers: headerParameters,
                    query: queryParameters,
                    body: SpacedRepetitionEntryViewedToJSON(requestParameters.spacedRepetitionEntryViewed),
                });
                return new JSONApiResponse(response);
            });
        }
        updateSpacedRepetitionEntry(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.updateSpacedRepetitionEntryRaw(requestParameters);
                return yield response.value();
            });
        }
    }

    class UserApi extends BaseAPI {
        deleteUserRaw(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                if (requestParameters.uuid === null || requestParameters.uuid === undefined) {
                    throw new RequiredError('uuid', 'Required parameter requestParameters.uuid was null or undefined when calling deleteUser.');
                }
                const queryParameters = {};
                const headerParameters = {};
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/user/{uuid}`.replace(`{${"uuid"}}`, encodeURIComponent(String(requestParameters.uuid))),
                    method: 'DELETE',
                    headers: headerParameters,
                    query: queryParameters,
                });
                return new JSONApiResponse(response, (jsonValue) => HttpResponseMessageFromJSON(jsonValue));
            });
        }
        deleteUser(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.deleteUserRaw(requestParameters);
                return yield response.value();
            });
        }
        loginWithUsernameAndPasswordRaw(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                if (requestParameters.httpUserLoginRequest === null || requestParameters.httpUserLoginRequest === undefined) {
                    throw new RequiredError('httpUserLoginRequest', 'Required parameter requestParameters.httpUserLoginRequest was null or undefined when calling loginWithUsernameAndPassword.');
                }
                const queryParameters = {};
                const headerParameters = {};
                headerParameters['Content-Type'] = 'application/json';
                const response = yield this.request({
                    path: `/user/login`,
                    method: 'POST',
                    headers: headerParameters,
                    query: queryParameters,
                    body: HttpUserLoginRequestToJSON(requestParameters.httpUserLoginRequest),
                });
                return new JSONApiResponse(response, (jsonValue) => HttpUserLoginResponseFromJSON(jsonValue));
            });
        }
        loginWithUsernameAndPassword(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.loginWithUsernameAndPasswordRaw(requestParameters);
                return yield response.value();
            });
        }
        registerUserWithUsernameAndPasswordRaw(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                if (requestParameters.httpUserRegisterInput === null || requestParameters.httpUserRegisterInput === undefined) {
                    throw new RequiredError('httpUserRegisterInput', 'Required parameter requestParameters.httpUserRegisterInput was null or undefined when calling registerUserWithUsernameAndPassword.');
                }
                const queryParameters = {};
                const headerParameters = {};
                headerParameters['Content-Type'] = 'application/json';
                if (this.configuration && (this.configuration.username !== undefined || this.configuration.password !== undefined)) {
                    headerParameters["Authorization"] = "Basic " + btoa(this.configuration.username + ":" + this.configuration.password);
                }
                if (this.configuration && this.configuration.accessToken) {
                    const token = this.configuration.accessToken;
                    const tokenString = typeof token === 'function' ? token("bearerAuth", []) : token;
                    if (tokenString) {
                        headerParameters["Authorization"] = `Bearer ${tokenString}`;
                    }
                }
                const response = yield this.request({
                    path: `/user/register`,
                    method: 'POST',
                    headers: headerParameters,
                    query: queryParameters,
                    body: HttpUserRegisterInputToJSON(requestParameters.httpUserRegisterInput),
                });
                return new JSONApiResponse(response, (jsonValue) => HttpUserRegisterResponseFromJSON(jsonValue));
            });
        }
        registerUserWithUsernameAndPassword(requestParameters) {
            return __awaiter(this, void 0, void 0, function* () {
                const response = yield this.registerUserWithUsernameAndPasswordRaw(requestParameters);
                return yield response.value();
            });
        }
    }

    const Services = {
      Default: DefaultApi,
      User: UserApi,
      Alist: AListApi,
      SpacedRepetition: SpacedRepetitionApi
    };

    function getServer() {
      const server = getConfiguration(KeySettingsServer, null);
      if (server === null) {
        throw new Error('settings.server.missing');
      }
      return server;
    }

    // getApi service = One of the services based on Services
    function getApi(service) {
      var config = new Configuration({
        basePath: `${getServer()}/api/v1`,
        accessToken: getConfiguration(KeyUserAuthentication, undefined),
      });

      return new service(config);
    }

    async function addList(aList) {
      try {
        const api = getApi(Services.Alist);

        const input = {
          alistInput: AlistInputFromJSON(aList)
        };
        return await api.addList(input);
      } catch (error) {
        console.error(error);
        throw new Error("Failed to save list");
      }
    }


    async function addSpacedRepetitionEntry(entry) {
      const response = {
        status: 500,
        body: {}
      };

      try {
        const api = getApi(Services.SpacedRepetition);
        const input = {
          body: entry,
        };
        const res = await api.addSpacedRepetitionEntryRaw(input);
        response.status = res.raw.status;
        response.body = await res.value();
        return response;
      } catch (error) {
        response.status = error.status;
        response.body = await error.json();
        return response;
      }
    }

    // Link any component to be able to notify the banner component
    const notify = (level, message, sticky) => {
        notifications.add(level, message, sticky);
    };

    const loggedIn = () => {
        return localStorage.hasOwnProperty(KeyUserAuthentication);
    };

    const emptyData$1 = {};

    let loaded = false;
    let data$1 = copyObject(emptyData$1);
    const { subscribe: subscribe$2, set: set$1, update: update$2 } = writable(data$1);
    const loading = writable(false);
    const error = writable('');


    const load = async (aList) => {
      set$1(aList);
      loaded = true;
    };

    const save$1 = async (input) => {
      try {
        error.set('');
        loading.set(true);

        console.log(data$1 === input);
        input.info.type = "v2";
        // TODO this is not being saved, I suspect due to openapi
        input.info.from = "https://quizlet.com/71954111/norwegian-flash-cards/";

        let aList = await addList(input);
        set$1(aList);

      } catch (e) {
        console.log(e);
        loading.set(false);
        error.set(`Error has been occurred. Details: ${e.message}`);
      }

    };

    const ImportPlayStore = () => ({
      subscribe: subscribe$2,
      loading,
      error,
      load,
      save: save$1,
      loaded: () => loaded,
      getServer: () => getServer()
    });

    var store = ImportPlayStore();

    /* src/browser-extension/import-play/info.svelte generated by Svelte v3.24.0 */

    const { console: console_1$1 } = globals;
    const file$1 = "src/browser-extension/import-play/info.svelte";

    function get_each_context(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[10] = list[i];
    	child_ctx[12] = i;
    	return child_ctx;
    }

    // (49:0) {#if loggedIn()}
    function create_if_block_3(ctx) {
    	let button0;
    	let t1;
    	let button1;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			button0 = element("button");
    			button0.textContent = "🧠 + 💪";
    			t1 = space();
    			button1 = element("button");
    			button1.textContent = "Save to Learnalist";
    			attr_dev(button0, "class", "br3");
    			add_location(button0, file$1, 53, 2, 1333);
    			attr_dev(button1, "class", "br3");
    			add_location(button1, file$1, 56, 2, 1437);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, button0, anchor);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, button1, anchor);

    			if (!mounted) {
    				dispose = [
    					listen_dev(button0, "click", /*click_handler_3*/ ctx[7], false, false, false),
    					listen_dev(button1, "click", /*handleSave*/ ctx[3], false, false, false)
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
    		id: create_if_block_3.name,
    		type: "if",
    		source: "(49:0) {#if loggedIn()}",
    		ctx
    	});

    	return block;
    }

    // (60:0) {#if show == 'overview'}
    function create_if_block_2(ctx) {
    	let h1;
    	let t0_value = /*aList*/ ctx[0].info.title + "";
    	let t0;
    	let t1;
    	let div;
    	let table;
    	let thead;
    	let tr;
    	let th0;
    	let t3;
    	let th1;
    	let t5;
    	let tbody;
    	let each_value = /*aList*/ ctx[0].data;
    	validate_each_argument(each_value);
    	let each_blocks = [];

    	for (let i = 0; i < each_value.length; i += 1) {
    		each_blocks[i] = create_each_block(get_each_context(ctx, each_value, i));
    	}

    	const block = {
    		c: function create() {
    			h1 = element("h1");
    			t0 = text(t0_value);
    			t1 = space();
    			div = element("div");
    			table = element("table");
    			thead = element("thead");
    			tr = element("tr");
    			th0 = element("th");
    			th0.textContent = "From";
    			t3 = space();
    			th1 = element("th");
    			th1.textContent = "To";
    			t5 = space();
    			tbody = element("tbody");

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			attr_dev(h1, "id", "list-title");
    			attr_dev(h1, "class", "svelte-269k60");
    			add_location(h1, file$1, 60, 2, 1541);
    			attr_dev(th0, "class", "fw6 bb b--black-20 pb3 tl");
    			add_location(th0, file$1, 65, 10, 1687);
    			attr_dev(th1, "class", "fw6 bb b--black-20 pb3 tl");
    			add_location(th1, file$1, 66, 10, 1745);
    			add_location(tr, file$1, 64, 8, 1672);
    			add_location(thead, file$1, 63, 6, 1656);
    			attr_dev(tbody, "class", "lh-copy");
    			add_location(tbody, file$1, 69, 6, 1826);
    			attr_dev(table, "class", "w-100");
    			attr_dev(table, "cellspacing", "0");
    			add_location(table, file$1, 62, 4, 1612);
    			attr_dev(div, "id", "list-data");
    			add_location(div, file$1, 61, 2, 1587);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, h1, anchor);
    			append_dev(h1, t0);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, div, anchor);
    			append_dev(div, table);
    			append_dev(table, thead);
    			append_dev(thead, tr);
    			append_dev(tr, th0);
    			append_dev(tr, t3);
    			append_dev(tr, th1);
    			append_dev(table, t5);
    			append_dev(table, tbody);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(tbody, null);
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*aList*/ 1 && t0_value !== (t0_value = /*aList*/ ctx[0].info.title + "")) set_data_dev(t0, t0_value);

    			if (dirty & /*aList*/ 1) {
    				each_value = /*aList*/ ctx[0].data;
    				validate_each_argument(each_value);
    				let i;

    				for (i = 0; i < each_value.length; i += 1) {
    					const child_ctx = get_each_context(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(tbody, null);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value.length;
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(h1);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(div);
    			destroy_each(each_blocks, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_2.name,
    		type: "if",
    		source: "(60:0) {#if show == 'overview'}",
    		ctx
    	});

    	return block;
    }

    // (71:8) {#each aList.data as item, index}
    function create_each_block(ctx) {
    	let tr;
    	let td0;
    	let t0_value = /*item*/ ctx[10].from + "";
    	let t0;
    	let t1;
    	let td1;
    	let t2_value = /*item*/ ctx[10].to + "";
    	let t2;
    	let t3;
    	let tr_data_index_value;

    	const block = {
    		c: function create() {
    			tr = element("tr");
    			td0 = element("td");
    			t0 = text(t0_value);
    			t1 = space();
    			td1 = element("td");
    			t2 = text(t2_value);
    			t3 = space();
    			attr_dev(td0, "class", "pv3 pr3 bb b--black-20");
    			add_location(td0, file$1, 72, 12, 1938);
    			attr_dev(td1, "class", "pv3 pr3 bb b--black-20");
    			add_location(td1, file$1, 73, 12, 2002);
    			attr_dev(tr, "data-index", tr_data_index_value = /*index*/ ctx[12]);
    			add_location(tr, file$1, 71, 10, 1902);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, tr, anchor);
    			append_dev(tr, td0);
    			append_dev(td0, t0);
    			append_dev(tr, t1);
    			append_dev(tr, td1);
    			append_dev(td1, t2);
    			append_dev(tr, t3);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*aList*/ 1 && t0_value !== (t0_value = /*item*/ ctx[10].from + "")) set_data_dev(t0, t0_value);
    			if (dirty & /*aList*/ 1 && t2_value !== (t2_value = /*item*/ ctx[10].to + "")) set_data_dev(t2, t2_value);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(tr);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block.name,
    		type: "each",
    		source: "(71:8) {#each aList.data as item, index}",
    		ctx
    	});

    	return block;
    }

    // (82:0) {#if show == 'saved'}
    function create_if_block$1(ctx) {
    	let if_block_anchor;

    	function select_block_type(ctx, dirty) {
    		if (!loggedIn()) return create_if_block_1;
    		return create_else_block$1;
    	}

    	let current_block_type = select_block_type();
    	let if_block = current_block_type(ctx);

    	const block = {
    		c: function create() {
    			if_block.c();
    			if_block_anchor = empty();
    		},
    		m: function mount(target, anchor) {
    			if_block.m(target, anchor);
    			insert_dev(target, if_block_anchor, anchor);
    		},
    		p: function update(ctx, dirty) {
    			if_block.p(ctx, dirty);
    		},
    		d: function destroy(detaching) {
    			if_block.d(detaching);
    			if (detaching) detach_dev(if_block_anchor);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$1.name,
    		type: "if",
    		source: "(82:0) {#if show == 'saved'}",
    		ctx
    	});

    	return block;
    }

    // (89:2) {:else}
    function create_else_block$1(ctx) {
    	let p0;
    	let t1;
    	let p1;
    	let a;
    	let t2;

    	const block = {
    		c: function create() {
    			p0 = element("p");
    			p0.textContent = "List has been saved";
    			t1 = space();
    			p1 = element("p");
    			a = element("a");
    			t2 = text("Open in the browser");
    			add_location(p0, file$1, 89, 4, 2311);
    			attr_dev(a, "target", "_blank");
    			attr_dev(a, "href", /*listUrl*/ ctx[1]);
    			add_location(a, file$1, 91, 6, 2352);
    			add_location(p1, file$1, 90, 4, 2342);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, p0, anchor);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, p1, anchor);
    			append_dev(p1, a);
    			append_dev(a, t2);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*listUrl*/ 2) {
    				attr_dev(a, "href", /*listUrl*/ ctx[1]);
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(p0);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(p1);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_else_block$1.name,
    		type: "else",
    		source: "(89:2) {:else}",
    		ctx
    	});

    	return block;
    }

    // (83:2) {#if !loggedIn()}
    function create_if_block_1(ctx) {
    	let p;
    	let a;
    	let t;
    	let a_href_value;

    	const block = {
    		c: function create() {
    			p = element("p");
    			a = element("a");
    			t = text("Log into learnalist.net");
    			attr_dev(a, "target", "_blank");
    			attr_dev(a, "href", a_href_value = `${store.getServer()}/login.html`);
    			add_location(a, file$1, 84, 6, 2184);
    			add_location(p, file$1, 83, 4, 2174);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, p, anchor);
    			append_dev(p, a);
    			append_dev(a, t);
    		},
    		p: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(p);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1.name,
    		type: "if",
    		source: "(83:2) {#if !loggedIn()}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$2(ctx) {
    	let button0;
    	let t1;
    	let button1;
    	let t3;
    	let button2;
    	let t5;
    	let show_if = loggedIn();
    	let t6;
    	let t7;
    	let if_block2_anchor;
    	let mounted;
    	let dispose;
    	let if_block0 = show_if && create_if_block_3(ctx);
    	let if_block1 = /*show*/ ctx[2] == "overview" && create_if_block_2(ctx);
    	let if_block2 = /*show*/ ctx[2] == "saved" && create_if_block$1(ctx);

    	const block = {
    		c: function create() {
    			button0 = element("button");
    			button0.textContent = "Total Recall";
    			t1 = space();
    			button1 = element("button");
    			button1.textContent = "Slideshow";
    			t3 = space();
    			button2 = element("button");
    			button2.textContent = "Settings";
    			t5 = space();
    			if (if_block0) if_block0.c();
    			t6 = space();
    			if (if_block1) if_block1.c();
    			t7 = space();
    			if (if_block2) if_block2.c();
    			if_block2_anchor = empty();
    			attr_dev(button0, "class", "br3");
    			add_location(button0, file$1, 42, 0, 891);
    			attr_dev(button1, "class", "br3");
    			add_location(button1, file$1, 45, 0, 981);
    			attr_dev(button2, "class", "br3");
    			add_location(button2, file$1, 47, 0, 1062);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, button0, anchor);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, button1, anchor);
    			insert_dev(target, t3, anchor);
    			insert_dev(target, button2, anchor);
    			insert_dev(target, t5, anchor);
    			if (if_block0) if_block0.m(target, anchor);
    			insert_dev(target, t6, anchor);
    			if (if_block1) if_block1.m(target, anchor);
    			insert_dev(target, t7, anchor);
    			if (if_block2) if_block2.m(target, anchor);
    			insert_dev(target, if_block2_anchor, anchor);

    			if (!mounted) {
    				dispose = [
    					listen_dev(button0, "click", /*click_handler*/ ctx[4], false, false, false),
    					listen_dev(button1, "click", /*click_handler_1*/ ctx[5], false, false, false),
    					listen_dev(button2, "click", /*click_handler_2*/ ctx[6], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (show_if) if_block0.p(ctx, dirty);

    			if (/*show*/ ctx[2] == "overview") {
    				if (if_block1) {
    					if_block1.p(ctx, dirty);
    				} else {
    					if_block1 = create_if_block_2(ctx);
    					if_block1.c();
    					if_block1.m(t7.parentNode, t7);
    				}
    			} else if (if_block1) {
    				if_block1.d(1);
    				if_block1 = null;
    			}

    			if (/*show*/ ctx[2] == "saved") {
    				if (if_block2) {
    					if_block2.p(ctx, dirty);
    				} else {
    					if_block2 = create_if_block$1(ctx);
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
    			if (detaching) detach_dev(button0);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(button1);
    			if (detaching) detach_dev(t3);
    			if (detaching) detach_dev(button2);
    			if (detaching) detach_dev(t5);
    			if (if_block0) if_block0.d(detaching);
    			if (detaching) detach_dev(t6);
    			if (if_block1) if_block1.d(detaching);
    			if (detaching) detach_dev(t7);
    			if (if_block2) if_block2.d(detaching);
    			if (detaching) detach_dev(if_block2_anchor);
    			mounted = false;
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
    	let $store;
    	validate_store(store, "store");
    	component_subscribe($$self, store, $$value => $$invalidate(9, $store = $$value));
    	let aList = $store;
    	let listUrl;
    	let show = "overview";
    	let saved = false;

    	async function handleSave(event) {
    		if (saved) {
    			return;
    		}

    		try {
    			await store.save(aList);
    			$$invalidate(0, aList = $store);
    			$$invalidate(1, listUrl = `${store.getServer()}/alist/${aList.uuid}.html`);
    			$$invalidate(2, show = "saved");
    			saved = true;
    		} catch(e) {
    			console.log("e", e);
    			alert("Fail");
    		}
    	}

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$1.warn(`<Info> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Info", $$slots, []);
    	const click_handler = () => push("/play/total_recall");
    	const click_handler_1 = () => push("/play/slideshow");
    	const click_handler_2 = () => push("/settings");
    	const click_handler_3 = () => push("/interact/spaced_repetition/add");

    	$$self.$capture_state = () => ({
    		push,
    		LoginModal: Login_modal,
    		loggedIn,
    		notify,
    		store,
    		aList,
    		listUrl,
    		show,
    		saved,
    		handleSave,
    		$store
    	});

    	$$self.$inject_state = $$props => {
    		if ("aList" in $$props) $$invalidate(0, aList = $$props.aList);
    		if ("listUrl" in $$props) $$invalidate(1, listUrl = $$props.listUrl);
    		if ("show" in $$props) $$invalidate(2, show = $$props.show);
    		if ("saved" in $$props) saved = $$props.saved;
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [
    		aList,
    		listUrl,
    		show,
    		handleSave,
    		click_handler,
    		click_handler_1,
    		click_handler_2,
    		click_handler_3
    	];
    }

    class Info extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$2, create_fragment$2, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Info",
    			options,
    			id: create_fragment$2.name
    		});
    	}
    }

    /* src/browser-extension/import-play/start.svelte generated by Svelte v3.24.0 */

    const { Object: Object_1$1, console: console_1$2 } = globals;
    const file$2 = "src/browser-extension/import-play/start.svelte";

    // (117:0) {#if show == 'welcome'}
    function create_if_block$2(ctx) {
    	let button;
    	let t1;
    	let main;
    	let h1;
    	let t3;
    	let p;
    	let t5;
    	let ul;
    	let li0;
    	let t7;
    	let li1;
    	let t9;
    	let li2;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			button = element("button");
    			button.textContent = "Settings";
    			t1 = space();
    			main = element("main");
    			h1 = element("h1");
    			h1.textContent = "Welcome!";
    			t3 = space();
    			p = element("p");
    			p.textContent = "We will only try and load the list from";
    			t5 = space();
    			ul = element("ul");
    			li0 = element("li");
    			li0.textContent = "quizlet.com";
    			t7 = space();
    			li1 = element("li");
    			li1.textContent = "cram.com";
    			t9 = space();
    			li2 = element("li");
    			li2.textContent = "brainscape.com";
    			attr_dev(button, "class", "br3");
    			add_location(button, file$2, 117, 2, 2331);
    			add_location(h1, file$2, 119, 4, 2417);
    			add_location(p, file$2, 120, 4, 2439);
    			add_location(li0, file$2, 122, 6, 2514);
    			add_location(li1, file$2, 123, 6, 2541);
    			add_location(li2, file$2, 124, 6, 2565);
    			attr_dev(ul, "class", "list");
    			add_location(ul, file$2, 121, 4, 2490);
    			attr_dev(main, "class", "svelte-1ihbojs");
    			add_location(main, file$2, 118, 2, 2406);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, button, anchor);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, main, anchor);
    			append_dev(main, h1);
    			append_dev(main, t3);
    			append_dev(main, p);
    			append_dev(main, t5);
    			append_dev(main, ul);
    			append_dev(ul, li0);
    			append_dev(ul, t7);
    			append_dev(ul, li1);
    			append_dev(ul, t9);
    			append_dev(ul, li2);

    			if (!mounted) {
    				dispose = listen_dev(button, "click", /*click_handler*/ ctx[1], false, false, false);
    				mounted = true;
    			}
    		},
    		p: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(button);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(main);
    			mounted = false;
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$2.name,
    		type: "if",
    		source: "(117:0) {#if show == 'welcome'}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$3(ctx) {
    	let if_block_anchor;
    	let if_block = /*show*/ ctx[0] == "welcome" && create_if_block$2(ctx);

    	const block = {
    		c: function create() {
    			if (if_block) if_block.c();
    			if_block_anchor = empty();
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			if (if_block) if_block.m(target, anchor);
    			insert_dev(target, if_block_anchor, anchor);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*show*/ ctx[0] == "welcome") {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$2(ctx);
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

    function brainscapeToAlist(input) {
    	const data = input.detail;

    	return {
    		info: { title: data.title, type: "v2" },
    		data: data.listData
    	};
    }

    function cramToAlist(input) {
    	// setID = UUID
    	const data = input.detail;

    	const listData = data.listData.map(term => {
    		return {
    			from: term.front_plain,
    			to: term.back_plain
    		};
    	});

    	return {
    		info: { title: data.title, type: "v2" },
    		data: listData
    	};
    }

    function quizletToAlist(input) {
    	const data = input.detail;

    	const listData = Object.values(data.listData.setPageData.termIdToTermsMap).map(term => {
    		return { from: term.word, to: term.definition };
    	});

    	return {
    		info: { title: data.title, type: "v2" },
    		data: listData
    	};
    }

    function instance$3($$self, $$props, $$invalidate) {
    	let aList;
    	let show = "";

    	onMount(async () => {
    		handle();
    	});

    	function handle(event) {
    		chrome.runtime.onMessageExternal.addListener(function (request, sender, sendResponse) {
    			if (request.kind == "quizlet") {
    				aList = quizletToAlist(request);
    			}

    			if (request.kind == "cram") {
    				aList = cramToAlist(request);
    			}

    			if (request.kind == "brainscape") {
    				aList = brainscapeToAlist(request);
    			}

    			aList = aList;
    			document.querySelector("#play-data").innerHTML = JSON.stringify(aList);

    			if (aList) {
    				store.load(aList);
    				push("/overview");
    			}
    		});

    		chrome.tabs.query({ active: true, currentWindow: true }, function (tabs) {
    			const load = tabs[0].url.includes("cram.com") || tabs[0].url.includes("quizlet.com") || tabs[0].url.includes("brainscape.com");

    			if (!load) {
    				$$invalidate(0, show = "welcome");
    				return;
    			}

    			$$invalidate(0, show = "");
    			chrome.tabs.sendMessage(tabs[0].id, { greeting: "hello" });
    		});
    	}

    	const writable_props = [];

    	Object_1$1.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$2.warn(`<Start> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Start", $$slots, []);
    	const click_handler = () => push("/settings");

    	$$self.$capture_state = () => ({
    		push,
    		Info,
    		store,
    		onMount,
    		aList,
    		show,
    		brainscapeToAlist,
    		cramToAlist,
    		quizletToAlist,
    		handle
    	});

    	$$self.$inject_state = $$props => {
    		if ("aList" in $$props) aList = $$props.aList;
    		if ("show" in $$props) $$invalidate(0, show = $$props.show);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*show*/ 1) {
    			 console.log("show", show);
    		}
    	};

    	return [show, click_handler];
    }

    class Start extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$3, create_fragment$3, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Start",
    			options,
    			id: create_fragment$3.name
    		});
    	}
    }

    /* src/browser-extension/import-play/settings.svelte generated by Svelte v3.24.0 */

    const file$3 = "src/browser-extension/import-play/settings.svelte";

    function create_fragment$4(ctx) {
    	let button0;
    	let t1;
    	let div0;
    	let h10;
    	let t3;
    	let p0;
    	let t5;
    	let p1;
    	let input;
    	let t6;
    	let button1;
    	let t8;
    	let div1;
    	let h11;
    	let t10;
    	let p2;
    	let button2;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			button0 = element("button");
    			button0.textContent = "Close";
    			t1 = space();
    			div0 = element("div");
    			h10 = element("h1");
    			h10.textContent = "Change server";
    			t3 = space();
    			p0 = element("p");
    			p0.textContent = "You only need to change this if you are running your own learnalist server\n    or developing the chrome extension";
    			t5 = space();
    			p1 = element("p");
    			input = element("input");
    			t6 = space();
    			button1 = element("button");
    			button1.textContent = "Submit";
    			t8 = space();
    			div1 = element("div");
    			h11 = element("h1");
    			h11.textContent = "Reset to default settings";
    			t10 = space();
    			p2 = element("p");
    			button2 = element("button");
    			button2.textContent = "Reset";
    			attr_dev(button0, "class", "br3");
    			add_location(button0, file$3, 25, 0, 653);
    			add_location(h10, file$3, 28, 2, 729);
    			add_location(p0, file$3, 29, 2, 754);
    			add_location(input, file$3, 34, 4, 893);
    			add_location(p1, file$3, 33, 2, 885);
    			attr_dev(button1, "class", "br3");
    			add_location(button1, file$3, 36, 2, 933);
    			add_location(div0, file$3, 27, 0, 721);
    			add_location(h11, file$3, 40, 2, 1024);
    			attr_dev(button2, "class", "br3");
    			add_location(button2, file$3, 42, 4, 1069);
    			add_location(p2, file$3, 41, 2, 1061);
    			add_location(div1, file$3, 39, 0, 1016);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, button0, anchor);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, div0, anchor);
    			append_dev(div0, h10);
    			append_dev(div0, t3);
    			append_dev(div0, p0);
    			append_dev(div0, t5);
    			append_dev(div0, p1);
    			append_dev(p1, input);
    			set_input_value(input, /*baseUrl*/ ctx[0]);
    			append_dev(div0, t6);
    			append_dev(div0, button1);
    			insert_dev(target, t8, anchor);
    			insert_dev(target, div1, anchor);
    			append_dev(div1, h11);
    			append_dev(div1, t10);
    			append_dev(div1, p2);
    			append_dev(p2, button2);

    			if (!mounted) {
    				dispose = [
    					listen_dev(button0, "click", /*click_handler*/ ctx[3], false, false, false),
    					listen_dev(input, "input", /*input_input_handler*/ ctx[4]),
    					listen_dev(button1, "click", prevent_default(/*handleSubmit*/ ctx[1]), false, true, false),
    					listen_dev(button2, "click", prevent_default(/*handleReset*/ ctx[2]), false, true, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (dirty & /*baseUrl*/ 1 && input.value !== /*baseUrl*/ ctx[0]) {
    				set_input_value(input, /*baseUrl*/ ctx[0]);
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(button0);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(div0);
    			if (detaching) detach_dev(t8);
    			if (detaching) detach_dev(div1);
    			mounted = false;
    			run_all(dispose);
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
    	let baseUrl = getConfiguration(KeySettingsServer, "https://learnalist.net");

    	function handleSubmit() {
    		clearConfiguration();
    		saveConfiguration(KeySettingsServer, baseUrl);
    		chrome.runtime.sendMessage({ kind: "lookup-login-info" });
    	}

    	function handleReset() {
    		clearConfiguration();
    		$$invalidate(0, baseUrl = getConfiguration(KeySettingsServer, "https://learnalist.net"));
    		chrome.runtime.sendMessage({ kind: "lookup-login-info" });
    	}

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Settings> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Settings", $$slots, []);
    	const click_handler = () => push("/start");

    	function input_input_handler() {
    		baseUrl = this.value;
    		$$invalidate(0, baseUrl);
    	}

    	$$self.$capture_state = () => ({
    		push,
    		saveConfiguration,
    		getConfiguration,
    		clearConfiguration,
    		KeySettingsServer,
    		baseUrl,
    		handleSubmit,
    		handleReset
    	});

    	$$self.$inject_state = $$props => {
    		if ("baseUrl" in $$props) $$invalidate(0, baseUrl = $$props.baseUrl);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [baseUrl, handleSubmit, handleReset, click_handler, input_input_handler];
    }

    class Settings extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$4, create_fragment$4, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Settings",
    			options,
    			id: create_fragment$4.name
    		});
    	}
    }

    /* src/browser-extension/import-play/redirect.svelte generated by Svelte v3.24.0 */

    function create_fragment$5(ctx) {
    	const block = {
    		c: noop,
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
    		id: create_fragment$5.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$5($$self, $$props, $$invalidate) {
    	if (store.loaded()) {
    		push("/overview");
    	} else {
    		push("/start");
    	}

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Redirect> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Redirect", $$slots, []);
    	$$self.$capture_state = () => ({ push, store });
    	return [];
    }

    class Redirect extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$5, create_fragment$5, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Redirect",
    			options,
    			id: create_fragment$5.name
    		});
    	}
    }

    /* src/browser-extension/import-play/app.svelte generated by Svelte v3.24.0 */

    function create_fragment$6(ctx) {
    	let router;
    	let current;

    	router = new Router({
    			props: { routes: /*routes*/ ctx[0] },
    			$$inline: true
    		});

    	router.$on("conditionsFailed", /*conditionsFailed_handler*/ ctx[1]);

    	const block = {
    		c: function create() {
    			create_component(router.$$.fragment);
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
    		id: create_fragment$6.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$6($$self, $$props, $$invalidate) {
    	const routes = {
    		"/overview": Info,
    		"/start": Start,
    		"/settings": Settings,
    		"/": Redirect
    	};

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<App> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("App", $$slots, []);
    	const conditionsFailed_handler = event => replace("/");

    	$$self.$capture_state = () => ({
    		Router,
    		replace,
    		Start,
    		Overview: Info,
    		Settings,
    		Redirect,
    		routes
    	});

    	return [routes, conditionsFailed_handler];
    }

    class App extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$6, create_fragment$6, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "App",
    			options,
    			id: create_fragment$6.name
    		});
    	}
    }

    /* src/components/interact/routes/nothing.svelte generated by Svelte v3.24.0 */

    function create_fragment$7(ctx) {
    	const block = {
    		c: noop,
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
    		id: create_fragment$7.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$7($$self, $$props, $$invalidate) {
    	let { params } = $$props;
    	document.querySelector("#list-info").style.display = "";
    	document.querySelector("#play").style.display = "none";
    	const writable_props = ["params"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Nothing> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Nothing", $$slots, []);

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

    class Nothing extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$7, create_fragment$7, safe_not_equal, { params: 0 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Nothing",
    			options,
    			id: create_fragment$7.name
    		});

    		const { ctx } = this.$$;
    		const props = options.props || {};

    		if (/*params*/ ctx[0] === undefined && !("params" in props)) {
    			console.warn("<Nothing> was created without expected prop 'params'");
    		}
    	}

    	get params() {
    		throw new Error("<Nothing>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set params(value) {
    		throw new Error("<Nothing>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src/components/interact/total_recall/recall.svelte generated by Svelte v3.24.0 */

    const { console: console_1$3 } = globals;
    const file$4 = "src/components/interact/total_recall/recall.svelte";

    function get_each_context$1(ctx, list, i) {
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

    // (126:0) {#if state === 'playing'}
    function create_if_block_4(ctx) {
    	let p;
    	let t0;
    	let t1;
    	let t2;
    	let div;
    	let button0;
    	let t4;
    	let button1;
    	let t6;
    	let button2;
    	let mounted;
    	let dispose;
    	let if_block = /*hasChecked*/ ctx[3] && create_if_block_5(ctx);
    	let each_value_2 = /*playData*/ ctx[1];
    	validate_each_argument(each_value_2);
    	let each_blocks = [];

    	for (let i = 0; i < each_value_2.length; i += 1) {
    		each_blocks[i] = create_each_block_2(get_each_context_2(ctx, each_value_2, i));
    	}

    	const block = {
    		c: function create() {
    			p = element("p");
    			t0 = text("How many do you remember?\n    ");
    			if (if_block) if_block.c();
    			t1 = space();

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			t2 = space();
    			div = element("div");
    			button0 = element("button");
    			button0.textContent = "check";
    			t4 = space();
    			button1 = element("button");
    			button1.textContent = "I give up, show me";
    			t6 = space();
    			button2 = element("button");
    			button2.textContent = "restart";
    			add_location(p, file$4, 126, 2, 2523);
    			attr_dev(button0, "class", "br3");
    			add_location(button0, file$4, 142, 4, 2891);
    			attr_dev(button1, "class", "br3");
    			add_location(button1, file$4, 143, 4, 2947);
    			attr_dev(button2, "class", "br3");
    			add_location(button2, file$4, 144, 4, 3017);
    			attr_dev(div, "class", "pv1");
    			add_location(div, file$4, 141, 2, 2869);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, p, anchor);
    			append_dev(p, t0);
    			if (if_block) if_block.m(p, null);
    			insert_dev(target, t1, anchor);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(target, anchor);
    			}

    			insert_dev(target, t2, anchor);
    			insert_dev(target, div, anchor);
    			append_dev(div, button0);
    			append_dev(div, t4);
    			append_dev(div, button1);
    			append_dev(div, t6);
    			append_dev(div, button2);

    			if (!mounted) {
    				dispose = [
    					listen_dev(button0, "click", /*check*/ ctx[8], false, false, false),
    					listen_dev(button1, "click", /*showMe*/ ctx[11], false, false, false),
    					listen_dev(button2, "click", /*restart*/ ctx[10], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (/*hasChecked*/ ctx[3]) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block_5(ctx);
    					if_block.c();
    					if_block.m(p, null);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}

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
    						each_blocks[i].m(t2.parentNode, t2);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value_2.length;
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(p);
    			if (if_block) if_block.d();
    			if (detaching) detach_dev(t1);
    			destroy_each(each_blocks, detaching);
    			if (detaching) detach_dev(t2);
    			if (detaching) detach_dev(div);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_4.name,
    		type: "if",
    		source: "(126:0) {#if state === 'playing'}",
    		ctx
    	});

    	return block;
    }

    // (129:4) {#if hasChecked}
    function create_if_block_5(ctx) {
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
    		id: create_if_block_5.name,
    		type: "if",
    		source: "(129:4) {#if hasChecked}",
    		ctx
    	});

    	return block;
    }

    // (132:2) {#each playData as item, index}
    function create_each_block_2(ctx) {
    	let div;
    	let input;
    	let input_class_value;
    	let input_disabled_value;
    	let mounted;
    	let dispose;

    	function input_input_handler() {
    		/*input_input_handler*/ ctx[13].call(input, /*index*/ ctx[21]);
    	}

    	const block = {
    		c: function create() {
    			div = element("div");
    			input = element("input");
    			attr_dev(input, "class", input_class_value = "w-100 " + /*feedback*/ ctx[6][/*index*/ ctx[21]] + " svelte-l8tfzi");
    			input.disabled = input_disabled_value = /*feedback*/ ctx[6][/*index*/ ctx[21]] === "found";
    			attr_dev(input, "type", "text");
    			attr_dev(input, "placeholder", "");
    			add_location(input, file$4, 133, 6, 2670);
    			attr_dev(div, "class", "pv1");
    			add_location(div, file$4, 132, 4, 2646);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div, anchor);
    			append_dev(div, input);
    			set_input_value(input, /*guesses*/ ctx[2][/*index*/ ctx[21]]);

    			if (!mounted) {
    				dispose = listen_dev(input, "input", input_input_handler);
    				mounted = true;
    			}
    		},
    		p: function update(new_ctx, dirty) {
    			ctx = new_ctx;

    			if (dirty & /*feedback*/ 64 && input_class_value !== (input_class_value = "w-100 " + /*feedback*/ ctx[6][/*index*/ ctx[21]] + " svelte-l8tfzi")) {
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
    			mounted = false;
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block_2.name,
    		type: "each",
    		source: "(132:2) {#each playData as item, index}",
    		ctx
    	});

    	return block;
    }

    // (149:0) {#if state === 'finished'}
    function create_if_block_2$1(ctx) {
    	let p0;
    	let t1;
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
    	let mounted;
    	let dispose;
    	let each_value_1 = /*playData*/ ctx[1];
    	validate_each_argument(each_value_1);
    	let each_blocks = [];

    	for (let i = 0; i < each_value_1.length; i += 1) {
    		each_blocks[i] = create_each_block_1(get_each_context_1(ctx, each_value_1, i));
    	}

    	let if_block = /*perfect*/ ctx[5] && create_if_block_3$1(ctx);

    	const block = {
    		c: function create() {
    			p0 = element("p");
    			p0.textContent = "Well done! You did it.";
    			t1 = space();

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

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
    			add_location(p0, file$4, 149, 2, 3118);
    			add_location(p1, file$4, 165, 2, 3462);
    			attr_dev(button0, "class", "br3");
    			add_location(button0, file$4, 168, 4, 3525);
    			attr_dev(button1, "class", "br3");
    			add_location(button1, file$4, 169, 4, 3590);
    			attr_dev(div, "class", "pv1");
    			add_location(div, file$4, 167, 2, 3503);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, p0, anchor);
    			insert_dev(target, t1, anchor);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(target, anchor);
    			}

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

    			if (!mounted) {
    				dispose = [
    					listen_dev(button0, "click", /*playAgain*/ ctx[9], false, false, false),
    					listen_dev(button1, "click", /*restart*/ ctx[10], false, false, false)
    				];

    				mounted = true;
    			}
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
    						each_blocks[i].m(t2.parentNode, t2);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value_1.length;
    			}

    			if (/*perfect*/ ctx[5]) {
    				if (if_block) ; else {
    					if_block = create_if_block_3$1(ctx);
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
    			if (detaching) detach_dev(p0);
    			if (detaching) detach_dev(t1);
    			destroy_each(each_blocks, detaching);
    			if (detaching) detach_dev(t2);
    			if (if_block) if_block.d(detaching);
    			if (detaching) detach_dev(t3);
    			if (detaching) detach_dev(p1);
    			if (detaching) detach_dev(t7);
    			if (detaching) detach_dev(div);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_2$1.name,
    		type: "if",
    		source: "(149:0) {#if state === 'finished'}",
    		ctx
    	});

    	return block;
    }

    // (152:2) {#each playData as item, index}
    function create_each_block_1(ctx) {
    	let div;
    	let input;
    	let input_class_value;
    	let input_disabled_value;
    	let mounted;
    	let dispose;

    	function input_input_handler_1() {
    		/*input_input_handler_1*/ ctx[14].call(input, /*index*/ ctx[21]);
    	}

    	const block = {
    		c: function create() {
    			div = element("div");
    			input = element("input");
    			attr_dev(input, "class", input_class_value = "w-100 " + /*feedback*/ ctx[6][/*index*/ ctx[21]] + " svelte-l8tfzi");
    			input.disabled = input_disabled_value = /*feedback*/ ctx[6][/*index*/ ctx[21]] === "found";
    			attr_dev(input, "type", "text");
    			attr_dev(input, "placeholder", "");
    			add_location(input, file$4, 153, 6, 3211);
    			attr_dev(div, "class", "pv1");
    			add_location(div, file$4, 152, 4, 3187);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div, anchor);
    			append_dev(div, input);
    			set_input_value(input, /*guesses*/ ctx[2][/*index*/ ctx[21]]);

    			if (!mounted) {
    				dispose = listen_dev(input, "input", input_input_handler_1);
    				mounted = true;
    			}
    		},
    		p: function update(new_ctx, dirty) {
    			ctx = new_ctx;

    			if (dirty & /*feedback*/ 64 && input_class_value !== (input_class_value = "w-100 " + /*feedback*/ ctx[6][/*index*/ ctx[21]] + " svelte-l8tfzi")) {
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
    			mounted = false;
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block_1.name,
    		type: "each",
    		source: "(152:2) {#each playData as item, index}",
    		ctx
    	});

    	return block;
    }

    // (163:2) {#if perfect}
    function create_if_block_3$1(ctx) {
    	let p;

    	const block = {
    		c: function create() {
    			p = element("p");
    			p.textContent = "Perfect recall!";
    			add_location(p, file$4, 163, 4, 3429);
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
    		id: create_if_block_3$1.name,
    		type: "if",
    		source: "(163:2) {#if perfect}",
    		ctx
    	});

    	return block;
    }

    // (174:0) {#if state === 'show-me'}
    function create_if_block$3(ctx) {
    	let p;
    	let t0;
    	let t1;
    	let t2;
    	let div;
    	let button0;
    	let t4;
    	let button1;
    	let mounted;
    	let dispose;
    	let if_block = /*hasChecked*/ ctx[3] && create_if_block_1$1(ctx);
    	let each_value = /*playData*/ ctx[1];
    	validate_each_argument(each_value);
    	let each_blocks = [];

    	for (let i = 0; i < each_value.length; i += 1) {
    		each_blocks[i] = create_each_block$1(get_each_context$1(ctx, each_value, i));
    	}

    	const block = {
    		c: function create() {
    			p = element("p");
    			t0 = text("How many do you remember?\n    ");
    			if (if_block) if_block.c();
    			t1 = space();

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			t2 = space();
    			div = element("div");
    			button0 = element("button");
    			button0.textContent = "play again";
    			t4 = space();
    			button1 = element("button");
    			button1.textContent = "restart";
    			add_location(p, file$4, 174, 2, 3690);
    			attr_dev(button0, "class", "br3");
    			add_location(button0, file$4, 191, 4, 4009);
    			attr_dev(button1, "class", "br3");
    			add_location(button1, file$4, 192, 4, 4074);
    			attr_dev(div, "class", "pv2");
    			add_location(div, file$4, 190, 2, 3987);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, p, anchor);
    			append_dev(p, t0);
    			if (if_block) if_block.m(p, null);
    			insert_dev(target, t1, anchor);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(target, anchor);
    			}

    			insert_dev(target, t2, anchor);
    			insert_dev(target, div, anchor);
    			append_dev(div, button0);
    			append_dev(div, t4);
    			append_dev(div, button1);

    			if (!mounted) {
    				dispose = [
    					listen_dev(button0, "click", /*playAgain*/ ctx[9], false, false, false),
    					listen_dev(button1, "click", /*restart*/ ctx[10], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (/*hasChecked*/ ctx[3]) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block_1$1(ctx);
    					if_block.c();
    					if_block.m(p, null);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}

    			if (dirty & /*playData*/ 2) {
    				each_value = /*playData*/ ctx[1];
    				validate_each_argument(each_value);
    				let i;

    				for (i = 0; i < each_value.length; i += 1) {
    					const child_ctx = get_each_context$1(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block$1(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(t2.parentNode, t2);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value.length;
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(p);
    			if (if_block) if_block.d();
    			if (detaching) detach_dev(t1);
    			destroy_each(each_blocks, detaching);
    			if (detaching) detach_dev(t2);
    			if (detaching) detach_dev(div);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$3.name,
    		type: "if",
    		source: "(174:0) {#if state === 'show-me'}",
    		ctx
    	});

    	return block;
    }

    // (177:4) {#if hasChecked}
    function create_if_block_1$1(ctx) {
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
    		id: create_if_block_1$1.name,
    		type: "if",
    		source: "(177:4) {#if hasChecked}",
    		ctx
    	});

    	return block;
    }

    // (180:2) {#each playData as item, index}
    function create_each_block$1(ctx) {
    	let div;
    	let input;
    	let input_value_value;

    	const block = {
    		c: function create() {
    			div = element("div");
    			input = element("input");
    			attr_dev(input, "class", "w-100 found svelte-l8tfzi");
    			input.disabled = "true";
    			attr_dev(input, "type", "text");
    			attr_dev(input, "placeholder", "");
    			input.value = input_value_value = /*item*/ ctx[19];
    			add_location(input, file$4, 181, 6, 3837);
    			attr_dev(div, "class", "pv1");
    			add_location(div, file$4, 180, 4, 3813);
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
    		id: create_each_block$1.name,
    		type: "each",
    		source: "(180:2) {#each playData as item, index}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$8(ctx) {
    	let t0;
    	let t1;
    	let if_block2_anchor;
    	let if_block0 = /*state*/ ctx[0] === "playing" && create_if_block_4(ctx);
    	let if_block1 = /*state*/ ctx[0] === "finished" && create_if_block_2$1(ctx);
    	let if_block2 = /*state*/ ctx[0] === "show-me" && create_if_block$3(ctx);

    	const block = {
    		c: function create() {
    			if (if_block0) if_block0.c();
    			t0 = space();
    			if (if_block1) if_block1.c();
    			t1 = space();
    			if (if_block2) if_block2.c();
    			if_block2_anchor = empty();
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
    					if_block0 = create_if_block_4(ctx);
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
    					if_block1 = create_if_block_2$1(ctx);
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
    					if_block2 = create_if_block$3(ctx);
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
    		id: create_fragment$8.name,
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

    function instance$8($$self, $$props, $$invalidate) {
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
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$3.warn(`<Recall> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Recall", $$slots, []);

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
    		input_input_handler,
    		input_input_handler_1
    	];
    }

    class Recall extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$8, create_fragment$8, safe_not_equal, { data: 12 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Recall",
    			options,
    			id: create_fragment$8.name
    		});
    	}

    	get data() {
    		throw new Error("<Recall>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<Recall>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src/components/interact/total_recall/view.svelte generated by Svelte v3.24.0 */
    const file$5 = "src/components/interact/total_recall/view.svelte";

    function create_fragment$9(ctx) {
    	let blockquote;
    	let p;
    	let t;

    	const block = {
    		c: function create() {
    			blockquote = element("blockquote");
    			p = element("p");
    			t = text(/*show*/ ctx[0]);
    			attr_dev(p, "class", "f3 lh-copy");
    			add_location(p, file$5, 33, 2, 625);
    			attr_dev(blockquote, "class", "athelas ml0 mt4 pl4 black-90 bl bw2 b--black");
    			add_location(blockquote, file$5, 32, 0, 557);
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
    		id: create_fragment$9.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$9($$self, $$props, $$invalidate) {
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
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<View> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("View", $$slots, []);

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

    class View extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$9, create_fragment$9, safe_not_equal, { data: 1, speed: 2 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "View",
    			options,
    			id: create_fragment$9.name
    		});
    	}

    	get data() {
    		throw new Error("<View>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<View>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get speed() {
    		throw new Error("<View>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set speed(value) {
    		throw new Error("<View>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src/components/interact/total_recall/v2.svelte generated by Svelte v3.24.0 */
    const file$6 = "src/components/interact/total_recall/v2.svelte";

    // (72:4) {#if state === 'not-playing'}
    function create_if_block_2$2(ctx) {
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
    	let p5;
    	let span2;
    	let t15;
    	let p6;
    	let input2;
    	let input2_value_value;
    	let t16;
    	let t17;
    	let p7;
    	let input3;
    	let input3_value_value;
    	let t18;
    	let t19;
    	let pre;
    	let t20_value = JSON.stringify(/*data*/ ctx[2].slice(0, 2), "", 2) + "";
    	let t20;
    	let t21;
    	let button;
    	let mounted;
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
    			p5 = element("p");
    			span2 = element("span");
    			span2.textContent = "Which to show?";
    			t15 = space();
    			p6 = element("p");
    			input2 = element("input");
    			t16 = text("\n        from");
    			t17 = space();
    			p7 = element("p");
    			input3 = element("input");
    			t18 = text("\n        to");
    			t19 = space();
    			pre = element("pre");
    			t20 = text(t20_value);
    			t21 = space();
    			button = element("button");
    			button.textContent = "Are you ready to play?";
    			add_location(h1, file$6, 72, 6, 1431);
    			add_location(p0, file$6, 73, 6, 1452);
    			add_location(p1, file$6, 74, 6, 1497);
    			add_location(p2, file$6, 75, 6, 1557);
    			add_location(span0, file$6, 78, 8, 1620);
    			attr_dev(input0, "type", "number");
    			attr_dev(input0, "max", /*maxSize*/ ctx[6]);
    			attr_dev(input0, "min", "1");
    			add_location(input0, file$6, 79, 8, 1661);
    			add_location(p3, file$6, 77, 6, 1608);
    			add_location(span1, file$6, 83, 8, 1759);
    			attr_dev(input1, "type", "number");
    			attr_dev(input1, "max", input1_max_value = 5);
    			attr_dev(input1, "min", "1");
    			add_location(input1, file$6, 84, 8, 1800);
    			add_location(p4, file$6, 82, 6, 1747);
    			add_location(span2, file$6, 88, 8, 1889);
    			add_location(p5, file$6, 87, 6, 1877);
    			attr_dev(input2, "type", "radio");
    			input2.__value = input2_value_value = "from";
    			input2.value = input2.__value;
    			/*$$binding_groups*/ ctx[16][0].push(input2);
    			add_location(input2, file$6, 91, 8, 1946);
    			add_location(p6, file$6, 90, 6, 1934);
    			attr_dev(input3, "type", "radio");
    			input3.__value = input3_value_value = "to";
    			input3.value = input3.__value;
    			/*$$binding_groups*/ ctx[16][0].push(input3);
    			add_location(input3, file$6, 95, 8, 2047);
    			add_location(p7, file$6, 94, 6, 2035);
    			add_location(pre, file$6, 98, 6, 2132);
    			attr_dev(button, "class", "br3");
    			add_location(button, file$6, 99, 6, 2191);
    		},
    		m: function mount(target, anchor) {
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
    			insert_dev(target, p5, anchor);
    			append_dev(p5, span2);
    			insert_dev(target, t15, anchor);
    			insert_dev(target, p6, anchor);
    			append_dev(p6, input2);
    			input2.checked = input2.__value === /*showKey*/ ctx[3];
    			append_dev(p6, t16);
    			insert_dev(target, t17, anchor);
    			insert_dev(target, p7, anchor);
    			append_dev(p7, input3);
    			input3.checked = input3.__value === /*showKey*/ ctx[3];
    			append_dev(p7, t18);
    			insert_dev(target, t19, anchor);
    			insert_dev(target, pre, anchor);
    			append_dev(pre, t20);
    			insert_dev(target, t21, anchor);
    			insert_dev(target, button, anchor);

    			if (!mounted) {
    				dispose = [
    					listen_dev(input0, "input", /*input0_input_handler*/ ctx[13]),
    					listen_dev(input1, "input", /*input1_input_handler*/ ctx[14]),
    					listen_dev(input2, "change", /*input2_change_handler*/ ctx[15]),
    					listen_dev(input3, "change", /*input3_change_handler*/ ctx[17]),
    					listen_dev(button, "click", /*play*/ ctx[8], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*maxSize*/ 64) {
    				attr_dev(input0, "max", /*maxSize*/ ctx[6]);
    			}

    			if (dirty & /*gameSize*/ 1 && to_number(input0.value) !== /*gameSize*/ ctx[0]) {
    				set_input_value(input0, /*gameSize*/ ctx[0]);
    			}

    			if (dirty & /*speed*/ 2 && to_number(input1.value) !== /*speed*/ ctx[1]) {
    				set_input_value(input1, /*speed*/ ctx[1]);
    			}

    			if (dirty & /*showKey*/ 8) {
    				input2.checked = input2.__value === /*showKey*/ ctx[3];
    			}

    			if (dirty & /*showKey*/ 8) {
    				input3.checked = input3.__value === /*showKey*/ ctx[3];
    			}

    			if (dirty & /*data*/ 4 && t20_value !== (t20_value = JSON.stringify(/*data*/ ctx[2].slice(0, 2), "", 2) + "")) set_data_dev(t20, t20_value);
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
    			if (detaching) detach_dev(p5);
    			if (detaching) detach_dev(t15);
    			if (detaching) detach_dev(p6);
    			/*$$binding_groups*/ ctx[16][0].splice(/*$$binding_groups*/ ctx[16][0].indexOf(input2), 1);
    			if (detaching) detach_dev(t17);
    			if (detaching) detach_dev(p7);
    			/*$$binding_groups*/ ctx[16][0].splice(/*$$binding_groups*/ ctx[16][0].indexOf(input3), 1);
    			if (detaching) detach_dev(t19);
    			if (detaching) detach_dev(pre);
    			if (detaching) detach_dev(t21);
    			if (detaching) detach_dev(button);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_2$2.name,
    		type: "if",
    		source: "(72:4) {#if state === 'not-playing'}",
    		ctx
    	});

    	return block;
    }

    // (103:4) {#if state === 'playing'}
    function create_if_block_1$2(ctx) {
    	let view;
    	let current;

    	view = new View({
    			props: {
    				data: /*playData*/ ctx[4],
    				speed: /*speed*/ ctx[1] * 1000
    			},
    			$$inline: true
    		});

    	view.$on("finished", /*handleFinished*/ ctx[10]);

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
    			if (dirty & /*playData*/ 16) view_changes.data = /*playData*/ ctx[4];
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
    		source: "(103:4) {#if state === 'playing'}",
    		ctx
    	});

    	return block;
    }

    // (107:4) {#if state === 'recall'}
    function create_if_block$4(ctx) {
    	let recall;
    	let current;

    	recall = new Recall({
    			props: { data: /*playData*/ ctx[4] },
    			$$inline: true
    		});

    	recall.$on("finished", /*finished*/ ctx[9]);

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
    			if (dirty & /*playData*/ 16) recall_changes.data = /*playData*/ ctx[4];
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
    		id: create_if_block$4.name,
    		type: "if",
    		source: "(107:4) {#if state === 'recall'}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$a(ctx) {
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
    	let mounted;
    	let dispose;
    	let if_block0 = /*state*/ ctx[5] === "not-playing" && create_if_block_2$2(ctx);
    	let if_block1 = /*state*/ ctx[5] === "playing" && create_if_block_1$2(ctx);
    	let if_block2 = /*state*/ ctx[5] === "recall" && create_if_block$4(ctx);

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
    			attr_dev(h1, "class", "f2 measure");
    			add_location(h1, file$6, 66, 4, 1255);
    			attr_dev(button, "class", "br3");
    			add_location(button, file$6, 67, 4, 1300);
    			add_location(header, file$6, 65, 2, 1242);
    			attr_dev(div, "class", "pv2");
    			add_location(div, file$6, 70, 2, 1373);
    			add_location(article, file$6, 64, 0, 1230);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
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

    			if (!mounted) {
    				dispose = listen_dev(button, "click", /*handleClose*/ ctx[7], false, false, false);
    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*state*/ ctx[5] === "not-playing") {
    				if (if_block0) {
    					if_block0.p(ctx, dirty);
    				} else {
    					if_block0 = create_if_block_2$2(ctx);
    					if_block0.c();
    					if_block0.m(div, t4);
    				}
    			} else if (if_block0) {
    				if_block0.d(1);
    				if_block0 = null;
    			}

    			if (/*state*/ ctx[5] === "playing") {
    				if (if_block1) {
    					if_block1.p(ctx, dirty);

    					if (dirty & /*state*/ 32) {
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

    			if (/*state*/ ctx[5] === "recall") {
    				if (if_block2) {
    					if_block2.p(ctx, dirty);

    					if (dirty & /*state*/ 32) {
    						transition_in(if_block2, 1);
    					}
    				} else {
    					if_block2 = create_if_block$4(ctx);
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
    			mounted = false;
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$a.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$a($$self, $$props, $$invalidate) {
    	let { listElement } = $$props;
    	let { playElement } = $$props;
    	let { data = [] } = $$props;
    	let { gameSize = 3 } = $$props;
    	let { speed = 1 } = $$props;
    	let showKey = "from";
    	playElement.style.display = "";
    	listElement.style.display = "none";

    	function handleClose(event) {
    		$$invalidate(12, playElement.style.display = "none", playElement);
    		$$invalidate(11, listElement.style.display = "", listElement);
    		push("/");
    	}

    	let playData = [];

    	// This needs to pick the data
    	let state = "not-playing";

    	const shuffle = (arr, key) => arr.map(a => [Math.random(), a]).sort((a, b) => a[0] - b[0]).map(a => a[1][key]);

    	function play() {
    		// reduce to 7
    		// shuffle
    		let temp = shuffle(data, showKey);

    		$$invalidate(4, playData = temp.slice(0, gameSize));
    		$$invalidate(5, state = "playing");
    	}

    	function finished(event) {
    		if (event.detail.playAgain) {
    			play();
    			return;
    		}

    		$$invalidate(5, state = "not-playing");
    	}

    	function handleFinished() {
    		$$invalidate(5, state = "recall");
    	}

    	const writable_props = ["listElement", "playElement", "data", "gameSize", "speed"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<V2> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("V2", $$slots, []);
    	const $$binding_groups = [[]];

    	function input0_input_handler() {
    		gameSize = to_number(this.value);
    		$$invalidate(0, gameSize);
    	}

    	function input1_input_handler() {
    		speed = to_number(this.value);
    		$$invalidate(1, speed);
    	}

    	function input2_change_handler() {
    		showKey = this.__value;
    		$$invalidate(3, showKey);
    	}

    	function input3_change_handler() {
    		showKey = this.__value;
    		$$invalidate(3, showKey);
    	}

    	$$self.$set = $$props => {
    		if ("listElement" in $$props) $$invalidate(11, listElement = $$props.listElement);
    		if ("playElement" in $$props) $$invalidate(12, playElement = $$props.playElement);
    		if ("data" in $$props) $$invalidate(2, data = $$props.data);
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
    		showKey,
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
    		if ("listElement" in $$props) $$invalidate(11, listElement = $$props.listElement);
    		if ("playElement" in $$props) $$invalidate(12, playElement = $$props.playElement);
    		if ("data" in $$props) $$invalidate(2, data = $$props.data);
    		if ("gameSize" in $$props) $$invalidate(0, gameSize = $$props.gameSize);
    		if ("speed" in $$props) $$invalidate(1, speed = $$props.speed);
    		if ("showKey" in $$props) $$invalidate(3, showKey = $$props.showKey);
    		if ("playData" in $$props) $$invalidate(4, playData = $$props.playData);
    		if ("state" in $$props) $$invalidate(5, state = $$props.state);
    		if ("maxSize" in $$props) $$invalidate(6, maxSize = $$props.maxSize);
    	};

    	let maxSize;

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data*/ 4) {
    			 $$invalidate(6, maxSize = data.length);
    		}
    	};

    	return [
    		gameSize,
    		speed,
    		data,
    		showKey,
    		playData,
    		state,
    		maxSize,
    		handleClose,
    		play,
    		finished,
    		handleFinished,
    		listElement,
    		playElement,
    		input0_input_handler,
    		input1_input_handler,
    		input2_change_handler,
    		$$binding_groups,
    		input3_change_handler
    	];
    }

    class V2 extends SvelteComponentDev {
    	constructor(options) {
    		super(options);

    		init(this, options, instance$a, create_fragment$a, safe_not_equal, {
    			listElement: 11,
    			playElement: 12,
    			data: 2,
    			gameSize: 0,
    			speed: 1
    		});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "V2",
    			options,
    			id: create_fragment$a.name
    		});

    		const { ctx } = this.$$;
    		const props = options.props || {};

    		if (/*listElement*/ ctx[11] === undefined && !("listElement" in props)) {
    			console.warn("<V2> was created without expected prop 'listElement'");
    		}

    		if (/*playElement*/ ctx[12] === undefined && !("playElement" in props)) {
    			console.warn("<V2> was created without expected prop 'playElement'");
    		}
    	}

    	get listElement() {
    		return this.$$.ctx[11];
    	}

    	set listElement(listElement) {
    		this.$set({ listElement });
    		flush();
    	}

    	get playElement() {
    		return this.$$.ctx[12];
    	}

    	set playElement(playElement) {
    		this.$set({ playElement });
    		flush();
    	}

    	get data() {
    		return this.$$.ctx[2];
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

    /* src/components/interact/routes/total_recall_v2.svelte generated by Svelte v3.24.0 */

    function create_fragment$b(ctx) {
    	let totalrecall;
    	let current;

    	totalrecall = new V2({
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
    		id: create_fragment$b.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$b($$self, $$props, $$invalidate) {
    	let aList = JSON.parse(document.querySelector("#play-data").innerHTML);
    	let listElement = document.querySelector("#list-info");
    	let playElement = document.querySelector("#play");
    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Total_recall_v2> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Total_recall_v2", $$slots, []);

    	$$self.$capture_state = () => ({
    		TotalRecall: V2,
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

    class Total_recall_v2 extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$b, create_fragment$b, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Total_recall_v2",
    			options,
    			id: create_fragment$b.name
    		});
    	}
    }

    function add(node, event, handler) {
    	node.addEventListener(event, handler);
    	return () => node.removeEventListener(event, handler);
    }

    function dispatch_tap(node, x, y) {
    	node.dispatchEvent(new CustomEvent('tap', {
    		detail: { x, y }
    	}));
    }

    function handle_focus(event) {
    	const remove_keydown_handler = add(event.currentTarget, 'keydown', (event) => {
    		if (event.which === 32) dispatch_tap(event.currentTarget, null, null);
    	});

    	const remove_blur_handler = add(event.currentTarget, 'blur', (event) => {
    		remove_keydown_handler();
    		remove_blur_handler();
    	});
    }

    function is_button(node) {
    	return node.tagName === 'BUTTON' || node.type === 'button';
    }

    function tap_pointer(node) {
    	function handle_pointerdown(event) {
    		if ((node ).disabled) return;
    		const { clientX, clientY } = event;

    		const remove_pointerup_handler = add(node, 'pointerup', (event) => {
    			if (Math.abs(event.clientX - clientX) > 5) return;
    			if (Math.abs(event.clientY - clientY) > 5) return;

    			dispatch_tap(node, event.clientX, event.clientY);
    			remove_pointerup_handler();
    		});

    		setTimeout(remove_pointerup_handler, 300);
    	}

    	const remove_pointerdown_handler = add(node, 'pointerdown', handle_pointerdown);
    	const remove_focus_handler = is_button(node ) && add(node, 'focus', handle_focus);

    	return {
    		destroy() {
    			remove_pointerdown_handler();
    			remove_focus_handler && remove_focus_handler();
    		}
    	};
    }

    function tap_legacy(node) {
    	let mouse_enabled = true;
    	let mouse_timeout;

    	function handle_mousedown(event) {
    		const { clientX, clientY } = event;

    		const remove_mouseup_handler = add(node, 'mouseup', (event) => {
    			if (!mouse_enabled) return;
    			if (Math.abs(event.clientX - clientX) > 5) return;
    			if (Math.abs(event.clientY - clientY) > 5) return;

    			dispatch_tap(node, event.clientX, event.clientY);
    			remove_mouseup_handler();
    		});

    		clearTimeout(mouse_timeout);
    		setTimeout(remove_mouseup_handler, 300);
    	}

    	function handle_touchstart(event) {
    		if (event.changedTouches.length !== 1) return;
    		if ((node ).disabled) return;

    		const touch = event.changedTouches[0];
    		const { identifier, clientX, clientY } = touch;

    		const remove_touchend_handler = add(node, 'touchend', (event) => {
    			const touch = Array.from(event.changedTouches).find(t => t.identifier === identifier);
    			if (!touch) return;

    			if (Math.abs(touch.clientX - clientX) > 5) return;
    			if (Math.abs(touch.clientY - clientY) > 5) return;

    			dispatch_tap(node, touch.clientX, touch.clientY);

    			mouse_enabled = false;
    			mouse_timeout = setTimeout(() => {
    				mouse_enabled = true;
    			}, 350);
    		});

    		setTimeout(remove_touchend_handler, 300);
    	}

    	const remove_mousedown_handler = add(node, 'mousedown', handle_mousedown);
    	const remove_touchstart_handler = add(node, 'touchstart', handle_touchstart);
    	const remove_focus_handler = is_button(node ) && add(node, 'focus', handle_focus);

    	return {
    		destroy() {
    			remove_mousedown_handler();
    			remove_touchstart_handler();
    			remove_focus_handler && remove_focus_handler();
    		}
    	};
    }

    const tap = typeof PointerEvent === 'function'
    	? tap_pointer
    	: tap_legacy;

    /* src/components/interact/slideshow/v2.svelte generated by Svelte v3.24.0 */

    const { console: console_1$4 } = globals;
    const file$7 = "src/components/interact/slideshow/v2.svelte";

    // (113:4) {#if loops > 0}
    function create_if_block$5(ctx) {
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
    			add_location(cite, file$7, 113, 6, 2392);
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
    		id: create_if_block$5.name,
    		type: "if",
    		source: "(113:4) {#if loops > 0}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$c(ctx) {
    	let tap_action;
    	let article;
    	let header;
    	let h1;
    	let t1;
    	let button0;
    	let t3;
    	let button1;
    	let t5;
    	let blockquote;
    	let div;
    	let p0;
    	let t6_value = /*show*/ ctx[1].from + "";
    	let t6;
    	let t7;
    	let p1;
    	let t8_value = /*show*/ ctx[1].to + "";
    	let t8;
    	let t9;
    	let mounted;
    	let dispose;
    	let if_block = /*loops*/ ctx[0] > 0 && create_if_block$5(ctx);

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
    			div = element("div");
    			p0 = element("p");
    			t6 = text(t6_value);
    			t7 = space();
    			p1 = element("p");
    			t8 = text(t8_value);
    			t9 = space();
    			if (if_block) if_block.c();
    			attr_dev(h1, "class", "f2 measure");
    			add_location(h1, file$7, 103, 4, 2041);
    			attr_dev(button0, "class", "br3");
    			add_location(button0, file$7, 104, 4, 2083);
    			attr_dev(button1, "class", "br3");
    			add_location(button1, file$7, 105, 4, 2140);
    			add_location(header, file$7, 102, 2, 2028);
    			add_location(p0, file$7, 109, 6, 2313);
    			add_location(p1, file$7, 110, 6, 2338);
    			attr_dev(div, "class", "f3 lh-copy");
    			add_location(div, file$7, 108, 4, 2282);
    			attr_dev(blockquote, "class", "athelas ml0 mt4 pl4 black-90 bl bw2 b--black");
    			add_location(blockquote, file$7, 107, 2, 2212);
    			add_location(article, file$7, 101, 0, 2016);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, article, anchor);
    			append_dev(article, header);
    			append_dev(header, h1);
    			append_dev(header, t1);
    			append_dev(header, button0);
    			append_dev(header, t3);
    			append_dev(header, button1);
    			append_dev(article, t5);
    			append_dev(article, blockquote);
    			append_dev(blockquote, div);
    			append_dev(div, p0);
    			append_dev(p0, t6);
    			append_dev(div, t7);
    			append_dev(div, p1);
    			append_dev(p1, t8);
    			append_dev(blockquote, t9);
    			if (if_block) if_block.m(blockquote, null);

    			if (!mounted) {
    				dispose = [
    					action_destroyer(tap_action = tap.call(null, window)),
    					listen_dev(window, "keydown", /*handleKeydown*/ ctx[4], false, false, false),
    					listen_dev(window, "tap", /*tapHandler*/ ctx[5], false, false, false),
    					listen_dev(button0, "click", /*forward*/ ctx[2], false, false, false),
    					listen_dev(button1, "click", /*handleClose*/ ctx[3], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (dirty & /*show*/ 2 && t6_value !== (t6_value = /*show*/ ctx[1].from + "")) set_data_dev(t6, t6_value);
    			if (dirty & /*show*/ 2 && t8_value !== (t8_value = /*show*/ ctx[1].to + "")) set_data_dev(t8, t8_value);

    			if (/*loops*/ ctx[0] > 0) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$5(ctx);
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
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$c.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$c($$self, $$props, $$invalidate) {
    	let { listElement } = $$props;
    	let { playElement } = $$props;
    	let { aList } = $$props;
    	playElement.style.display = "";
    	listElement.style.display = "none";
    	let loops = 0;
    	let index = -1;

    	let firstTime = {
    		from: "Welcome, to beginning,",
    		to: "click next, or use the right arrow key.."
    	};

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
    		$$invalidate(7, playElement.style.display = "none", playElement);
    		$$invalidate(6, listElement.style.display = "", listElement);
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

    	function tapHandler(event) {
    		event.preventDefault();
    		const margin = 150;
    		const width = event.target.innerWidth; // window
    		const pageX = event.detail.x; // event.pageX when touchstart
    		const left = 0 + pageX < margin;
    		const right = width - margin < pageX;

    		if (left) {
    			backward();
    			return;
    		}

    		if (right) {
    			forward();
    			return;
    		}

    		return;
    	}

    	const writable_props = ["listElement", "playElement", "aList"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$4.warn(`<V2> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("V2", $$slots, []);

    	$$self.$set = $$props => {
    		if ("listElement" in $$props) $$invalidate(6, listElement = $$props.listElement);
    		if ("playElement" in $$props) $$invalidate(7, playElement = $$props.playElement);
    		if ("aList" in $$props) $$invalidate(8, aList = $$props.aList);
    	};

    	$$self.$capture_state = () => ({
    		push,
    		tap,
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
    		handleKeydown,
    		tapHandler
    	});

    	$$self.$inject_state = $$props => {
    		if ("listElement" in $$props) $$invalidate(6, listElement = $$props.listElement);
    		if ("playElement" in $$props) $$invalidate(7, playElement = $$props.playElement);
    		if ("aList" in $$props) $$invalidate(8, aList = $$props.aList);
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
    		tapHandler,
    		listElement,
    		playElement,
    		aList
    	];
    }

    class V2$1 extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$c, create_fragment$c, safe_not_equal, { listElement: 6, playElement: 7, aList: 8 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "V2",
    			options,
    			id: create_fragment$c.name
    		});

    		const { ctx } = this.$$;
    		const props = options.props || {};

    		if (/*listElement*/ ctx[6] === undefined && !("listElement" in props)) {
    			console_1$4.warn("<V2> was created without expected prop 'listElement'");
    		}

    		if (/*playElement*/ ctx[7] === undefined && !("playElement" in props)) {
    			console_1$4.warn("<V2> was created without expected prop 'playElement'");
    		}

    		if (/*aList*/ ctx[8] === undefined && !("aList" in props)) {
    			console_1$4.warn("<V2> was created without expected prop 'aList'");
    		}
    	}

    	get listElement() {
    		return this.$$.ctx[6];
    	}

    	set listElement(listElement) {
    		this.$set({ listElement });
    		flush();
    	}

    	get playElement() {
    		return this.$$.ctx[7];
    	}

    	set playElement(playElement) {
    		this.$set({ playElement });
    		flush();
    	}

    	get aList() {
    		return this.$$.ctx[8];
    	}

    	set aList(aList) {
    		this.$set({ aList });
    		flush();
    	}
    }

    /* src/components/interact/routes/slideshow_v2.svelte generated by Svelte v3.24.0 */

    function create_fragment$d(ctx) {
    	let slideshow;
    	let current;

    	slideshow = new V2$1({
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
    		id: create_fragment$d.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$d($$self, $$props, $$invalidate) {
    	let aList = JSON.parse(document.querySelector("#play-data").innerHTML);
    	let listElement = document.querySelector("#list-info");
    	let playElement = document.querySelector("#play");
    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Slideshow_v2> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Slideshow_v2", $$slots, []);

    	$$self.$capture_state = () => ({
    		Slideshow: V2$1,
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

    class Slideshow_v2 extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$d, create_fragment$d, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Slideshow_v2",
    			options,
    			id: create_fragment$d.name
    		});
    	}
    }

    /* src/components/interact/spaced_repetition/spaced_repetition_modal.svelte generated by Svelte v3.24.0 */
    const file$8 = "src/components/interact/spaced_repetition/spaced_repetition_modal.svelte";

    // (50:0) {#if show}
    function create_if_block$6(ctx) {
    	let div0;
    	let t0;
    	let div1;
    	let current_block_type_index;
    	let if_block;
    	let t1;
    	let button;
    	let current;
    	let mounted;
    	let dispose;
    	const if_block_creators = [create_if_block_1$3, create_else_block$2];
    	const if_blocks = [];

    	function select_block_type(ctx, dirty) {
    		if (loggedIn()) return 0;
    		return 1;
    	}

    	current_block_type_index = select_block_type();
    	if_block = if_blocks[current_block_type_index] = if_block_creators[current_block_type_index](ctx);

    	const block = {
    		c: function create() {
    			div0 = element("div");
    			t0 = space();
    			div1 = element("div");
    			if_block.c();
    			t1 = space();
    			button = element("button");
    			button.textContent = "cancel";
    			attr_dev(div0, "class", "modal-background svelte-1lb3tzo");
    			add_location(div0, file$8, 50, 2, 949);
    			attr_dev(button, "class", "br3");
    			add_location(button, file$8, 65, 4, 1428);
    			attr_dev(div1, "class", "modal svelte-1lb3tzo");
    			attr_dev(div1, "role", "dialog");
    			attr_dev(div1, "aria-modal", "true");
    			add_location(div1, file$8, 52, 2, 1008);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div0, anchor);
    			insert_dev(target, t0, anchor);
    			insert_dev(target, div1, anchor);
    			if_blocks[current_block_type_index].m(div1, null);
    			append_dev(div1, t1);
    			append_dev(div1, button);
    			current = true;

    			if (!mounted) {
    				dispose = [
    					listen_dev(div0, "click", /*handleClose*/ ctx[3], false, false, false),
    					listen_dev(button, "click", /*handleClose*/ ctx[3], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, dirty) {
    			if_block.p(ctx, dirty);
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
    			if (detaching) detach_dev(div0);
    			if (detaching) detach_dev(t0);
    			if (detaching) detach_dev(div1);
    			if_blocks[current_block_type_index].d();
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$6.name,
    		type: "if",
    		source: "(50:0) {#if show}",
    		ctx
    	});

    	return block;
    }

    // (62:4) {:else}
    function create_else_block$2(ctx) {
    	let p;
    	let t1;
    	let button;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			p = element("p");
    			p.textContent = "You need to be logged in to use spaced repetition";
    			t1 = space();
    			button = element("button");
    			button.textContent = "Login";
    			add_location(p, file$8, 62, 6, 1293);
    			attr_dev(button, "class", "br3");
    			add_location(button, file$8, 63, 6, 1356);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, p, anchor);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, button, anchor);

    			if (!mounted) {
    				dispose = listen_dev(button, "click", handleLogin$1, false, false, false);
    				mounted = true;
    			}
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(p);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(button);
    			mounted = false;
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_else_block$2.name,
    		type: "else",
    		source: "(62:4) {:else}",
    		ctx
    	});

    	return block;
    }

    // (54:4) {#if loggedIn()}
    function create_if_block_1$3(ctx) {
    	let t;
    	let if_block1_anchor;
    	let current;
    	let if_block0 = /*state*/ ctx[1] === "edit" && create_if_block_3$2(ctx);
    	let if_block1 = /*state*/ ctx[1] === "feedback" && create_if_block_2$3(ctx);

    	const block = {
    		c: function create() {
    			if (if_block0) if_block0.c();
    			t = space();
    			if (if_block1) if_block1.c();
    			if_block1_anchor = empty();
    		},
    		m: function mount(target, anchor) {
    			if (if_block0) if_block0.m(target, anchor);
    			insert_dev(target, t, anchor);
    			if (if_block1) if_block1.m(target, anchor);
    			insert_dev(target, if_block1_anchor, anchor);
    			current = true;
    		},
    		p: function update(ctx, dirty) {
    			if (/*state*/ ctx[1] === "edit") {
    				if (if_block0) {
    					if_block0.p(ctx, dirty);

    					if (dirty & /*state*/ 2) {
    						transition_in(if_block0, 1);
    					}
    				} else {
    					if_block0 = create_if_block_3$2(ctx);
    					if_block0.c();
    					transition_in(if_block0, 1);
    					if_block0.m(t.parentNode, t);
    				}
    			} else if (if_block0) {
    				group_outros();

    				transition_out(if_block0, 1, 1, () => {
    					if_block0 = null;
    				});

    				check_outros();
    			}

    			if (/*state*/ ctx[1] === "feedback") {
    				if (if_block1) {
    					if_block1.p(ctx, dirty);

    					if (dirty & /*state*/ 2) {
    						transition_in(if_block1, 1);
    					}
    				} else {
    					if_block1 = create_if_block_2$3(ctx);
    					if_block1.c();
    					transition_in(if_block1, 1);
    					if_block1.m(if_block1_anchor.parentNode, if_block1_anchor);
    				}
    			} else if (if_block1) {
    				group_outros();

    				transition_out(if_block1, 1, 1, () => {
    					if_block1 = null;
    				});

    				check_outros();
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(if_block0);
    			transition_in(if_block1);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(if_block0);
    			transition_out(if_block1);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (if_block0) if_block0.d(detaching);
    			if (detaching) detach_dev(t);
    			if (if_block1) if_block1.d(detaching);
    			if (detaching) detach_dev(if_block1_anchor);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1$3.name,
    		type: "if",
    		source: "(54:4) {#if loggedIn()}",
    		ctx
    	});

    	return block;
    }

    // (55:6) {#if state === 'edit'}
    function create_if_block_3$2(ctx) {
    	let t0;
    	let button;
    	let current;
    	let mounted;
    	let dispose;
    	const default_slot_template = /*$$slots*/ ctx[5].default;
    	const default_slot = create_slot(default_slot_template, ctx, /*$$scope*/ ctx[4], null);

    	const block = {
    		c: function create() {
    			if (default_slot) default_slot.c();
    			t0 = space();
    			button = element("button");
    			button.textContent = "Add";
    			attr_dev(button, "class", "br3");
    			add_location(button, file$8, 56, 8, 1135);
    		},
    		m: function mount(target, anchor) {
    			if (default_slot) {
    				default_slot.m(target, anchor);
    			}

    			insert_dev(target, t0, anchor);
    			insert_dev(target, button, anchor);
    			current = true;

    			if (!mounted) {
    				dispose = listen_dev(button, "click", /*click_handler*/ ctx[6], false, false, false);
    				mounted = true;
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (default_slot) {
    				if (default_slot.p && dirty & /*$$scope*/ 16) {
    					update_slot(default_slot, default_slot_template, ctx, /*$$scope*/ ctx[4], dirty, null, null);
    				}
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(default_slot, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(default_slot, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (default_slot) default_slot.d(detaching);
    			if (detaching) detach_dev(t0);
    			if (detaching) detach_dev(button);
    			mounted = false;
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_3$2.name,
    		type: "if",
    		source: "(55:6) {#if state === 'edit'}",
    		ctx
    	});

    	return block;
    }

    // (59:6) {#if state === 'feedback'}
    function create_if_block_2$3(ctx) {
    	let current;
    	const default_slot_template = /*$$slots*/ ctx[5].default;
    	const default_slot = create_slot(default_slot_template, ctx, /*$$scope*/ ctx[4], null);

    	const block = {
    		c: function create() {
    			if (default_slot) default_slot.c();
    		},
    		m: function mount(target, anchor) {
    			if (default_slot) {
    				default_slot.m(target, anchor);
    			}

    			current = true;
    		},
    		p: function update(ctx, dirty) {
    			if (default_slot) {
    				if (default_slot.p && dirty & /*$$scope*/ 16) {
    					update_slot(default_slot, default_slot_template, ctx, /*$$scope*/ ctx[4], dirty, null, null);
    				}
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(default_slot, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(default_slot, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (default_slot) default_slot.d(detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_2$3.name,
    		type: "if",
    		source: "(59:6) {#if state === 'feedback'}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$e(ctx) {
    	let if_block_anchor;
    	let current;
    	let if_block = /*show*/ ctx[0] && create_if_block$6(ctx);

    	const block = {
    		c: function create() {
    			if (if_block) if_block.c();
    			if_block_anchor = empty();
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			if (if_block) if_block.m(target, anchor);
    			insert_dev(target, if_block_anchor, anchor);
    			current = true;
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*show*/ ctx[0]) {
    				if (if_block) {
    					if_block.p(ctx, dirty);

    					if (dirty & /*show*/ 1) {
    						transition_in(if_block, 1);
    					}
    				} else {
    					if_block = create_if_block$6(ctx);
    					if_block.c();
    					transition_in(if_block, 1);
    					if_block.m(if_block_anchor.parentNode, if_block_anchor);
    				}
    			} else if (if_block) {
    				group_outros();

    				transition_out(if_block, 1, 1, () => {
    					if_block = null;
    				});

    				check_outros();
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
    			if (if_block) if_block.d(detaching);
    			if (detaching) detach_dev(if_block_anchor);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$e.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function handleLogin$1() {
    	window.location = "/login.html";
    }

    function instance$e($$self, $$props, $$invalidate) {
    	const dispatch = createEventDispatcher();
    	const close = () => dispatch("close");
    	let { show } = $$props;
    	let { state } = $$props;

    	function handleClose() {
    		dispatch("close");
    	}

    	const writable_props = ["show", "state"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Spaced_repetition_modal> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Spaced_repetition_modal", $$slots, ['default']);
    	const click_handler = () => dispatch("add");

    	$$self.$set = $$props => {
    		if ("show" in $$props) $$invalidate(0, show = $$props.show);
    		if ("state" in $$props) $$invalidate(1, state = $$props.state);
    		if ("$$scope" in $$props) $$invalidate(4, $$scope = $$props.$$scope);
    	};

    	$$self.$capture_state = () => ({
    		loggedIn,
    		push,
    		createEventDispatcher,
    		dispatch,
    		close,
    		show,
    		state,
    		handleClose,
    		handleLogin: handleLogin$1
    	});

    	$$self.$inject_state = $$props => {
    		if ("show" in $$props) $$invalidate(0, show = $$props.show);
    		if ("state" in $$props) $$invalidate(1, state = $$props.state);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [show, state, dispatch, handleClose, $$scope, $$slots, click_handler];
    }

    class Spaced_repetition_modal extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$e, create_fragment$e, safe_not_equal, { show: 0, state: 1 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Spaced_repetition_modal",
    			options,
    			id: create_fragment$e.name
    		});

    		const { ctx } = this.$$;
    		const props = options.props || {};

    		if (/*show*/ ctx[0] === undefined && !("show" in props)) {
    			console.warn("<Spaced_repetition_modal> was created without expected prop 'show'");
    		}

    		if (/*state*/ ctx[1] === undefined && !("state" in props)) {
    			console.warn("<Spaced_repetition_modal> was created without expected prop 'state'");
    		}
    	}

    	get show() {
    		return this.$$.ctx[0];
    	}

    	set show(show) {
    		this.$set({ show });
    		flush();
    	}

    	get state() {
    		return this.$$.ctx[1];
    	}

    	set state(state) {
    		this.$set({ state });
    		flush();
    	}
    }

    async function addEntry(input) {
        return await addSpacedRepetitionEntry(input);
    }

    /* src/components/interact/spaced_repetition/add_v2.svelte generated by Svelte v3.24.0 */

    const { console: console_1$5 } = globals;
    const file$9 = "src/components/interact/spaced_repetition/add_v2.svelte";

    function get_each_context$2(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[14] = list[i];
    	child_ctx[16] = i;
    	return child_ctx;
    }

    // (97:6) {#each aList.data as item, index}
    function create_each_block$2(ctx) {
    	let tr;
    	let td0;
    	let t0_value = /*item*/ ctx[14].from + "";
    	let t0;
    	let t1;
    	let td1;
    	let t2_value = /*item*/ ctx[14].to + "";
    	let t2;
    	let t3;
    	let tr_data_index_value;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			tr = element("tr");
    			td0 = element("td");
    			t0 = text(t0_value);
    			t1 = space();
    			td1 = element("td");
    			t2 = text(t2_value);
    			t3 = space();
    			attr_dev(td0, "class", "pv3 pr3 bb b--black-20");
    			add_location(td0, file$9, 98, 10, 2090);
    			attr_dev(td1, "class", "pv3 pr3 bb b--black-20");
    			add_location(td1, file$9, 99, 10, 2152);
    			attr_dev(tr, "data-index", tr_data_index_value = /*index*/ ctx[16]);
    			add_location(tr, file$9, 97, 8, 2040);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, tr, anchor);
    			append_dev(tr, td0);
    			append_dev(td0, t0);
    			append_dev(tr, t1);
    			append_dev(tr, td1);
    			append_dev(td1, t2);
    			append_dev(tr, t3);

    			if (!mounted) {
    				dispose = listen_dev(tr, "click", /*edit*/ ctx[6], false, false, false);
    				mounted = true;
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*aList*/ 1 && t0_value !== (t0_value = /*item*/ ctx[14].from + "")) set_data_dev(t0, t0_value);
    			if (dirty & /*aList*/ 1 && t2_value !== (t2_value = /*item*/ ctx[14].to + "")) set_data_dev(t2, t2_value);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(tr);
    			mounted = false;
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block$2.name,
    		type: "each",
    		source: "(97:6) {#each aList.data as item, index}",
    		ctx
    	});

    	return block;
    }

    // (108:2) {#if state === 'edit'}
    function create_if_block_1$4(ctx) {
    	let p0;
    	let span;
    	let t1;
    	let p1;
    	let input0;
    	let input0_value_value;
    	let t2;
    	let t3;
    	let p2;
    	let input1;
    	let input1_value_value;
    	let t4;
    	let t5;
    	let pre;
    	let t6_value = JSON.stringify(/*data*/ ctx[1], "", 2) + "";
    	let t6;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			p0 = element("p");
    			span = element("span");
    			span.textContent = "Which to show?";
    			t1 = space();
    			p1 = element("p");
    			input0 = element("input");
    			t2 = text("\n      from");
    			t3 = space();
    			p2 = element("p");
    			input1 = element("input");
    			t4 = text("\n      to");
    			t5 = space();
    			pre = element("pre");
    			t6 = text(t6_value);
    			add_location(span, file$9, 109, 6, 2354);
    			add_location(p0, file$9, 108, 4, 2344);
    			attr_dev(input0, "type", "radio");
    			input0.__value = input0_value_value = "from";
    			input0.value = input0.__value;
    			/*$$binding_groups*/ ctx[12][0].push(input0);
    			add_location(input0, file$9, 112, 6, 2405);
    			add_location(p1, file$9, 111, 4, 2395);
    			attr_dev(input1, "type", "radio");
    			input1.__value = input1_value_value = "to";
    			input1.value = input1.__value;
    			/*$$binding_groups*/ ctx[12][0].push(input1);
    			add_location(input1, file$9, 116, 6, 2498);
    			add_location(p2, file$9, 115, 4, 2488);
    			add_location(pre, file$9, 119, 4, 2577);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, p0, anchor);
    			append_dev(p0, span);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, p1, anchor);
    			append_dev(p1, input0);
    			input0.checked = input0.__value === /*showKey*/ ctx[3];
    			append_dev(p1, t2);
    			insert_dev(target, t3, anchor);
    			insert_dev(target, p2, anchor);
    			append_dev(p2, input1);
    			input1.checked = input1.__value === /*showKey*/ ctx[3];
    			append_dev(p2, t4);
    			insert_dev(target, t5, anchor);
    			insert_dev(target, pre, anchor);
    			append_dev(pre, t6);

    			if (!mounted) {
    				dispose = [
    					listen_dev(input0, "change", /*input0_change_handler*/ ctx[11]),
    					listen_dev(input1, "change", /*input1_change_handler*/ ctx[13])
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*showKey*/ 8) {
    				input0.checked = input0.__value === /*showKey*/ ctx[3];
    			}

    			if (dirty & /*showKey*/ 8) {
    				input1.checked = input1.__value === /*showKey*/ ctx[3];
    			}

    			if (dirty & /*data*/ 2 && t6_value !== (t6_value = JSON.stringify(/*data*/ ctx[1], "", 2) + "")) set_data_dev(t6, t6_value);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(p0);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(p1);
    			/*$$binding_groups*/ ctx[12][0].splice(/*$$binding_groups*/ ctx[12][0].indexOf(input0), 1);
    			if (detaching) detach_dev(t3);
    			if (detaching) detach_dev(p2);
    			/*$$binding_groups*/ ctx[12][0].splice(/*$$binding_groups*/ ctx[12][0].indexOf(input1), 1);
    			if (detaching) detach_dev(t5);
    			if (detaching) detach_dev(pre);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1$4.name,
    		type: "if",
    		source: "(108:2) {#if state === 'edit'}",
    		ctx
    	});

    	return block;
    }

    // (123:2) {#if state === 'feedback'}
    function create_if_block$7(ctx) {
    	let p0;
    	let t1;
    	let p1;
    	let t2;
    	let t3_value = /*data*/ ctx[1].settings.when_next + "";
    	let t3;

    	const block = {
    		c: function create() {
    			p0 = element("p");
    			p0.textContent = "Already in the system";
    			t1 = space();
    			p1 = element("p");
    			t2 = text("You will be reminded on ");
    			t3 = text(t3_value);
    			add_location(p0, file$9, 123, 4, 2660);
    			add_location(p1, file$9, 124, 4, 2693);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, p0, anchor);
    			insert_dev(target, t1, anchor);
    			insert_dev(target, p1, anchor);
    			append_dev(p1, t2);
    			append_dev(p1, t3);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*data*/ 2 && t3_value !== (t3_value = /*data*/ ctx[1].settings.when_next + "")) set_data_dev(t3, t3_value);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(p0);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(p1);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$7.name,
    		type: "if",
    		source: "(123:2) {#if state === 'feedback'}",
    		ctx
    	});

    	return block;
    }

    // (107:0) <Modal {show} {state} on:add={add} on:close={close}>
    function create_default_slot(ctx) {
    	let t;
    	let if_block1_anchor;
    	let if_block0 = /*state*/ ctx[2] === "edit" && create_if_block_1$4(ctx);
    	let if_block1 = /*state*/ ctx[2] === "feedback" && create_if_block$7(ctx);

    	const block = {
    		c: function create() {
    			if (if_block0) if_block0.c();
    			t = space();
    			if (if_block1) if_block1.c();
    			if_block1_anchor = empty();
    		},
    		m: function mount(target, anchor) {
    			if (if_block0) if_block0.m(target, anchor);
    			insert_dev(target, t, anchor);
    			if (if_block1) if_block1.m(target, anchor);
    			insert_dev(target, if_block1_anchor, anchor);
    		},
    		p: function update(ctx, dirty) {
    			if (/*state*/ ctx[2] === "edit") {
    				if (if_block0) {
    					if_block0.p(ctx, dirty);
    				} else {
    					if_block0 = create_if_block_1$4(ctx);
    					if_block0.c();
    					if_block0.m(t.parentNode, t);
    				}
    			} else if (if_block0) {
    				if_block0.d(1);
    				if_block0 = null;
    			}

    			if (/*state*/ ctx[2] === "feedback") {
    				if (if_block1) {
    					if_block1.p(ctx, dirty);
    				} else {
    					if_block1 = create_if_block$7(ctx);
    					if_block1.c();
    					if_block1.m(if_block1_anchor.parentNode, if_block1_anchor);
    				}
    			} else if (if_block1) {
    				if_block1.d(1);
    				if_block1 = null;
    			}
    		},
    		d: function destroy(detaching) {
    			if (if_block0) if_block0.d(detaching);
    			if (detaching) detach_dev(t);
    			if (if_block1) if_block1.d(detaching);
    			if (detaching) detach_dev(if_block1_anchor);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_default_slot.name,
    		type: "slot",
    		source: "(107:0) <Modal {show} {state} on:add={add} on:close={close}>",
    		ctx
    	});

    	return block;
    }

    function create_fragment$f(ctx) {
    	let header;
    	let button;
    	let t1;
    	let h1;
    	let t3;
    	let h3;
    	let t5;
    	let div;
    	let table;
    	let thead;
    	let tr;
    	let th0;
    	let t7;
    	let th1;
    	let t9;
    	let tbody;
    	let t10;
    	let modal;
    	let current;
    	let mounted;
    	let dispose;
    	let each_value = /*aList*/ ctx[0].data;
    	validate_each_argument(each_value);
    	let each_blocks = [];

    	for (let i = 0; i < each_value.length; i += 1) {
    		each_blocks[i] = create_each_block$2(get_each_context$2(ctx, each_value, i));
    	}

    	modal = new Spaced_repetition_modal({
    			props: {
    				show: /*show*/ ctx[4],
    				state: /*state*/ ctx[2],
    				$$slots: { default: [create_default_slot] },
    				$$scope: { ctx }
    			},
    			$$inline: true
    		});

    	modal.$on("add", /*add*/ ctx[8]);
    	modal.$on("close", /*close*/ ctx[7]);

    	const block = {
    		c: function create() {
    			header = element("header");
    			button = element("button");
    			button.textContent = "Close";
    			t1 = space();
    			h1 = element("h1");
    			h1.textContent = "🧠 + 💪";
    			t3 = space();
    			h3 = element("h3");
    			h3.textContent = "Click on the row you want to add";
    			t5 = space();
    			div = element("div");
    			table = element("table");
    			thead = element("thead");
    			tr = element("tr");
    			th0 = element("th");
    			th0.textContent = "From";
    			t7 = space();
    			th1 = element("th");
    			th1.textContent = "To";
    			t9 = space();
    			tbody = element("tbody");

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			t10 = space();
    			create_component(modal.$$.fragment);
    			attr_dev(button, "class", "br3");
    			add_location(button, file$9, 82, 2, 1568);
    			attr_dev(h1, "class", "f2 measure");
    			attr_dev(h1, "title", "Spaced Repetition");
    			add_location(h1, file$9, 83, 2, 1628);
    			add_location(h3, file$9, 84, 2, 1692);
    			add_location(header, file$9, 81, 0, 1557);
    			attr_dev(th0, "class", "fw6 bb b--black-20 pb3 tl");
    			add_location(th0, file$9, 91, 8, 1837);
    			attr_dev(th1, "class", "fw6 bb b--black-20 pb3 tl");
    			add_location(th1, file$9, 92, 8, 1893);
    			add_location(tr, file$9, 90, 6, 1824);
    			add_location(thead, file$9, 89, 4, 1810);
    			attr_dev(tbody, "class", "lh-copy");
    			add_location(tbody, file$9, 95, 4, 1968);
    			attr_dev(table, "class", "w-100");
    			attr_dev(table, "cellspacing", "0");
    			add_location(table, file$9, 88, 2, 1768);
    			attr_dev(div, "id", "list-data");
    			add_location(div, file$9, 87, 0, 1745);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, header, anchor);
    			append_dev(header, button);
    			append_dev(header, t1);
    			append_dev(header, h1);
    			append_dev(header, t3);
    			append_dev(header, h3);
    			insert_dev(target, t5, anchor);
    			insert_dev(target, div, anchor);
    			append_dev(div, table);
    			append_dev(table, thead);
    			append_dev(thead, tr);
    			append_dev(tr, th0);
    			append_dev(tr, t7);
    			append_dev(tr, th1);
    			append_dev(table, t9);
    			append_dev(table, tbody);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(tbody, null);
    			}

    			insert_dev(target, t10, anchor);
    			mount_component(modal, target, anchor);
    			current = true;

    			if (!mounted) {
    				dispose = listen_dev(button, "click", /*handleClose*/ ctx[5], false, false, false);
    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (dirty & /*edit, aList*/ 65) {
    				each_value = /*aList*/ ctx[0].data;
    				validate_each_argument(each_value);
    				let i;

    				for (i = 0; i < each_value.length; i += 1) {
    					const child_ctx = get_each_context$2(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block$2(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(tbody, null);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value.length;
    			}

    			const modal_changes = {};
    			if (dirty & /*show*/ 16) modal_changes.show = /*show*/ ctx[4];
    			if (dirty & /*state*/ 4) modal_changes.state = /*state*/ ctx[2];

    			if (dirty & /*$$scope, data, state, showKey*/ 131086) {
    				modal_changes.$$scope = { dirty, ctx };
    			}

    			modal.$set(modal_changes);
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(modal.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(modal.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(header);
    			if (detaching) detach_dev(t5);
    			if (detaching) detach_dev(div);
    			destroy_each(each_blocks, detaching);
    			if (detaching) detach_dev(t10);
    			destroy_component(modal, detaching);
    			mounted = false;
    			dispose();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$f.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$f($$self, $$props, $$invalidate) {
    	let { playElement } = $$props;
    	let { listElement } = $$props;
    	let { aList } = $$props;
    	playElement.style.display = "";
    	listElement.style.display = "none";
    	let data;
    	let state = "edit";
    	let showKey = "from";
    	let show = false;

    	function handleClose(event) {
    		$$invalidate(9, playElement.style.display = "none", playElement);
    		push("/");
    	}

    	function edit(event) {
    		// How did this work before?
    		const index = event.target.closest("[data-index]").getAttribute("data-index");

    		if (!index) {
    			return;
    		}

    		$$invalidate(1, data = aList.data[index]);
    		$$invalidate(4, show = true);
    	}

    	function close() {
    		$$invalidate(1, data = null);
    		$$invalidate(2, state = "edit");
    		$$invalidate(3, showKey = "from");
    		$$invalidate(4, show = false);
    	}

    	async function add(event) {
    		const input = {
    			show: data[showKey],
    			data,
    			settings: { show: showKey },
    			kind: aList.info.type
    		};

    		const response = await addEntry(input);

    		switch (response.status) {
    			case 201:
    				close();
    				break;
    			case 200:
    				$$invalidate(2, state = "feedback");
    				$$invalidate(1, data = response.body);
    				break;
    			default:
    				console.log("failed to add for spaced learning");
    				console.log(response);
    				break;
    		}
    	}

    	const writable_props = ["playElement", "listElement", "aList"];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$5.warn(`<Add_v2> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Add_v2", $$slots, []);
    	const $$binding_groups = [[]];

    	function input0_change_handler() {
    		showKey = this.__value;
    		$$invalidate(3, showKey);
    	}

    	function input1_change_handler() {
    		showKey = this.__value;
    		$$invalidate(3, showKey);
    	}

    	$$self.$set = $$props => {
    		if ("playElement" in $$props) $$invalidate(9, playElement = $$props.playElement);
    		if ("listElement" in $$props) $$invalidate(10, listElement = $$props.listElement);
    		if ("aList" in $$props) $$invalidate(0, aList = $$props.aList);
    	};

    	$$self.$capture_state = () => ({
    		Modal: Spaced_repetition_modal,
    		addEntry,
    		push,
    		tap,
    		playElement,
    		listElement,
    		aList,
    		data,
    		state,
    		showKey,
    		show,
    		handleClose,
    		edit,
    		close,
    		add
    	});

    	$$self.$inject_state = $$props => {
    		if ("playElement" in $$props) $$invalidate(9, playElement = $$props.playElement);
    		if ("listElement" in $$props) $$invalidate(10, listElement = $$props.listElement);
    		if ("aList" in $$props) $$invalidate(0, aList = $$props.aList);
    		if ("data" in $$props) $$invalidate(1, data = $$props.data);
    		if ("state" in $$props) $$invalidate(2, state = $$props.state);
    		if ("showKey" in $$props) $$invalidate(3, showKey = $$props.showKey);
    		if ("show" in $$props) $$invalidate(4, show = $$props.show);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [
    		aList,
    		data,
    		state,
    		showKey,
    		show,
    		handleClose,
    		edit,
    		close,
    		add,
    		playElement,
    		listElement,
    		input0_change_handler,
    		$$binding_groups,
    		input1_change_handler
    	];
    }

    class Add_v2 extends SvelteComponentDev {
    	constructor(options) {
    		super(options);

    		init(this, options, instance$f, create_fragment$f, safe_not_equal, {
    			playElement: 9,
    			listElement: 10,
    			aList: 0
    		});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Add_v2",
    			options,
    			id: create_fragment$f.name
    		});

    		const { ctx } = this.$$;
    		const props = options.props || {};

    		if (/*playElement*/ ctx[9] === undefined && !("playElement" in props)) {
    			console_1$5.warn("<Add_v2> was created without expected prop 'playElement'");
    		}

    		if (/*listElement*/ ctx[10] === undefined && !("listElement" in props)) {
    			console_1$5.warn("<Add_v2> was created without expected prop 'listElement'");
    		}

    		if (/*aList*/ ctx[0] === undefined && !("aList" in props)) {
    			console_1$5.warn("<Add_v2> was created without expected prop 'aList'");
    		}
    	}

    	get playElement() {
    		return this.$$.ctx[9];
    	}

    	set playElement(playElement) {
    		this.$set({ playElement });
    		flush();
    	}

    	get listElement() {
    		return this.$$.ctx[10];
    	}

    	set listElement(listElement) {
    		this.$set({ listElement });
    		flush();
    	}

    	get aList() {
    		return this.$$.ctx[0];
    	}

    	set aList(aList) {
    		this.$set({ aList });
    		flush();
    	}
    }

    /* src/components/interact/routes/spaced_repetition_v2.svelte generated by Svelte v3.24.0 */

    function create_fragment$g(ctx) {
    	let spacedrepetitionadd;
    	let current;

    	spacedrepetitionadd = new Add_v2({
    			props: {
    				aList: /*aList*/ ctx[0],
    				listElement: /*listElement*/ ctx[1],
    				playElement: /*playElement*/ ctx[2]
    			},
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			create_component(spacedrepetitionadd.$$.fragment);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			mount_component(spacedrepetitionadd, target, anchor);
    			current = true;
    		},
    		p: noop,
    		i: function intro(local) {
    			if (current) return;
    			transition_in(spacedrepetitionadd.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(spacedrepetitionadd.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			destroy_component(spacedrepetitionadd, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$g.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$g($$self, $$props, $$invalidate) {
    	let aList = JSON.parse(document.querySelector("#play-data").innerHTML);
    	let listElement = document.querySelector("#list-info");
    	let playElement = document.querySelector("#play");
    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Spaced_repetition_v2> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Spaced_repetition_v2", $$slots, []);

    	$$self.$capture_state = () => ({
    		SpacedRepetitionAdd: Add_v2,
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

    class Spaced_repetition_v2 extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$g, create_fragment$g, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Spaced_repetition_v2",
    			options,
    			id: create_fragment$g.name
    		});
    	}
    }

    /* src/components/interact/interact_v2.svelte generated by Svelte v3.24.0 */

    function create_fragment$h(ctx) {
    	let router;
    	let current;

    	router = new Router({
    			props: { routes: /*routes*/ ctx[0] },
    			$$inline: true
    		});

    	router.$on("conditionsFailed", /*conditionsFailed_handler*/ ctx[1]);

    	const block = {
    		c: function create() {
    			create_component(router.$$.fragment);
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
    		id: create_fragment$h.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$h($$self, $$props, $$invalidate) {
    	const routes = {
    		"/play/total_recall": Total_recall_v2,
    		"/play/slideshow": Slideshow_v2,
    		"/interact/spaced_repetition/add": Spaced_repetition_v2,
    		// Catch-all, must be last
    		"*": Nothing
    	};

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Interact_v2> was created with unknown prop '${key}'`);
    	});

    	let { $$slots = {}, $$scope } = $$props;
    	validate_slots("Interact_v2", $$slots, []);
    	const conditionsFailed_handler = event => replace("/");

    	$$self.$capture_state = () => ({
    		Router,
    		replace,
    		Nothing,
    		TotalRecall: Total_recall_v2,
    		Slideshow: Slideshow_v2,
    		SpacedRepetitionAdd: Spaced_repetition_v2,
    		routes
    	});

    	return [routes, conditionsFailed_handler];
    }

    class Interact_v2 extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$h, create_fragment$h, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Interact_v2",
    			options,
    			id: create_fragment$h.name
    		});
    	}
    }

    new Interact_v2({
    	target: document.querySelector("#play-screen"),
    });

    new App({
    	target: document.querySelector("#list-info"),
    });

}());
//# sourceMappingURL=bundle.js.map
