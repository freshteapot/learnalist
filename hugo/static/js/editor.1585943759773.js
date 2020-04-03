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
function get_store_value(store) {
    let value;
    subscribe(store, _ => value = _)();
    return value;
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
    if (value != null || input.value) {
        input.value = value;
    }
}
function set_style(node, key, value, important) {
    node.style.setProperty(key, value, important ? 'important' : '');
}
function select_option(select, value) {
    for (let i = 0; i < select.options.length; i += 1) {
        const option = select.options[i];
        if (option.__value === value) {
            option.selected = true;
            return;
        }
    }
}
function select_value(select) {
    const selected_option = select.querySelector(':checked') || select.options[0];
    return selected_option && selected_option.__value;
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
function afterUpdate(fn) {
    get_current_component().$$.after_update.push(fn);
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
function add_flush_callback(fn) {
    flush_callbacks.push(fn);
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

const globals = (typeof window !== 'undefined' ? window : global);

function get_spread_update(levels, updates) {
    const update = {};
    const to_null_out = {};
    const accounted_for = { $$scope: 1 };
    let i = levels.length;
    while (i--) {
        const o = levels[i];
        const n = updates[i];
        if (n) {
            for (const key in o) {
                if (!(key in n))
                    to_null_out[key] = 1;
            }
            for (const key in n) {
                if (!accounted_for[key]) {
                    update[key] = n[key];
                    accounted_for[key] = 1;
                }
            }
            levels[i] = n;
        }
        else {
            for (const key in o) {
                accounted_for[key] = 1;
            }
        }
    }
    for (const key in to_null_out) {
        if (!(key in update))
            update[key] = undefined;
    }
    return update;
}
function get_spread_object(spread_props) {
    return typeof spread_props === 'object' && spread_props !== null ? spread_props : {};
}

function bind(component, name, callback) {
    const index = component.$$.props[name];
    if (index !== undefined) {
        component.$$.bound[index] = callback;
        callback(component.$$.ctx[index]);
    }
}
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

const keys = {
  'lists.by.me': 'my.lists',
  'my.edited.lists': 'my.edited.lists',
  'last.screen': 'last.screen',
  'authentication.bearer': 'auth.bearer',
  'user.uuid': 'user.uuid',
  'settings.server': 'settings.server',
  'settings.install.defaults': 'settings.install.defaults',
};

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

function clear() {
  localStorage.clear();
  save(keys['settings.install.defaults'], true);
  save(keys['settings.server'], 'https://learnalist.net');
  // TODO why is this not showing up?
  save(keys['my.edited.lists'], []);
  save(keys['lists.by.me'], []);
}

var cache = {
  get,
  save,
  rm,
  clear,
  keys
};

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

/* node_modules/svelte-spa-router/Router.svelte generated by Svelte v3.20.1 */

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
function create_if_block(ctx) {
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
		id: create_if_block.name,
		type: "if",
		source: "(207:0) {#if componentParams}",
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

const location$1 = derived(loc, $loc => $loc.location);
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

function instance($$self, $$props, $$invalidate) {
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
		wrap,
		getLocation,
		loc,
		location: location$1,
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

function copyObject(item) {
	return JSON.parse(JSON.stringify(item))
}

function isStringEmpty(input) {
	return (input === '' || input === undefined);
}

function isDeviceMobile() {
	var check = false;
	(function (a) { if (/(android|bb\d+|meego).+mobile|avantgo|bada\/|blackberry|blazer|compal|elaine|fennec|hiptop|iemobile|ip(hone|od)|iris|kindle|lge |maemo|midp|mmp|mobile.+firefox|netfront|opera m(ob|in)i|palm( os)?|phone|p(ixi|re)\/|plucker|pocket|psp|series(4|6)0|symbian|treo|up\.(browser|link)|vodafone|wap|windows ce|xda|xiino/i.test(a) || /1207|6310|6590|3gso|4thp|50[1-6]i|770s|802s|a wa|abac|ac(er|oo|s\-)|ai(ko|rn)|al(av|ca|co)|amoi|an(ex|ny|yw)|aptu|ar(ch|go)|as(te|us)|attw|au(di|\-m|r |s )|avan|be(ck|ll|nq)|bi(lb|rd)|bl(ac|az)|br(e|v)w|bumb|bw\-(n|u)|c55\/|capi|ccwa|cdm\-|cell|chtm|cldc|cmd\-|co(mp|nd)|craw|da(it|ll|ng)|dbte|dc\-s|devi|dica|dmob|do(c|p)o|ds(12|\-d)|el(49|ai)|em(l2|ul)|er(ic|k0)|esl8|ez([4-7]0|os|wa|ze)|fetc|fly(\-|_)|g1 u|g560|gene|gf\-5|g\-mo|go(\.w|od)|gr(ad|un)|haie|hcit|hd\-(m|p|t)|hei\-|hi(pt|ta)|hp( i|ip)|hs\-c|ht(c(\-| |_|a|g|p|s|t)|tp)|hu(aw|tc)|i\-(20|go|ma)|i230|iac( |\-|\/)|ibro|idea|ig01|ikom|im1k|inno|ipaq|iris|ja(t|v)a|jbro|jemu|jigs|kddi|keji|kgt( |\/)|klon|kpt |kwc\-|kyo(c|k)|le(no|xi)|lg( g|\/(k|l|u)|50|54|\-[a-w])|libw|lynx|m1\-w|m3ga|m50\/|ma(te|ui|xo)|mc(01|21|ca)|m\-cr|me(rc|ri)|mi(o8|oa|ts)|mmef|mo(01|02|bi|de|do|t(\-| |o|v)|zz)|mt(50|p1|v )|mwbp|mywa|n10[0-2]|n20[2-3]|n30(0|2)|n50(0|2|5)|n7(0(0|1)|10)|ne((c|m)\-|on|tf|wf|wg|wt)|nok(6|i)|nzph|o2im|op(ti|wv)|oran|owg1|p800|pan(a|d|t)|pdxg|pg(13|\-([1-8]|c))|phil|pire|pl(ay|uc)|pn\-2|po(ck|rt|se)|prox|psio|pt\-g|qa\-a|qc(07|12|21|32|60|\-[2-7]|i\-)|qtek|r380|r600|raks|rim9|ro(ve|zo)|s55\/|sa(ge|ma|mm|ms|ny|va)|sc(01|h\-|oo|p\-)|sdk\/|se(c(\-|0|1)|47|mc|nd|ri)|sgh\-|shar|sie(\-|m)|sk\-0|sl(45|id)|sm(al|ar|b3|it|t5)|so(ft|ny)|sp(01|h\-|v\-|v )|sy(01|mb)|t2(18|50)|t6(00|10|18)|ta(gt|lk)|tcl\-|tdg\-|tel(i|m)|tim\-|t\-mo|to(pl|sh)|ts(70|m\-|m3|m5)|tx\-9|up(\.b|g1|si)|utst|v400|v750|veri|vi(rg|te)|vk(40|5[0-3]|\-v)|vm40|voda|vulc|vx(52|53|60|61|70|80|81|83|85|98)|w3c(\-| )|webc|whit|wi(g |nc|nw)|wmlb|wonu|x700|yas\-|your|zeto|zte\-/i.test(a.substr(0, 4))) check = true; })(navigator.userAgent || navigator.vendor || window.opera);
	return check;
}

function hasWhiteSpace(s) {
	return s.indexOf(' ') >= 0;
}

function focusThis(el) {
	el.focus();
}

function loginHelperSingleton() {
	const defaultRedirectURL = '/';
	let obj = {
		redirectURL: defaultRedirectURL,
		loggedIn: (() => {
			const auth = cache.get(cache.keys['authentication.bearer']);
			return auth ? true : false;
		})()
	};

	const { subscribe, set, update } = writable(obj);

	return {
		subscribe,

		login: ((session) => {
			cache.save(cache.keys['authentication.bearer'], session.token);
			update(n => {
				n.loggedIn = true;
				return n;
			});
		}),

		logout: () => {
			cache.clear();
			update(n => {
				n.loggedIn = false;
				return n;
			});
		},

		redirectURLAfterLogin: (redirectURL) => {
			if (isStringEmpty(redirectURL)) {
				redirectURL = defaultRedirectURL;
			}

			update(n => {
				n.redirectURL = redirectURL;
				return n;
			});
		}
	};
}

const loginHelper = loginHelperSingleton();

/* src/editor/components/menu_top.svelte generated by Svelte v3.20.1 */
const file = "src/editor/components/menu_top.svelte";

// (37:38) 
function create_if_block_1(ctx) {
	let a;
	let link_action;
	let dispose;

	const block = {
		c: function create() {
			a = element("a");
			a.textContent = "Login";
			attr_dev(a, "title", "Components");
			attr_dev(a, "href", "/login");
			attr_dev(a, "class", "f6 fw6 hover-red link black-70 mr2 mr3-m mr4-l dib");
			add_location(a, file, 37, 8, 919);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, a, anchor);
			if (remount) dispose();
			dispose = action_destroyer(link_action = link.call(null, a));
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(a);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_1.name,
		type: "if",
		source: "(37:38) ",
		ctx
	});

	return block;
}

// (20:6) {#if $loginHelper.loggedIn}
function create_if_block$1(ctx) {
	let a0;
	let link_action;
	let t1;
	let a1;
	let link_action_1;
	let dispose;

	const block = {
		c: function create() {
			a0 = element("a");
			a0.textContent = "Create";
			t1 = space();
			a1 = element("a");
			a1.textContent = "Find";
			attr_dev(a0, "title", "Documentation");
			attr_dev(a0, "href", "/create");
			attr_dev(a0, "class", "f6 fw6 hover-blue link black-70 mr2 mr3-m mr2-l di");
			add_location(a0, file, 20, 8, 488);
			attr_dev(a1, "title", "Components");
			attr_dev(a1, "href", "/lists/by/me");
			attr_dev(a1, "class", "f6 fw6 hover-blue link black-70 mr2 mr3-m mr5-l di");
			add_location(a1, file, 28, 8, 684);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, a0, anchor);
			insert_dev(target, t1, anchor);
			insert_dev(target, a1, anchor);
			if (remount) run_all(dispose);

			dispose = [
				action_destroyer(link_action = link.call(null, a0)),
				action_destroyer(link_action_1 = link.call(null, a1))
			];
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(a0);
			if (detaching) detach_dev(t1);
			if (detaching) detach_dev(a1);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$1.name,
		type: "if",
		source: "(20:6) {#if $loginHelper.loggedIn}",
		ctx
	});

	return block;
}

function create_fragment$1(ctx) {
	let header;
	let div0;
	let a;
	let small;
	let link_action;
	let t1;
	let div2;
	let div1;
	let dispose;

	function select_block_type(ctx, dirty) {
		if (/*$loginHelper*/ ctx[0].loggedIn) return create_if_block$1;
		if (/*$location*/ ctx[1] != "/login") return create_if_block_1;
	}

	let current_block_type = select_block_type(ctx);
	let if_block = current_block_type && current_block_type(ctx);

	const block = {
		c: function create() {
			header = element("header");
			div0 = element("div");
			a = element("a");
			small = element("small");
			small.textContent = "Learnalist";
			t1 = space();
			div2 = element("div");
			div1 = element("div");
			if (if_block) if_block.c();
			attr_dev(small, "class", "nowrap f9 mt2 mt3-ns pr2 black-70 fw9");
			add_location(small, file, 13, 6, 295);
			attr_dev(a, "href", "/");
			attr_dev(a, "class", "di f5 f4-ns fw6 mt0 mb1 link black-20");
			attr_dev(a, "title", "Home");
			add_location(a, file, 7, 4, 179);
			attr_dev(div0, "class", "w-25 pa3 mr2");
			add_location(div0, file, 6, 2, 148);
			attr_dev(div1, "class", "fr mt0");
			add_location(div1, file, 18, 4, 425);
			attr_dev(div2, "class", "w-75 pa3 items-end");
			add_location(div2, file, 17, 2, 388);
			attr_dev(header, "class", "flex");
			add_location(header, file, 5, 0, 124);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, header, anchor);
			append_dev(header, div0);
			append_dev(div0, a);
			append_dev(a, small);
			append_dev(header, t1);
			append_dev(header, div2);
			append_dev(div2, div1);
			if (if_block) if_block.m(div1, null);
			if (remount) dispose();
			dispose = action_destroyer(link_action = link.call(null, a));
		},
		p: function update(ctx, [dirty]) {
			if (current_block_type !== (current_block_type = select_block_type(ctx))) {
				if (if_block) if_block.d(1);
				if_block = current_block_type && current_block_type(ctx);

				if (if_block) {
					if_block.c();
					if_block.m(div1, null);
				}
			}
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(header);

			if (if_block) {
				if_block.d();
			}

			dispose();
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

function instance$1($$self, $$props, $$invalidate) {
	let $loginHelper;
	let $location;
	validate_store(loginHelper, "loginHelper");
	component_subscribe($$self, loginHelper, $$value => $$invalidate(0, $loginHelper = $$value));
	validate_store(location$1, "location");
	component_subscribe($$self, location$1, $$value => $$invalidate(1, $location = $$value));
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Menu_top> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Menu_top", $$slots, []);

	$$self.$capture_state = () => ({
		link,
		location: location$1,
		loginHelper,
		$loginHelper,
		$location
	});

	return [$loginHelper, $location];
}

class Menu_top extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$1, create_fragment$1, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Menu_top",
			options,
			id: create_fragment$1.name
		});
	}
}

/* src/editor/components/footer.svelte generated by Svelte v3.20.1 */
const file$1 = "src/editor/components/footer.svelte";

// (8:4) {#if $loginHelper.loggedIn}
function create_if_block$2(ctx) {
	let div;
	let a;
	let link_action;
	let dispose;

	const block = {
		c: function create() {
			div = element("div");
			a = element("a");
			a.textContent = "Logout";
			attr_dev(a, "title", "Components");
			attr_dev(a, "href", "/logout");
			attr_dev(a, "class", "f6 fw6 hover-red link black-70 mr2 mr3-m mr4-l dib");
			add_location(a, file$1, 9, 8, 265);
			attr_dev(div, "class", "fr");
			add_location(div, file$1, 8, 6, 240);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div, anchor);
			append_dev(div, a);
			if (remount) dispose();
			dispose = action_destroyer(link_action = link.call(null, a));
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$2.name,
		type: "if",
		source: "(8:4) {#if $loginHelper.loggedIn}",
		ctx
	});

	return block;
}

function create_fragment$2(ctx) {
	let header;
	let div;
	let if_block = /*$loginHelper*/ ctx[0].loggedIn && create_if_block$2(ctx);

	const block = {
		c: function create() {
			header = element("header");
			div = element("div");
			if (if_block) if_block.c();
			attr_dev(div, "class", "w-100 pa3 items-end");
			add_location(div, file$1, 6, 2, 168);
			attr_dev(header, "class", "flex w-100 bt b--black-10 bg-white");
			add_location(header, file$1, 5, 0, 114);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, header, anchor);
			append_dev(header, div);
			if (if_block) if_block.m(div, null);
		},
		p: function update(ctx, [dirty]) {
			if (/*$loginHelper*/ ctx[0].loggedIn) {
				if (!if_block) {
					if_block = create_if_block$2(ctx);
					if_block.c();
					if_block.m(div, null);
				}
			} else if (if_block) {
				if_block.d(1);
				if_block = null;
			}
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(header);
			if (if_block) if_block.d();
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
	let $loginHelper;
	validate_store(loginHelper, "loginHelper");
	component_subscribe($$self, loginHelper, $$value => $$invalidate(0, $loginHelper = $$value));
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Footer> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Footer", $$slots, []);
	$$self.$capture_state = () => ({ link, loginHelper, $loginHelper });
	return [$loginHelper];
}

class Footer extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$2, create_fragment$2, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Footer",
			options,
			id: create_fragment$2.name
		});
	}
}

/* src/editor/components/menu.svelte generated by Svelte v3.20.1 */
const file$2 = "src/editor/components/menu.svelte";

function create_fragment$3(ctx) {
	let header;
	let div;
	let button0;
	let t1;
	let button1;
	let t3;
	let span;
	let t5;
	let button2;
	let t7;
	let button3;
	let dispose;

	const block = {
		c: function create() {
			header = element("header");
			div = element("div");
			button0 = element("button");
			button0.textContent = "My Lists";
			t1 = space();
			button1 = element("button");
			button1.textContent = "My Labels";
			t3 = space();
			span = element("span");
			span.textContent = "/";
			t5 = space();
			button2 = element("button");
			button2.textContent = "New List";
			t7 = space();
			button3 = element("button");
			button3.textContent = "New Label";
			add_location(button0, file$2, 7, 4, 174);
			add_location(button1, file$2, 8, 4, 244);
			add_location(span, file$2, 9, 4, 316);
			add_location(button2, file$2, 10, 4, 335);
			add_location(button3, file$2, 11, 4, 402);
			attr_dev(div, "class", "dtc v-mid tc white ph3 ph4-l");
			add_location(div, file$2, 6, 2, 127);
			attr_dev(header, "class", "pt2 dt w-100");
			add_location(header, file$2, 5, 0, 95);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, header, anchor);
			append_dev(header, div);
			append_dev(div, button0);
			append_dev(div, t1);
			append_dev(div, button1);
			append_dev(div, t3);
			append_dev(div, span);
			append_dev(div, t5);
			append_dev(div, button2);
			append_dev(div, t7);
			append_dev(div, button3);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(button0, "click", /*click_handler*/ ctx[1], false, false, false),
				listen_dev(button1, "click", /*click_handler_1*/ ctx[2], false, false, false),
				listen_dev(button2, "click", /*click_handler_2*/ ctx[3], false, false, false),
				listen_dev(button3, "click", /*click_handler_3*/ ctx[4], false, false, false)
			];
		},
		p: noop,
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(header);
			run_all(dispose);
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
	let isLoggedIn = true;
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Menu> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Menu", $$slots, []);
	const click_handler = () => push("/lists/by/me");
	const click_handler_1 = () => push("/labels/by/me");
	const click_handler_2 = () => push("/list/new");
	const click_handler_3 = () => push("/list/new");
	$$self.$capture_state = () => ({ push, link, isLoggedIn });

	$$self.$inject_state = $$props => {
		if ("isLoggedIn" in $$props) isLoggedIn = $$props.isLoggedIn;
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [isLoggedIn, click_handler, click_handler_1, click_handler_2, click_handler_3];
}

class Menu extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$3, create_fragment$3, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Menu",
			options,
			id: create_fragment$3.name
		});
	}
}

/* src/editor/components/Box.svelte generated by Svelte v3.20.1 */

const file$3 = "src/editor/components/Box.svelte";

function add_css() {
	var style = element("style");
	style.id = "svelte-16tveot-style";
	style.textContent = ".box.svelte-16tveot{border:1px solid #aaa;border-radius:2px;box-shadow:2px 2px 8px rgba(0, 0, 0, 0.1);padding:1em;margin:0 0 1em 0}\n/*# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiQm94LnN2ZWx0ZSIsInNvdXJjZXMiOlsiQm94LnN2ZWx0ZSJdLCJzb3VyY2VzQ29udGVudCI6WyI8c3R5bGU+XG4gIC5ib3gge1xuICAgIGJvcmRlcjogMXB4IHNvbGlkICNhYWE7XG4gICAgYm9yZGVyLXJhZGl1czogMnB4O1xuICAgIGJveC1zaGFkb3c6IDJweCAycHggOHB4IHJnYmEoMCwgMCwgMCwgMC4xKTtcbiAgICBwYWRkaW5nOiAxZW07XG4gICAgbWFyZ2luOiAwIDAgMWVtIDA7XG4gIH1cblxuLyojIHNvdXJjZU1hcHBpbmdVUkw9ZGF0YTphcHBsaWNhdGlvbi9qc29uO2Jhc2U2NCxleUoyWlhKemFXOXVJam96TENKemIzVnlZMlZ6SWpwYkluTnlZeTlsWkdsMGIzSXZZMjl0Y0c5dVpXNTBjeTlDYjNndWMzWmxiSFJsSWwwc0ltNWhiV1Z6SWpwYlhTd2liV0Z3Y0dsdVozTWlPaUk3UlVGRFJUdEpRVU5GTEhOQ1FVRnpRanRKUVVOMFFpeHJRa0ZCYTBJN1NVRkRiRUlzTUVOQlFUQkRPMGxCUXpGRExGbEJRVms3U1VGRFdpeHBRa0ZCYVVJN1JVRkRia0lpTENKbWFXeGxJam9pYzNKakwyVmthWFJ2Y2k5amIyMXdiMjVsYm5SekwwSnZlQzV6ZG1Wc2RHVWlMQ0p6YjNWeVkyVnpRMjl1ZEdWdWRDSTZXeUpjYmlBZ0xtSnZlQ0I3WEc0Z0lDQWdZbTl5WkdWeU9pQXhjSGdnYzI5c2FXUWdJMkZoWVR0Y2JpQWdJQ0JpYjNKa1pYSXRjbUZrYVhWek9pQXljSGc3WEc0Z0lDQWdZbTk0TFhOb1lXUnZkem9nTW5CNElESndlQ0E0Y0hnZ2NtZGlZU2d3TENBd0xDQXdMQ0F3TGpFcE8xeHVJQ0FnSUhCaFpHUnBibWM2SURGbGJUdGNiaUFnSUNCdFlYSm5hVzQ2SURBZ01DQXhaVzBnTUR0Y2JpQWdmVnh1SWwxOSAqLzwvc3R5bGU+XG5cbjxkaXYgY2xhc3M9XCJib3hcIj5cbiAgPHNsb3QgLz5cbjwvZGl2PlxuIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUNFLElBQUksZUFBQyxDQUFDLEFBQ0osTUFBTSxDQUFFLEdBQUcsQ0FBQyxLQUFLLENBQUMsSUFBSSxDQUN0QixhQUFhLENBQUUsR0FBRyxDQUNsQixVQUFVLENBQUUsR0FBRyxDQUFDLEdBQUcsQ0FBQyxHQUFHLENBQUMsS0FBSyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FDMUMsT0FBTyxDQUFFLEdBQUcsQ0FDWixNQUFNLENBQUUsQ0FBQyxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxBQUNuQixDQUFDIn0= */";
	append_dev(document.head, style);
}

function create_fragment$4(ctx) {
	let div;
	let current;
	const default_slot_template = /*$$slots*/ ctx[1].default;
	const default_slot = create_slot(default_slot_template, ctx, /*$$scope*/ ctx[0], null);

	const block = {
		c: function create() {
			div = element("div");
			if (default_slot) default_slot.c();
			attr_dev(div, "class", "box svelte-16tveot");
			add_location(div, file$3, 11, 0, 738);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, div, anchor);

			if (default_slot) {
				default_slot.m(div, null);
			}

			current = true;
		},
		p: function update(ctx, [dirty]) {
			if (default_slot) {
				if (default_slot.p && dirty & /*$$scope*/ 1) {
					default_slot.p(get_slot_context(default_slot_template, ctx, /*$$scope*/ ctx[0], null), get_slot_changes(default_slot_template, /*$$scope*/ ctx[0], dirty, null));
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
			if (detaching) detach_dev(div);
			if (default_slot) default_slot.d(detaching);
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
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Box> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Box", $$slots, ['default']);

	$$self.$set = $$props => {
		if ("$$scope" in $$props) $$invalidate(0, $$scope = $$props.$$scope);
	};

	return [$$scope, $$slots];
}

class Box extends SvelteComponentDev {
	constructor(options) {
		super(options);
		if (!document.getElementById("svelte-16tveot-style")) add_css();
		init(this, options, instance$4, create_fragment$4, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Box",
			options,
			id: create_fragment$4.name
		});
	}
}

// List of nodes to update
const nodes = [];

// Current location
let location$2;

// Function that updates all nodes marking the active ones
function checkActive(el) {
    // Remove the active class from all elements
    el.node.classList.remove(el.className);

    // If the pattern matches, then set the active class
    if (el.pattern.test(location$2)) {
        el.node.classList.add(el.className);
    }
}

// Listen to changes in the location
loc.subscribe((value) => {
    // Update the location
    location$2 = value.location + (value.querystring ? '?' + value.querystring : '');

    // Update all nodes
    nodes.map(checkActive);
});

/**
 * @typedef {Object} ActiveOptions
 * @property {string|RegExp} [path] - Path expression that makes the link active when matched (must start with '/' or '*'); default is the link's href
 * @property {string} [className] - CSS class to apply to the element when active; default value is "active"
 */

/**
 * Svelte Action for automatically adding the "active" class to elements (links, or any other DOM element) when the current location matches a certain path.
 * 
 * @param {HTMLElement} node - The target node (automatically set by Svelte)
 * @param {ActiveOptions|string|RegExp} [opts] - Can be an object of type ActiveOptions, or a string (or regular expressions) representing ActiveOptions.path.
 */
function active(node, opts) {
    // Check options
    if (opts && (typeof opts == 'string' || (typeof opts == 'object' && opts instanceof RegExp))) {
        // Interpret strings and regular expressions as opts.path
        opts = {
            path: opts
        };
    }
    else {
        // Ensure opts is a dictionary
        opts = opts || {};
    }

    // Path defaults to link target
    if (!opts.path && node.hasAttribute('href')) {
        opts.path = node.getAttribute('href');
        if (opts.path && opts.path.length > 1 && opts.path.charAt(0) == '#') {
            opts.path = opts.path.substring(1);
        }
    }

    // Default class name
    if (!opts.className) {
        opts.className = 'active';
    }

    // If path is a string, it must start with '/' or '*'
    if (!opts.path || 
        typeof opts.path == 'string' && (opts.path.length < 1 || (opts.path.charAt(0) != '/' && opts.path.charAt(0) != '*'))
    ) {
        throw Error('Invalid value for "path" argument')
    }

    // If path is not a regular expression already, make it
    const {pattern} = typeof opts.path == 'string' ?
        regexparam(opts.path) :
        {pattern: opts.path};

    // Add the node to the list
    const el = {
        node,
        className: opts.className,
        pattern
    };
    nodes.push(el);

    // Trigger the action right away
    checkActive(el);

    return {
        // When the element is destroyed, remove it from the list
        destroy() {
            nodes.splice(nodes.indexOf(el), 1);
        }
    }
}

/* src/editor/routes/home.svelte generated by Svelte v3.20.1 */
const file$4 = "src/editor/routes/home.svelte";

function create_fragment$5(ctx) {
	let h1;
	let t1;
	let ul;
	let li0;
	let a0;
	let link_action;
	let t3;
	let li1;
	let t5;
	let li2;
	let a1;
	let link_action_1;
	let t7;
	let li3;
	let a2;
	let dispose;

	const block = {
		c: function create() {
			h1 = element("h1");
			h1.textContent = "The Editor";
			t1 = space();
			ul = element("ul");
			li0 = element("li");
			a0 = element("a");
			a0.textContent = "Create";
			t3 = space();
			li1 = element("li");
			li1.textContent = "Share";
			t5 = space();
			li2 = element("li");
			a1 = element("a");
			a1.textContent = "Server information";
			t7 = space();
			li3 = element("li");
			a2 = element("a");
			a2.textContent = "Clear cache";
			add_location(h1, file$4, 12, 0, 251);
			attr_dev(a0, "class", "f5 link black");
			attr_dev(a0, "href", "/create");
			add_location(a0, file$4, 16, 4, 288);
			add_location(li0, file$4, 15, 2, 279);
			add_location(li1, file$4, 18, 2, 358);
			attr_dev(a1, "class", "f5 link black");
			attr_dev(a1, "href", "/settings/server_information");
			add_location(a1, file$4, 20, 4, 384);
			add_location(li2, file$4, 19, 2, 375);
			attr_dev(a2, "class", "f5 link black");
			attr_dev(a2, "href", "#");
			add_location(a2, file$4, 25, 4, 508);
			add_location(li3, file$4, 24, 2, 499);
			add_location(ul, file$4, 14, 0, 272);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, h1, anchor);
			insert_dev(target, t1, anchor);
			insert_dev(target, ul, anchor);
			append_dev(ul, li0);
			append_dev(li0, a0);
			append_dev(ul, t3);
			append_dev(ul, li1);
			append_dev(ul, t5);
			append_dev(ul, li2);
			append_dev(li2, a1);
			append_dev(ul, t7);
			append_dev(ul, li3);
			append_dev(li3, a2);
			if (remount) run_all(dispose);

			dispose = [
				action_destroyer(link_action = link.call(null, a0)),
				action_destroyer(link_action_1 = link.call(null, a1)),
				listen_dev(a2, "click", /*click_handler*/ ctx[0], false, false, false)
			];
		},
		p: noop,
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(h1);
			if (detaching) detach_dev(t1);
			if (detaching) detach_dev(ul);
			run_all(dispose);
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

function reset() {
	cache.clear();
	loginHelper.logout();
	replace("/");
}

function instance$5($$self, $$props, $$invalidate) {
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Home> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Home", $$slots, []);
	const click_handler = () => reset();
	$$self.$capture_state = () => ({ link, replace, loginHelper, cache, reset });
	return [click_handler];
}

class Home extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$5, create_fragment$5, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Home",
			options,
			id: create_fragment$5.name
		});
	}
}

function getAuth() {
  if (!localStorage.hasOwnProperty(cache.keys['authentication.bearer'])) {
    throw new Error('login.required');
  }
  return "Bearer " + cache.get(cache.keys['authentication.bearer']);
}

function getServer() {
  if (!localStorage.hasOwnProperty(cache.keys['settings.server'])) {
    throw new Error('settings.server.missing');
  }
  return cache.get(cache.keys['settings.server']);
}


function getHeaders() {
  return {
    "Content-Type": "application/json",
    Authorization: getAuth()
  };
}

async function getVersion() {
  const url = getServer() + "/api/v1/version";
  const res = await fetch(url);
  const data = await res.json();

  if (res.ok) {
    return data;
  }
  throw new Error("Failed to get learnalist server version information");
}

async function getListsByMe() {
  const url = getServer() + "/api/v1/alist/by/me";
  const res = await fetch(url, {
    headers: getHeaders()
  });

  let manyLists = await res.json();
  if (res.ok) {
    return manyLists;
  }
  throw new Error("Failed to get lists by me");
}


async function putList(aList) {
  const response = {
    status: 400,
    data: {}
  };

  const url = getServer() + '/api/v1/alist/' + aList.uuid;
  const res = await fetch(url, {
    method: 'PUT',
    headers: getHeaders(),
    body: JSON.stringify(aList)
  });

  const data = await res.json();
  switch(res.status) {
    case 200:
    case 403:
    case 400:
      response.status = res.status;
      response.data = data;
      return response;
  }
  throw new Error('Unexpected response from the server');
}


// Look at https://github.com/freshteapot/learnalist-api/blob/master/server/doc/api.user.login.md
async function postLogin(username, password) {
  const response = {
    status: 400,
    body: {}
  };

  const input = {
    username: username,
    password: password
  };

  const url = getServer() + "/api/v1/user/login";
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(input)
  });

  const data = await res.json();
  switch(res.status) {
    case 200:
    case 403:
    case 400:
      response.status = res.status;
      response.body = data;
      return response;
  }
  throw new Error('Unexpected response from the server');
}


// postList title: string listType: string
async function postList(title, listType) {
  const input = {
    data: [],
    info: {
      title: title,
      type: listType,
      labels: []
    }
  };

  const url = getServer() + '/api/v1/alist';
  const res = await fetch(url, {
    method: "POST",
    headers: getHeaders(),
    body: JSON.stringify(input)
  });

  const data = await res.json();
  // TODO double check we handle the codes we send from the server.
  if (res.status === 400 || res.status === 201) {
    return {
      status: res.status,
      body: data
    };
  }
  throw new Error('Unexpected response from the server when posting a list');
}

// deleteList uuid: string
async function deleteList(uuid) {
  const url = getServer() + "/api/v1/alist/" + uuid;
  const res = await fetch(url, {
    method: "DELETE",
    headers: getHeaders()
  });

  const data = await res.json();
  // TODO double check we handle the codes we send from the server.
  if (res.status === 400 || res.status === 200 || res.status === 404) {
    return {
      status: res.status,
      body: data
    };
  }
  throw new Error('Unexpected response from the server when deleting a list');
}

/* src/editor/components/error_box.svelte generated by Svelte v3.20.1 */
const file$5 = "src/editor/components/error_box.svelte";

function create_fragment$6(ctx) {
	let article;
	let h2;
	let t;
	let dispose;

	const block = {
		c: function create() {
			article = element("article");
			h2 = element("h2");
			t = text(/*message*/ ctx[0]);
			attr_dev(h2, "class", "f4 br3 b--red mw-100 black-70 mv0 pv2 ph4");
			add_location(h2, file$5, 17, 2, 307);
			attr_dev(article, "class", "mw10 bt bw3 b--red bg-washed-red mw-100");
			attr_dev(article, "title", "Click to dismiss.");
			add_location(article, file$5, 12, 0, 187);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, article, anchor);
			append_dev(article, h2);
			append_dev(h2, t);
			if (remount) dispose();
			dispose = listen_dev(article, "click", /*click_handler*/ ctx[3], false, false, false);
		},
		p: function update(ctx, [dirty]) {
			if (dirty & /*message*/ 1) set_data_dev(t, /*message*/ ctx[0]);
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(article);
			dispose();
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
	let { message } = $$props;
	const dispatch = createEventDispatcher();

	function clear() {
		dispatch("clear");
	}

	const writable_props = ["message"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Error_box> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Error_box", $$slots, []);
	const click_handler = () => clear();

	$$self.$set = $$props => {
		if ("message" in $$props) $$invalidate(0, message = $$props.message);
	};

	$$self.$capture_state = () => ({
		message,
		createEventDispatcher,
		dispatch,
		clear
	});

	$$self.$inject_state = $$props => {
		if ("message" in $$props) $$invalidate(0, message = $$props.message);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [message, clear, dispatch, click_handler];
}

class Error_box extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$6, create_fragment$6, safe_not_equal, { message: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Error_box",
			options,
			id: create_fragment$6.name
		});

		const { ctx } = this.$$;
		const props = options.props || {};

		if (/*message*/ ctx[0] === undefined && !("message" in props)) {
			console.warn("<Error_box> was created without expected prop 'message'");
		}
	}

	get message() {
		throw new Error("<Error_box>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set message(value) {
		throw new Error("<Error_box>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/login.svelte generated by Svelte v3.20.1 */
const file$6 = "src/editor/components/login.svelte";

// (35:0) {#if message}
function create_if_block_1$1(ctx) {
	let current;

	const errorbox = new Error_box({
			props: { message: /*message*/ ctx[2] },
			$$inline: true
		});

	errorbox.$on("clear", /*clearMessage*/ ctx[5]);

	const block = {
		c: function create() {
			create_component(errorbox.$$.fragment);
		},
		m: function mount(target, anchor) {
			mount_component(errorbox, target, anchor);
			current = true;
		},
		p: function update(ctx, dirty) {
			const errorbox_changes = {};
			if (dirty & /*message*/ 4) errorbox_changes.message = /*message*/ ctx[2];
			errorbox.$set(errorbox_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(errorbox.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(errorbox.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(errorbox, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_1$1.name,
		type: "if",
		source: "(35:0) {#if message}",
		ctx
	});

	return block;
}

// (103:2) {:else}
function create_else_block$1(ctx) {
	let p;
	let t0;
	let br;
	let t1;
	let a;

	const block = {
		c: function create() {
			p = element("p");
			t0 = text("You are already logged in.\n      ");
			br = element("br");
			t1 = text("\n      Goto the\n      ");
			a = element("a");
			a.textContent = "welcome page";
			add_location(br, file$6, 105, 6, 2990);
			attr_dev(a, "href", "/welcome.html");
			add_location(a, file$6, 107, 6, 3018);
			attr_dev(p, "class", "measure center");
			add_location(p, file$6, 103, 4, 2924);
		},
		m: function mount(target, anchor) {
			insert_dev(target, p, anchor);
			append_dev(p, t0);
			append_dev(p, br);
			append_dev(p, t1);
			append_dev(p, a);
		},
		p: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(p);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_else_block$1.name,
		type: "else",
		source: "(103:2) {:else}",
		ctx
	});

	return block;
}

// (39:2) {#if !isLoggedIn}
function create_if_block$3(ctx) {
	let form;
	let fieldset;
	let div0;
	let label0;
	let t1;
	let input0;
	let t2;
	let div1;
	let label1;
	let t4;
	let input1;
	let t5;
	let div7;
	let div6;
	let div5;
	let div2;
	let button;
	let t7;
	let div3;
	let span0;
	let t8;
	let a0;
	let t10;
	let div4;
	let span1;
	let t11;
	let a1;
	let dispose;

	const block = {
		c: function create() {
			form = element("form");
			fieldset = element("fieldset");
			div0 = element("div");
			label0 = element("label");
			label0.textContent = "Username";
			t1 = space();
			input0 = element("input");
			t2 = space();
			div1 = element("div");
			label1 = element("label");
			label1.textContent = "Password";
			t4 = space();
			input1 = element("input");
			t5 = space();
			div7 = element("div");
			div6 = element("div");
			div5 = element("div");
			div2 = element("div");
			button = element("button");
			button.textContent = "Login";
			t7 = space();
			div3 = element("div");
			span0 = element("span");
			t8 = text("or with\n                ");
			a0 = element("a");
			a0.textContent = "google";
			t10 = space();
			div4 = element("div");
			span1 = element("span");
			t11 = text("or via\n                ");
			a1 = element("a");
			a1.textContent = "learnalist login";
			attr_dev(label0, "class", "db fw6 lh-copy f6");
			attr_dev(label0, "for", "username");
			add_location(label0, file$6, 43, 10, 1085);
			attr_dev(input0, "class", "pa2 input-reset ba bg-transparent b--black-20 w-100 br2");
			attr_dev(input0, "type", "text");
			attr_dev(input0, "name", "username");
			attr_dev(input0, "id", "username");
			attr_dev(input0, "autocapitalize", "none");
			add_location(input0, file$6, 44, 10, 1160);
			attr_dev(div0, "class", "mt3");
			add_location(div0, file$6, 42, 8, 1057);
			attr_dev(label1, "class", "db fw6 lh-copy f6");
			attr_dev(label1, "for", "password");
			add_location(label1, file$6, 54, 10, 1455);
			attr_dev(input1, "class", "b pa2 input-reset ba bg-transparent b--black-20 w-100 br2");
			attr_dev(input1, "type", "password");
			attr_dev(input1, "name", "password");
			attr_dev(input1, "autocomplete", "off");
			attr_dev(input1, "id", "password");
			add_location(input1, file$6, 55, 10, 1530);
			attr_dev(div1, "class", "mv3");
			add_location(div1, file$6, 53, 8, 1427);
			attr_dev(fieldset, "id", "sign_up");
			attr_dev(fieldset, "class", "ba b--transparent ph0 mh0");
			add_location(fieldset, file$6, 41, 6, 991);
			attr_dev(button, "class", "db w-100");
			attr_dev(button, "type", "submit");
			add_location(button, file$6, 70, 14, 1971);
			attr_dev(div2, "class", "flex items-center mb2");
			add_location(div2, file$6, 69, 12, 1921);
			attr_dev(a0, "target", "_blank");
			attr_dev(a0, "href", "https://learnalist.net/api/v1/oauth/google/redirect");
			attr_dev(a0, "class", "f6 link underline dib black");
			add_location(a0, file$6, 76, 16, 2180);
			attr_dev(span0, "class", "f6 link dib black");
			add_location(span0, file$6, 74, 14, 2107);
			attr_dev(div3, "class", "flex items-center mb2");
			add_location(div3, file$6, 73, 12, 2057);
			attr_dev(a1, "target", "_blank");
			attr_dev(a1, "href", "https://learnalist.net/login.html");
			attr_dev(a1, "class", "f6 link underline dib black");
			add_location(a1, file$6, 89, 16, 2588);
			attr_dev(span1, "class", "f6 link dib black");
			add_location(span1, file$6, 87, 14, 2516);
			attr_dev(div4, "class", "flex items-center mb2");
			add_location(div4, file$6, 86, 12, 2466);
			attr_dev(div5, "class", "fr");
			add_location(div5, file$6, 68, 10, 1892);
			attr_dev(div6, "class", "w-100 items-end");
			add_location(div6, file$6, 67, 8, 1852);
			attr_dev(div7, "class", "measure flex");
			add_location(div7, file$6, 66, 6, 1817);
			attr_dev(form, "class", "measure center");
			add_location(form, file$6, 39, 4, 912);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, form, anchor);
			append_dev(form, fieldset);
			append_dev(fieldset, div0);
			append_dev(div0, label0);
			append_dev(div0, t1);
			append_dev(div0, input0);
			set_input_value(input0, /*username*/ ctx[0]);
			append_dev(fieldset, t2);
			append_dev(fieldset, div1);
			append_dev(div1, label1);
			append_dev(div1, t4);
			append_dev(div1, input1);
			set_input_value(input1, /*password*/ ctx[1]);
			append_dev(form, t5);
			append_dev(form, div7);
			append_dev(div7, div6);
			append_dev(div6, div5);
			append_dev(div5, div2);
			append_dev(div2, button);
			append_dev(div5, t7);
			append_dev(div5, div3);
			append_dev(div3, span0);
			append_dev(span0, t8);
			append_dev(span0, a0);
			append_dev(div5, t10);
			append_dev(div5, div4);
			append_dev(div4, span1);
			append_dev(span1, t11);
			append_dev(span1, a1);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(input0, "input", /*input0_input_handler*/ ctx[7]),
				listen_dev(input1, "input", /*input1_input_handler*/ ctx[8]),
				listen_dev(form, "submit", prevent_default(/*handleSubmit*/ ctx[4]), false, true, false)
			];
		},
		p: function update(ctx, dirty) {
			if (dirty & /*username*/ 1 && input0.value !== /*username*/ ctx[0]) {
				set_input_value(input0, /*username*/ ctx[0]);
			}

			if (dirty & /*password*/ 2 && input1.value !== /*password*/ ctx[1]) {
				set_input_value(input1, /*password*/ ctx[1]);
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(form);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$3.name,
		type: "if",
		source: "(39:2) {#if !isLoggedIn}",
		ctx
	});

	return block;
}

function create_fragment$7(ctx) {
	let t;
	let main;
	let current;
	let if_block0 = /*message*/ ctx[2] && create_if_block_1$1(ctx);

	function select_block_type(ctx, dirty) {
		if (!/*isLoggedIn*/ ctx[3]) return create_if_block$3;
		return create_else_block$1;
	}

	let current_block_type = select_block_type(ctx);
	let if_block1 = current_block_type(ctx);

	const block = {
		c: function create() {
			if (if_block0) if_block0.c();
			t = space();
			main = element("main");
			if_block1.c();
			attr_dev(main, "class", "pa4 black-80");
			add_location(main, file$6, 37, 0, 860);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			if (if_block0) if_block0.m(target, anchor);
			insert_dev(target, t, anchor);
			insert_dev(target, main, anchor);
			if_block1.m(main, null);
			current = true;
		},
		p: function update(ctx, [dirty]) {
			if (/*message*/ ctx[2]) {
				if (if_block0) {
					if_block0.p(ctx, dirty);
					transition_in(if_block0, 1);
				} else {
					if_block0 = create_if_block_1$1(ctx);
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

			if_block1.p(ctx, dirty);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(if_block0);
			current = true;
		},
		o: function outro(local) {
			transition_out(if_block0);
			current = false;
		},
		d: function destroy(detaching) {
			if (if_block0) if_block0.d(detaching);
			if (detaching) detach_dev(t);
			if (detaching) detach_dev(main);
			if_block1.d();
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
	let $loginHelper;
	validate_store(loginHelper, "loginHelper");
	component_subscribe($$self, loginHelper, $$value => $$invalidate(6, $loginHelper = $$value));
	let isLoggedIn = $loginHelper.loggedIn;
	let username = "";
	let password = "";
	let message;

	async function handleSubmit() {
		if (username === "" || password === "") {
			$$invalidate(2, message = "Please enter in a username and password");
			return;
		}

		let response = await postLogin(username, password);

		if (response.status != 200) {
			alert("Try again");
			return;
		}

		loginHelper.login(response.body);
		push($loginHelper.redirectURL);

		// loginHelper.redirectURLAfterLogin();
		return;
	}

	function clearMessage() {
		$$invalidate(2, message = null);
	}

	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Login> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Login", $$slots, []);

	function input0_input_handler() {
		username = this.value;
		$$invalidate(0, username);
	}

	function input1_input_handler() {
		password = this.value;
		$$invalidate(1, password);
	}

	$$self.$capture_state = () => ({
		push,
		loginHelper,
		postLogin,
		ErrorBox: Error_box,
		isLoggedIn,
		username,
		password,
		message,
		handleSubmit,
		clearMessage,
		$loginHelper
	});

	$$self.$inject_state = $$props => {
		if ("isLoggedIn" in $$props) $$invalidate(3, isLoggedIn = $$props.isLoggedIn);
		if ("username" in $$props) $$invalidate(0, username = $$props.username);
		if ("password" in $$props) $$invalidate(1, password = $$props.password);
		if ("message" in $$props) $$invalidate(2, message = $$props.message);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [
		username,
		password,
		message,
		isLoggedIn,
		handleSubmit,
		clearMessage,
		$loginHelper,
		input0_input_handler,
		input1_input_handler
	];
}

class Login extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$7, create_fragment$7, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Login",
			options,
			id: create_fragment$7.name
		});
	}
}

/* src/editor/routes/login.svelte generated by Svelte v3.20.1 */
const file$7 = "src/editor/routes/login.svelte";

// (7:0) {#if !$loginHelper.loggedIn}
function create_if_block_1$2(ctx) {
	let current;
	const login = new Login({ $$inline: true });

	const block = {
		c: function create() {
			create_component(login.$$.fragment);
		},
		m: function mount(target, anchor) {
			mount_component(login, target, anchor);
			current = true;
		},
		i: function intro(local) {
			if (current) return;
			transition_in(login.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(login.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(login, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_1$2.name,
		type: "if",
		source: "(7:0) {#if !$loginHelper.loggedIn}",
		ctx
	});

	return block;
}

// (11:0) {#if $loginHelper.loggedIn}
function create_if_block$4(ctx) {
	let p;

	const block = {
		c: function create() {
			p = element("p");
			p.textContent = "You are logged in";
			add_location(p, file$7, 11, 2, 252);
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
		id: create_if_block$4.name,
		type: "if",
		source: "(11:0) {#if $loginHelper.loggedIn}",
		ctx
	});

	return block;
}

function create_fragment$8(ctx) {
	let t;
	let if_block1_anchor;
	let current;
	let if_block0 = !/*$loginHelper*/ ctx[0].loggedIn && create_if_block_1$2(ctx);
	let if_block1 = /*$loginHelper*/ ctx[0].loggedIn && create_if_block$4(ctx);

	const block = {
		c: function create() {
			if (if_block0) if_block0.c();
			t = space();
			if (if_block1) if_block1.c();
			if_block1_anchor = empty();
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			if (if_block0) if_block0.m(target, anchor);
			insert_dev(target, t, anchor);
			if (if_block1) if_block1.m(target, anchor);
			insert_dev(target, if_block1_anchor, anchor);
			current = true;
		},
		p: function update(ctx, [dirty]) {
			if (!/*$loginHelper*/ ctx[0].loggedIn) {
				if (!if_block0) {
					if_block0 = create_if_block_1$2(ctx);
					if_block0.c();
					transition_in(if_block0, 1);
					if_block0.m(t.parentNode, t);
				} else {
					transition_in(if_block0, 1);
				}
			} else if (if_block0) {
				group_outros();

				transition_out(if_block0, 1, 1, () => {
					if_block0 = null;
				});

				check_outros();
			}

			if (/*$loginHelper*/ ctx[0].loggedIn) {
				if (!if_block1) {
					if_block1 = create_if_block$4(ctx);
					if_block1.c();
					if_block1.m(if_block1_anchor.parentNode, if_block1_anchor);
				}
			} else if (if_block1) {
				if_block1.d(1);
				if_block1 = null;
			}
		},
		i: function intro(local) {
			if (current) return;
			transition_in(if_block0);
			current = true;
		},
		o: function outro(local) {
			transition_out(if_block0);
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
		id: create_fragment$8.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$8($$self, $$props, $$invalidate) {
	let $loginHelper;
	validate_store(loginHelper, "loginHelper");
	component_subscribe($$self, loginHelper, $$value => $$invalidate(0, $loginHelper = $$value));
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Login> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Login", $$slots, []);

	$$self.$capture_state = () => ({
		link,
		location: location$1,
		loginHelper,
		Login,
		$loginHelper
	});

	return [$loginHelper];
}

class Login_1 extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$8, create_fragment$8, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Login_1",
			options,
			id: create_fragment$8.name
		});
	}
}

/* src/editor/routes/logout.svelte generated by Svelte v3.20.1 */
const file$8 = "src/editor/routes/logout.svelte";

function create_fragment$9(ctx) {
	let p;
	let t0;
	let t1_value = /*$loginHelper*/ ctx[0].loggedIn + "";
	let t1;

	const block = {
		c: function create() {
			p = element("p");
			t0 = text("You are logged out ");
			t1 = text(t1_value);
			add_location(p, file$8, 4, 0, 70);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, p, anchor);
			append_dev(p, t0);
			append_dev(p, t1);
		},
		p: function update(ctx, [dirty]) {
			if (dirty & /*$loginHelper*/ 1 && t1_value !== (t1_value = /*$loginHelper*/ ctx[0].loggedIn + "")) set_data_dev(t1, t1_value);
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(p);
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
	let $loginHelper;
	validate_store(loginHelper, "loginHelper");
	component_subscribe($$self, loginHelper, $$value => $$invalidate(0, $loginHelper = $$value));
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Logout> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Logout", $$slots, []);
	$$self.$capture_state = () => ({ loginHelper, $loginHelper });
	return [$loginHelper];
}

class Logout extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$9, create_fragment$9, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Logout",
			options,
			id: create_fragment$9.name
		});
	}
}

/* src/editor/routes/not_found.svelte generated by Svelte v3.20.1 */

const file$9 = "src/editor/routes/not_found.svelte";

function create_fragment$a(ctx) {
	let h2;
	let t1;
	let p;

	const block = {
		c: function create() {
			h2 = element("h2");
			h2.textContent = "NotFound";
			t1 = space();
			p = element("p");
			p.textContent = "Oops, this route doesn't exist!";
			attr_dev(h2, "class", "routetitle");
			add_location(h2, file$9, 0, 0, 0);
			add_location(p, file$9, 2, 0, 38);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, h2, anchor);
			insert_dev(target, t1, anchor);
			insert_dev(target, p, anchor);
		},
		p: noop,
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(h2);
			if (detaching) detach_dev(t1);
			if (detaching) detach_dev(p);
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

function instance$a($$self, $$props) {
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Not_found> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Not_found", $$slots, []);
	return [];
}

class Not_found extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$a, create_fragment$a, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Not_found",
			options,
			id: create_fragment$a.name
		});
	}
}

let paths = {
  list: {
    edit: (uuid) => {
      push("/list/edit/" + uuid);
    },
    view: (uuid) => {
      push("/list/view/" + uuid);
    }
  }
};

const current = cache.get(cache.keys["lists.by.me"]);
const { subscribe: subscribe$1, set, update: update$1 } = writable(current);
const loading = writable(false);
const error = writable('');

const ListsByMeStore = () => ({
  subscribe: subscribe$1,
  set,
  loading,
  error,
  async get() {
    let key = cache.keys['lists.by.me'];
    let data = [];
    try {
      data = cache.get(key, data);
      set(data);
      error.set('');
      if (data.length === 0) {
        loading.set(true);
      }

      const response = await getListsByMe();
      loading.set(false);
      cache.save(key, response);
      set(response);
      return response;
    } catch (e) {
      loading.set(false);
      data = cache.get(key, data);
      set(data);
      error.set(`Error has been occurred. Details: ${e.message}`);
    }
  },

  find(uuid) {
    return get_store_value(this).find(aList => {
      return aList.uuid === uuid
    })
  },

  add(aList) {
    update$1(myLists => {
      myLists.push(aList);
      cache.save(cache.keys["lists.by.me"], myLists);
      return myLists;
    });
  },

  update(aList) {
    update$1(myLists => {
      const updated = myLists.map(item => {
        if (item.uuid === aList.uuid) {
          item = aList;
        }
        return item;
      });
      cache.save(cache.keys["lists.by.me"], updated);
      return updated;
    });
  },

  remove(uuid) {
    update$1(myLists => {
      const found = myLists.filter(aList => aList.uuid !== uuid);
      cache.save(cache.keys["lists.by.me"], found);
      return found;
    });
  }
});

var myLists = ListsByMeStore();

const current$1 = cache.get(cache.keys["my.edited.lists"]);
const {subscribe: subscribe$2, set: set$1, update: update$2} = writable(current$1);

const ListsEditsStore = () => ({
    subscribe: subscribe$2,
    set: set$1,

    find(uuid) {
      let found;
      update$2(edits => {
        found = edits.find(aList => {
            return aList.uuid === uuid
        });
        return edits;
      });
      return found;
    },

    add(aList) {
      update$2(edits => {
        const found = edits.some(item => item.uuid === aList.uuid);
        if (!found) {
          edits.push(aList);
        }
        return edits;
      });
    },

    update(aList) {
      update$2(edits => {
        const updated = edits.map(item => {
          if (item.uuid === aList.uuid) {
            item = aList;
          }
          return item;
        });
        cache.save(cache.keys["my.edited.lists"], updated);
        return updated;
      });
    },

    remove(uuid) {
      update$2(edits => {
        const found = edits.filter(aList => aList.uuid !== uuid);
        cache.save(cache.keys["my.edited.lists"], found);
        return found;
      });
    }
});

var listsEdits = ListsEditsStore();

/* src/editor/routes/create_list.svelte generated by Svelte v3.20.1 */
const file$a = "src/editor/routes/create_list.svelte";

function get_each_context(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[9] = list[i];
	child_ctx[11] = i;
	return child_ctx;
}

// (85:2) {#if message}
function create_if_block$5(ctx) {
	let current;

	const errorbox = new Error_box({
			props: { message: /*message*/ ctx[2] },
			$$inline: true
		});

	errorbox.$on("clear", /*clearMessage*/ ctx[4]);

	const block = {
		c: function create() {
			create_component(errorbox.$$.fragment);
		},
		m: function mount(target, anchor) {
			mount_component(errorbox, target, anchor);
			current = true;
		},
		p: function update(ctx, dirty) {
			const errorbox_changes = {};
			if (dirty & /*message*/ 4) errorbox_changes.message = /*message*/ ctx[2];
			errorbox.$set(errorbox_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(errorbox.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(errorbox.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(errorbox, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$5.name,
		type: "if",
		source: "(85:2) {#if message}",
		ctx
	});

	return block;
}

// (106:10) {#each listTypes as listType, pos}
function create_each_block(ctx) {
	let div;
	let input;
	let input_id_value;
	let input_value_value;
	let t0;
	let label;
	let t1_value = /*listType*/ ctx[9].description + "";
	let t1;
	let label_for_value;
	let t2;
	let dispose;

	const block = {
		c: function create() {
			div = element("div");
			input = element("input");
			t0 = space();
			label = element("label");
			t1 = text(t1_value);
			t2 = space();
			attr_dev(input, "class", "mr2");
			attr_dev(input, "type", "radio");
			attr_dev(input, "id", input_id_value = "list-type-" + /*pos*/ ctx[11]);
			attr_dev(input, "name", "type");
			input.__value = input_value_value = /*listType*/ ctx[9].key;
			input.value = input.__value;
			/*$$binding_groups*/ ctx[8][0].push(input);
			add_location(input, file$a, 107, 14, 2583);
			attr_dev(label, "for", label_for_value = "list-type-" + /*pos*/ ctx[11]);
			attr_dev(label, "class", "lh-copy");
			add_location(label, file$a, 115, 14, 2822);
			attr_dev(div, "class", "flex items-center mb2");
			add_location(div, file$a, 106, 12, 2533);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div, anchor);
			append_dev(div, input);
			input.checked = input.__value === /*selected*/ ctx[1];
			append_dev(div, t0);
			append_dev(div, label);
			append_dev(label, t1);
			append_dev(div, t2);
			if (remount) dispose();
			dispose = listen_dev(input, "change", /*input_change_handler*/ ctx[7]);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*selected*/ 2) {
				input.checked = input.__value === /*selected*/ ctx[1];
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div);
			/*$$binding_groups*/ ctx[8][0].splice(/*$$binding_groups*/ ctx[8][0].indexOf(input), 1);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block.name,
		type: "each",
		source: "(106:10) {#each listTypes as listType, pos}",
		ctx
	});

	return block;
}

function create_fragment$b(ctx) {
	let div4;
	let div0;
	let svg;
	let title_1;
	let t0;
	let path;
	let t1;
	let span;
	let t3;
	let t4;
	let section;
	let h1;
	let t6;
	let form;
	let div1;
	let input;
	let init_action;
	let t7;
	let div2;
	let fieldset;
	let t8;
	let div3;
	let button;
	let current;
	let dispose;
	let if_block = /*message*/ ctx[2] && create_if_block$5(ctx);
	let each_value = /*listTypes*/ ctx[3];
	validate_each_argument(each_value);
	let each_blocks = [];

	for (let i = 0; i < each_value.length; i += 1) {
		each_blocks[i] = create_each_block(get_each_context(ctx, each_value, i));
	}

	const block = {
		c: function create() {
			div4 = element("div");
			div0 = element("div");
			svg = svg_element("svg");
			title_1 = svg_element("title");
			t0 = text("info icon");
			path = svg_element("path");
			t1 = space();
			span = element("span");
			span.textContent = "Some info that you want to call attention to.";
			t3 = space();
			if (if_block) if_block.c();
			t4 = space();
			section = element("section");
			h1 = element("h1");
			h1.textContent = "Create a list";
			t6 = space();
			form = element("form");
			div1 = element("div");
			input = element("input");
			t7 = space();
			div2 = element("div");
			fieldset = element("fieldset");

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			t8 = space();
			div3 = element("div");
			button = element("button");
			button.textContent = "Submit";
			add_location(title_1, file$a, 73, 6, 1513);
			attr_dev(path, "d", "M11 15h2v2h-2v-2zm0-8h2v6h-2V7zm.99-5C6.47 2 2 6.48 2 12s4.47 10 9.99\n        10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42\n        0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z");
			add_location(path, file$a, 74, 6, 1544);
			attr_dev(svg, "class", "w1");
			attr_dev(svg, "data-icon", "info");
			attr_dev(svg, "viewBox", "0 0 24 24");
			set_style(svg, "fill", "currentcolor");
			add_location(svg, file$a, 67, 4, 1398);
			attr_dev(span, "class", "lh-title ml3");
			add_location(span, file$a, 80, 4, 1781);
			attr_dev(div0, "class", "flex items-center justify-center pa1 bg-light-red pv3");
			add_location(div0, file$a, 66, 2, 1326);
			attr_dev(h1, "class", "f4 br3 b--yellow black-70 mv0 pv2 ph4");
			add_location(h1, file$a, 89, 4, 2002);
			attr_dev(input, "class", "input-reset ba b--black-20 pa2 mb2 db w-100");
			attr_dev(input, "type", "text");
			attr_dev(input, "aria-describedby", "title-desc");
			attr_dev(input, "placeholder", "Title");
			add_location(input, file$a, 93, 8, 2182);
			attr_dev(div1, "class", "measure");
			add_location(div1, file$a, 92, 6, 2152);
			attr_dev(fieldset, "class", "bn");
			add_location(fieldset, file$a, 104, 8, 2454);
			attr_dev(div2, "class", "measure");
			add_location(div2, file$a, 103, 6, 2424);
			attr_dev(button, "type", "submit");
			add_location(button, file$a, 123, 8, 3036);
			attr_dev(div3, "class", "measure");
			add_location(div3, file$a, 122, 6, 3006);
			attr_dev(form, "class", "pa4 black-80");
			add_location(form, file$a, 91, 4, 2076);
			attr_dev(section, "class", "center pa3 ph1-ns");
			add_location(section, file$a, 88, 2, 1962);
			attr_dev(div4, "class", "pv0 mw100");
			add_location(div4, file$a, 65, 0, 1300);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div4, anchor);
			append_dev(div4, div0);
			append_dev(div0, svg);
			append_dev(svg, title_1);
			append_dev(title_1, t0);
			append_dev(svg, path);
			append_dev(div0, t1);
			append_dev(div0, span);
			append_dev(div4, t3);
			if (if_block) if_block.m(div4, null);
			append_dev(div4, t4);
			append_dev(div4, section);
			append_dev(section, h1);
			append_dev(section, t6);
			append_dev(section, form);
			append_dev(form, div1);
			append_dev(div1, input);
			set_input_value(input, /*title*/ ctx[0]);
			append_dev(form, t7);
			append_dev(form, div2);
			append_dev(div2, fieldset);

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(fieldset, null);
			}

			append_dev(form, t8);
			append_dev(form, div3);
			append_dev(div3, button);
			current = true;
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(input, "input", /*input_input_handler*/ ctx[6]),
				action_destroyer(init_action = init$1.call(null, input)),
				listen_dev(form, "submit", prevent_default(/*handleSubmit*/ ctx[5]), false, true, false)
			];
		},
		p: function update(ctx, [dirty]) {
			if (/*message*/ ctx[2]) {
				if (if_block) {
					if_block.p(ctx, dirty);
					transition_in(if_block, 1);
				} else {
					if_block = create_if_block$5(ctx);
					if_block.c();
					transition_in(if_block, 1);
					if_block.m(div4, t4);
				}
			} else if (if_block) {
				group_outros();

				transition_out(if_block, 1, 1, () => {
					if_block = null;
				});

				check_outros();
			}

			if (dirty & /*title*/ 1 && input.value !== /*title*/ ctx[0]) {
				set_input_value(input, /*title*/ ctx[0]);
			}

			if (dirty & /*listTypes, selected*/ 10) {
				each_value = /*listTypes*/ ctx[3];
				validate_each_argument(each_value);
				let i;

				for (i = 0; i < each_value.length; i += 1) {
					const child_ctx = get_each_context(ctx, each_value, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
					} else {
						each_blocks[i] = create_each_block(child_ctx);
						each_blocks[i].c();
						each_blocks[i].m(fieldset, null);
					}
				}

				for (; i < each_blocks.length; i += 1) {
					each_blocks[i].d(1);
				}

				each_blocks.length = each_value.length;
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
			if (detaching) detach_dev(div4);
			if (if_block) if_block.d();
			destroy_each(each_blocks, detaching);
			run_all(dispose);
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

function init$1(el) {
	el.focus();
}

function instance$b($$self, $$props, $$invalidate) {
	let title = "";

	let listTypes = [
		{ key: "v1", description: "free text" },
		{ key: "v2", description: "From -> To" },
		{
			key: "v4",
			description: "A url and some text"
		},
		{
			key: "v3",
			description: "Concept2 rowing machine log"
		}
	];

	//"v1", "v2", "v4"
	let selected;

	let message;

	function clearMessage() {
		$$invalidate(2, message = null);
	}

	async function handleSubmit() {
		if (title === "") {
			$$invalidate(2, message = "Title cant be empty");
			return;
		}

		if (!selected) {
			$$invalidate(2, message = "Pick a list type");
			return;
		}

		const response = await postList(title, selected);

		if (response.status === 201) {
			const aList = response.body;
			const uuid = aList.uuid;
			listsEdits.add(aList);
			myLists.add(aList);
			paths.list.edit(uuid);
			return;
		}

		$$invalidate(2, message = response.body.message);
	}

	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Create_list> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Create_list", $$slots, []);
	const $$binding_groups = [[]];

	function input_input_handler() {
		title = this.value;
		$$invalidate(0, title);
	}

	function input_change_handler() {
		selected = this.__value;
		$$invalidate(1, selected);
	}

	$$self.$capture_state = () => ({
		goto: paths,
		cache,
		postList,
		myLists,
		listsEdits,
		push,
		ErrorBox: Error_box,
		title,
		listTypes,
		selected,
		message,
		clearMessage,
		handleSubmit,
		init: init$1
	});

	$$self.$inject_state = $$props => {
		if ("title" in $$props) $$invalidate(0, title = $$props.title);
		if ("listTypes" in $$props) $$invalidate(3, listTypes = $$props.listTypes);
		if ("selected" in $$props) $$invalidate(1, selected = $$props.selected);
		if ("message" in $$props) $$invalidate(2, message = $$props.message);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [
		title,
		selected,
		message,
		listTypes,
		clearMessage,
		handleSubmit,
		input_input_handler,
		input_change_handler,
		$$binding_groups
	];
}

class Create_list extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$b, create_fragment$b, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Create_list",
			options,
			id: create_fragment$b.name
		});
	}
}

/* src/editor/routes/create_label.svelte generated by Svelte v3.20.1 */
const file$b = "src/editor/routes/create_label.svelte";

// (46:4) {#if message}
function create_if_block$6(ctx) {
	let current;

	const errorbox = new Error_box({
			props: { message: /*message*/ ctx[1] },
			$$inline: true
		});

	errorbox.$on("clear", /*clearMessage*/ ctx[2]);

	const block = {
		c: function create() {
			create_component(errorbox.$$.fragment);
		},
		m: function mount(target, anchor) {
			mount_component(errorbox, target, anchor);
			current = true;
		},
		p: function update(ctx, dirty) {
			const errorbox_changes = {};
			if (dirty & /*message*/ 2) errorbox_changes.message = /*message*/ ctx[1];
			errorbox.$set(errorbox_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(errorbox.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(errorbox.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(errorbox, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$6.name,
		type: "if",
		source: "(46:4) {#if message}",
		ctx
	});

	return block;
}

function create_fragment$c(ctx) {
	let div3;
	let section;
	let t0;
	let article;
	let h1;
	let t2;
	let div2;
	let form;
	let div0;
	let input;
	let focusThis_action;
	let t3;
	let div1;
	let button;
	let current;
	let dispose;
	let if_block = /*message*/ ctx[1] && create_if_block$6(ctx);

	const block = {
		c: function create() {
			div3 = element("div");
			section = element("section");
			if (if_block) if_block.c();
			t0 = space();
			article = element("article");
			h1 = element("h1");
			h1.textContent = "Create a label";
			t2 = space();
			div2 = element("div");
			form = element("form");
			div0 = element("div");
			input = element("input");
			t3 = space();
			div1 = element("div");
			button = element("button");
			button.textContent = "Submit";
			attr_dev(h1, "class", "f4 br3 b--yellow mw-100 black-70 mv0 pv2 ph4");
			add_location(h1, file$b, 50, 6, 1269);
			attr_dev(input, "class", "input-reset ba b--black-20 pa2 mb2 db w-100");
			attr_dev(input, "type", "text");
			attr_dev(input, "aria-describedby", "title-desc");
			attr_dev(input, "placeholder", "Label");
			add_location(input, file$b, 57, 12, 1525);
			attr_dev(div0, "class", "measure");
			add_location(div0, file$b, 56, 10, 1491);
			attr_dev(button, "type", "submit");
			add_location(button, file$b, 68, 12, 1845);
			attr_dev(div1, "class", "measure");
			add_location(div1, file$b, 67, 10, 1811);
			attr_dev(form, "class", "pa4 black-80");
			add_location(form, file$b, 55, 8, 1411);
			attr_dev(div2, "class", "bt b--washed-yellow");
			add_location(div2, file$b, 54, 6, 1369);
			attr_dev(article, "class", "mw10 bt bw3 b--yellow mw-100");
			add_location(article, file$b, 49, 4, 1216);
			attr_dev(section, "class", "ph0 mh0 pv0");
			add_location(section, file$b, 44, 2, 1098);
			attr_dev(div3, "class", "pv0");
			add_location(div3, file$b, 43, 0, 1078);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div3, anchor);
			append_dev(div3, section);
			if (if_block) if_block.m(section, null);
			append_dev(section, t0);
			append_dev(section, article);
			append_dev(article, h1);
			append_dev(article, t2);
			append_dev(article, div2);
			append_dev(div2, form);
			append_dev(form, div0);
			append_dev(div0, input);
			set_input_value(input, /*newLabel*/ ctx[0]);
			append_dev(form, t3);
			append_dev(form, div1);
			append_dev(div1, button);
			current = true;
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(input, "input", /*input_input_handler*/ ctx[5]),
				action_destroyer(focusThis_action = focusThis.call(null, input)),
				listen_dev(form, "submit", prevent_default(/*handleSubmit*/ ctx[3]), false, true, false)
			];
		},
		p: function update(ctx, [dirty]) {
			if (/*message*/ ctx[1]) {
				if (if_block) {
					if_block.p(ctx, dirty);
					transition_in(if_block, 1);
				} else {
					if_block = create_if_block$6(ctx);
					if_block.c();
					transition_in(if_block, 1);
					if_block.m(section, t0);
				}
			} else if (if_block) {
				group_outros();

				transition_out(if_block, 1, 1, () => {
					if_block = null;
				});

				check_outros();
			}

			if (dirty & /*newLabel*/ 1 && input.value !== /*newLabel*/ ctx[0]) {
				set_input_value(input, /*newLabel*/ ctx[0]);
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
			if (detaching) detach_dev(div3);
			if (if_block) if_block.d();
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
	let newLabel = "";
	let message;

	function clearMessage() {
		$$invalidate(1, message = null);
	}

	async function add() {
		if (newLabel === "" || hasWhiteSpace(newLabel)) {
			$$invalidate(1, message = "The label cannot be empty.");
			$$invalidate(0, newLabel = "");
			return;
		}

		let response = await label.save(newLabel);

		if (response.status === 201 || response.status === 200) {
			$$invalidate(1, message = response.body.message);
			router.showScreenLabelView(newLabel);
		} else {
			$$invalidate(1, message = response.body.message);
		}
	}

	async function handleSubmit() {
		if (newLabel === "" || hasWhiteSpace(newLabel)) {
			$$invalidate(1, message = "The label cannot be empty.");
			$$invalidate(0, newLabel = "");
			return;
		}
	} /*
  let response = await label.save(newLabel);
  if (response.status === 201 || response.status === 200) {
message = response.body.message;
router.showScreenLabelView(newLabel);
  } else {
message = response.body.message;
  }
  */

	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Create_label> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Create_label", $$slots, []);

	function input_input_handler() {
		newLabel = this.value;
		$$invalidate(0, newLabel);
	}

	$$self.$capture_state = () => ({
		hasWhiteSpace,
		focusThis,
		ErrorBox: Error_box,
		newLabel,
		message,
		clearMessage,
		add,
		handleSubmit
	});

	$$self.$inject_state = $$props => {
		if ("newLabel" in $$props) $$invalidate(0, newLabel = $$props.newLabel);
		if ("message" in $$props) $$invalidate(1, message = $$props.message);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [newLabel, message, clearMessage, handleSubmit, add, input_input_handler];
}

class Create_label extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$c, create_fragment$c, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Create_label",
			options,
			id: create_fragment$c.name
		});
	}
}

/* src/editor/routes/create.svelte generated by Svelte v3.20.1 */
const file$c = "src/editor/routes/create.svelte";

function create_fragment$d(ctx) {
	let div1;
	let article;
	let h1;
	let t1;
	let div0;
	let a0;
	let link_action;
	let t3;
	let a1;
	let link_action_1;
	let dispose;

	const block = {
		c: function create() {
			div1 = element("div");
			article = element("article");
			h1 = element("h1");
			h1.textContent = "I would like to create a";
			t1 = space();
			div0 = element("div");
			a0 = element("a");
			a0.textContent = "List";
			t3 = space();
			a1 = element("a");
			a1.textContent = "Label";
			attr_dev(h1, "class", "fw3 f3 f2-ns lh-title mt0 mb3");
			add_location(h1, file$c, 9, 4, 188);
			attr_dev(a0, "class", "f6 link dim br1 ba bw1 ph3 pv2 mb2 dib black mr6-ns");
			attr_dev(a0, "href", "/create/list");
			add_location(a0, file$c, 11, 6, 285);
			attr_dev(a1, "class", "f6 link dim br1 ba bw1 ph3 pv2 mb2 dib black");
			attr_dev(a1, "href", "/create/label");
			add_location(a1, file$c, 19, 6, 440);
			attr_dev(div0, "class", "");
			add_location(div0, file$c, 10, 4, 264);
			attr_dev(article, "class", "pt5 mw10 center ph3 ph5-ns tc br2 pv5 bg-washed-yellow dark-yellow\n    mb5");
			add_location(article, file$c, 5, 2, 84);
			attr_dev(div1, "class", "pt2");
			add_location(div1, file$c, 4, 0, 64);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div1, anchor);
			append_dev(div1, article);
			append_dev(article, h1);
			append_dev(article, t1);
			append_dev(article, div0);
			append_dev(div0, a0);
			append_dev(div0, t3);
			append_dev(div0, a1);
			if (remount) run_all(dispose);

			dispose = [
				action_destroyer(link_action = link.call(null, a0)),
				action_destroyer(link_action_1 = link.call(null, a1))
			];
		},
		p: noop,
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(div1);
			run_all(dispose);
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
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Create> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Create", $$slots, []);
	$$self.$capture_state = () => ({ link });
	return [];
}

class Create extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$d, create_fragment$d, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Create",
			options,
			id: create_fragment$d.name
		});
	}
}

/* src/editor/components/list_find_item.svelte generated by Svelte v3.20.1 */

const { console: console_1$1 } = globals;
const file$d = "src/editor/components/list_find_item.svelte";

function create_fragment$e(ctx) {
	let li;
	let t;
	let dispose;

	const block = {
		c: function create() {
			li = element("li");
			t = text(/*title*/ ctx[0]);
			attr_dev(li, "class", "lh-copy pv3 ba bl-0 bt-0 br-0 b--dotted b--black-30");
			add_location(li, file$d, 11, 0, 187);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, li, anchor);
			append_dev(li, t);
			if (remount) dispose();
			dispose = listen_dev(li, "click", /*handleClick*/ ctx[1], false, false, false);
		},
		p: function update(ctx, [dirty]) {
			if (dirty & /*title*/ 1) set_data_dev(t, /*title*/ ctx[0]);
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(li);
			dispose();
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

function instance$e($$self, $$props, $$invalidate) {
	let { title = "" } = $$props;
	let { uuid = "" } = $$props;

	function handleClick() {
		console.log(uuid);
		paths.list.view(uuid);
	}

	const writable_props = ["title", "uuid"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$1.warn(`<List_find_item> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_find_item", $$slots, []);

	$$self.$set = $$props => {
		if ("title" in $$props) $$invalidate(0, title = $$props.title);
		if ("uuid" in $$props) $$invalidate(2, uuid = $$props.uuid);
	};

	$$self.$capture_state = () => ({ goto: paths, title, uuid, handleClick });

	$$self.$inject_state = $$props => {
		if ("title" in $$props) $$invalidate(0, title = $$props.title);
		if ("uuid" in $$props) $$invalidate(2, uuid = $$props.uuid);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [title, handleClick, uuid];
}

class List_find_item extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$e, create_fragment$e, safe_not_equal, { title: 0, uuid: 2 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_find_item",
			options,
			id: create_fragment$e.name
		});
	}

	get title() {
		throw new Error("<List_find_item>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set title(value) {
		throw new Error("<List_find_item>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	get uuid() {
		throw new Error("<List_find_item>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set uuid(value) {
		throw new Error("<List_find_item>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/routes/list_find.svelte generated by Svelte v3.20.1 */
const file$e = "src/editor/routes/list_find.svelte";

function get_each_context$1(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[18] = list[i];
	return child_ctx;
}

function get_each_context_1(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[21] = list[i];
	return child_ctx;
}

function get_each_context_2(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[24] = list[i];
	return child_ctx;
}

// (111:4) {:else}
function create_else_block$2(ctx) {
	let div3;
	let fieldset;
	let div0;
	let span;
	let t1;
	let div1;
	let select;
	let t2;
	let show_if = /*hasLabels*/ ctx[9](/*listLabels*/ ctx[4]);
	let t3;
	let div2;
	let button;
	let t5;
	let ul;
	let current;
	let dispose;
	let each_value_2 = /*listTypes*/ ctx[3];
	validate_each_argument(each_value_2);
	let each_blocks_1 = [];

	for (let i = 0; i < each_value_2.length; i += 1) {
		each_blocks_1[i] = create_each_block_2(get_each_context_2(ctx, each_value_2, i));
	}

	let if_block = show_if && create_if_block_2(ctx);
	let each_value = /*filterLists*/ ctx[2];
	validate_each_argument(each_value);
	let each_blocks = [];

	for (let i = 0; i < each_value.length; i += 1) {
		each_blocks[i] = create_each_block$1(get_each_context$1(ctx, each_value, i));
	}

	const out = i => transition_out(each_blocks[i], 1, 1, () => {
		each_blocks[i] = null;
	});

	const block = {
		c: function create() {
			div3 = element("div");
			fieldset = element("fieldset");
			div0 = element("div");
			span = element("span");
			span.textContent = "Filter";
			t1 = space();
			div1 = element("div");
			select = element("select");

			for (let i = 0; i < each_blocks_1.length; i += 1) {
				each_blocks_1[i].c();
			}

			t2 = space();
			if (if_block) if_block.c();
			t3 = space();
			div2 = element("div");
			button = element("button");
			button.textContent = "reset";
			t5 = space();
			ul = element("ul");

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			add_location(span, file$e, 114, 12, 2586);
			attr_dev(div0, "class", "flex items-center mb2");
			add_location(div0, file$e, 113, 10, 2538);
			if (/*find*/ ctx[0] === void 0) add_render_callback(() => /*select_change_handler*/ ctx[16].call(select));
			add_location(select, file$e, 118, 12, 2682);
			attr_dev(div1, "class", "flex items-center mb2");
			add_location(div1, file$e, 117, 10, 2634);
			add_location(button, file$e, 136, 12, 3279);
			attr_dev(div2, "class", "flex items-center mb2");
			add_location(div2, file$e, 135, 10, 3231);
			attr_dev(fieldset, "class", "bn");
			add_location(fieldset, file$e, 112, 8, 2506);
			add_location(div3, file$e, 111, 6, 2492);
			attr_dev(ul, "class", "list pl0 measure center");
			add_location(ul, file$e, 140, 6, 3377);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div3, anchor);
			append_dev(div3, fieldset);
			append_dev(fieldset, div0);
			append_dev(div0, span);
			append_dev(fieldset, t1);
			append_dev(fieldset, div1);
			append_dev(div1, select);

			for (let i = 0; i < each_blocks_1.length; i += 1) {
				each_blocks_1[i].m(select, null);
			}

			select_option(select, /*find*/ ctx[0]);
			append_dev(fieldset, t2);
			if (if_block) if_block.m(fieldset, null);
			append_dev(fieldset, t3);
			append_dev(fieldset, div2);
			append_dev(div2, button);
			insert_dev(target, t5, anchor);
			insert_dev(target, ul, anchor);

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(ul, null);
			}

			current = true;
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(select, "change", /*select_change_handler*/ ctx[16]),
				listen_dev(button, "click", /*reset*/ ctx[10], false, false, false)
			];
		},
		p: function update(ctx, dirty) {
			if (dirty & /*listTypes*/ 8) {
				each_value_2 = /*listTypes*/ ctx[3];
				validate_each_argument(each_value_2);
				let i;

				for (i = 0; i < each_value_2.length; i += 1) {
					const child_ctx = get_each_context_2(ctx, each_value_2, i);

					if (each_blocks_1[i]) {
						each_blocks_1[i].p(child_ctx, dirty);
					} else {
						each_blocks_1[i] = create_each_block_2(child_ctx);
						each_blocks_1[i].c();
						each_blocks_1[i].m(select, null);
					}
				}

				for (; i < each_blocks_1.length; i += 1) {
					each_blocks_1[i].d(1);
				}

				each_blocks_1.length = each_value_2.length;
			}

			if (dirty & /*find*/ 1) {
				select_option(select, /*find*/ ctx[0]);
			}

			if (dirty & /*listLabels*/ 16) show_if = /*hasLabels*/ ctx[9](/*listLabels*/ ctx[4]);

			if (show_if) {
				if (if_block) {
					if_block.p(ctx, dirty);
				} else {
					if_block = create_if_block_2(ctx);
					if_block.c();
					if_block.m(fieldset, t3);
				}
			} else if (if_block) {
				if_block.d(1);
				if_block = null;
			}

			if (dirty & /*filterLists*/ 4) {
				each_value = /*filterLists*/ ctx[2];
				validate_each_argument(each_value);
				let i;

				for (i = 0; i < each_value.length; i += 1) {
					const child_ctx = get_each_context$1(ctx, each_value, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
						transition_in(each_blocks[i], 1);
					} else {
						each_blocks[i] = create_each_block$1(child_ctx);
						each_blocks[i].c();
						transition_in(each_blocks[i], 1);
						each_blocks[i].m(ul, null);
					}
				}

				group_outros();

				for (i = each_value.length; i < each_blocks.length; i += 1) {
					out(i);
				}

				check_outros();
			}
		},
		i: function intro(local) {
			if (current) return;

			for (let i = 0; i < each_value.length; i += 1) {
				transition_in(each_blocks[i]);
			}

			current = true;
		},
		o: function outro(local) {
			each_blocks = each_blocks.filter(Boolean);

			for (let i = 0; i < each_blocks.length; i += 1) {
				transition_out(each_blocks[i]);
			}

			current = false;
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div3);
			destroy_each(each_blocks_1, detaching);
			if (if_block) if_block.d();
			if (detaching) detach_dev(t5);
			if (detaching) detach_dev(ul);
			destroy_each(each_blocks, detaching);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_else_block$2.name,
		type: "else",
		source: "(111:4) {:else}",
		ctx
	});

	return block;
}

// (109:23) 
function create_if_block_1$3(ctx) {
	let t;

	const block = {
		c: function create() {
			t = text("Loading...");
		},
		m: function mount(target, anchor) {
			insert_dev(target, t, anchor);
		},
		p: noop,
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(t);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_1$3.name,
		type: "if",
		source: "(109:23) ",
		ctx
	});

	return block;
}

// (107:4) {#if $error}
function create_if_block$7(ctx) {
	let t0;
	let t1;

	const block = {
		c: function create() {
			t0 = text("error is ");
			t1 = text(/*$error*/ ctx[5]);
		},
		m: function mount(target, anchor) {
			insert_dev(target, t0, anchor);
			insert_dev(target, t1, anchor);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*$error*/ 32) set_data_dev(t1, /*$error*/ ctx[5]);
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(t0);
			if (detaching) detach_dev(t1);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$7.name,
		type: "if",
		source: "(107:4) {#if $error}",
		ctx
	});

	return block;
}

// (120:14) {#each listTypes as listType}
function create_each_block_2(ctx) {
	let option;
	let t_value = /*listType*/ ctx[24].description + "";
	let t;
	let option_value_value;

	const block = {
		c: function create() {
			option = element("option");
			t = text(t_value);
			option.__value = option_value_value = /*listType*/ ctx[24].key;
			option.value = option.__value;
			add_location(option, file$e, 120, 16, 2771);
		},
		m: function mount(target, anchor) {
			insert_dev(target, option, anchor);
			append_dev(option, t);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*listTypes*/ 8 && t_value !== (t_value = /*listType*/ ctx[24].description + "")) set_data_dev(t, t_value);

			if (dirty & /*listTypes*/ 8 && option_value_value !== (option_value_value = /*listType*/ ctx[24].key)) {
				prop_dev(option, "__value", option_value_value);
			}

			option.value = option.__value;
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(option);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block_2.name,
		type: "each",
		source: "(120:14) {#each listTypes as listType}",
		ctx
	});

	return block;
}

// (126:10) {#if hasLabels(listLabels)}
function create_if_block_2(ctx) {
	let div;
	let select;
	let dispose;
	let each_value_1 = /*listLabels*/ ctx[4];
	validate_each_argument(each_value_1);
	let each_blocks = [];

	for (let i = 0; i < each_value_1.length; i += 1) {
		each_blocks[i] = create_each_block_1(get_each_context_1(ctx, each_value_1, i));
	}

	const block = {
		c: function create() {
			div = element("div");
			select = element("select");

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			if (/*filterByLabel*/ ctx[1] === void 0) add_render_callback(() => /*select_change_handler_1*/ ctx[17].call(select));
			add_location(select, file$e, 127, 14, 2996);
			attr_dev(div, "class", "flex items-center mb2");
			add_location(div, file$e, 126, 12, 2946);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div, anchor);
			append_dev(div, select);

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(select, null);
			}

			select_option(select, /*filterByLabel*/ ctx[1]);
			if (remount) dispose();
			dispose = listen_dev(select, "change", /*select_change_handler_1*/ ctx[17]);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*listLabels*/ 16) {
				each_value_1 = /*listLabels*/ ctx[4];
				validate_each_argument(each_value_1);
				let i;

				for (i = 0; i < each_value_1.length; i += 1) {
					const child_ctx = get_each_context_1(ctx, each_value_1, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
					} else {
						each_blocks[i] = create_each_block_1(child_ctx);
						each_blocks[i].c();
						each_blocks[i].m(select, null);
					}
				}

				for (; i < each_blocks.length; i += 1) {
					each_blocks[i].d(1);
				}

				each_blocks.length = each_value_1.length;
			}

			if (dirty & /*filterByLabel*/ 2) {
				select_option(select, /*filterByLabel*/ ctx[1]);
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div);
			destroy_each(each_blocks, detaching);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_2.name,
		type: "if",
		source: "(126:10) {#if hasLabels(listLabels)}",
		ctx
	});

	return block;
}

// (129:16) {#each listLabels as label}
function create_each_block_1(ctx) {
	let option;
	let t_value = /*label*/ ctx[21] + "";
	let t;
	let option_value_value;

	const block = {
		c: function create() {
			option = element("option");
			t = text(t_value);
			option.__value = option_value_value = /*label*/ ctx[21];
			option.value = option.__value;
			add_location(option, file$e, 129, 18, 3096);
		},
		m: function mount(target, anchor) {
			insert_dev(target, option, anchor);
			append_dev(option, t);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*listLabels*/ 16 && t_value !== (t_value = /*label*/ ctx[21] + "")) set_data_dev(t, t_value);

			if (dirty & /*listLabels*/ 16 && option_value_value !== (option_value_value = /*label*/ ctx[21])) {
				prop_dev(option, "__value", option_value_value);
			}

			option.value = option.__value;
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(option);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block_1.name,
		type: "each",
		source: "(129:16) {#each listLabels as label}",
		ctx
	});

	return block;
}

// (142:8) {#each filterLists as aList}
function create_each_block$1(ctx) {
	let current;

	const listitem = new List_find_item({
			props: {
				title: /*aList*/ ctx[18].info.title,
				uuid: /*aList*/ ctx[18].uuid
			},
			$$inline: true
		});

	const block = {
		c: function create() {
			create_component(listitem.$$.fragment);
		},
		m: function mount(target, anchor) {
			mount_component(listitem, target, anchor);
			current = true;
		},
		p: function update(ctx, dirty) {
			const listitem_changes = {};
			if (dirty & /*filterLists*/ 4) listitem_changes.title = /*aList*/ ctx[18].info.title;
			if (dirty & /*filterLists*/ 4) listitem_changes.uuid = /*aList*/ ctx[18].uuid;
			listitem.$set(listitem_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(listitem.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(listitem.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(listitem, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block$1.name,
		type: "each",
		source: "(142:8) {#each filterLists as aList}",
		ctx
	});

	return block;
}

function create_fragment$f(ctx) {
	let div1;
	let div0;
	let current_block_type_index;
	let if_block;
	let current;
	const if_block_creators = [create_if_block$7, create_if_block_1$3, create_else_block$2];
	const if_blocks = [];

	function select_block_type(ctx, dirty) {
		if (/*$error*/ ctx[5]) return 0;
		if (/*$loading*/ ctx[6]) return 1;
		return 2;
	}

	current_block_type_index = select_block_type(ctx);
	if_block = if_blocks[current_block_type_index] = if_block_creators[current_block_type_index](ctx);

	const block = {
		c: function create() {
			div1 = element("div");
			div0 = element("div");
			if_block.c();
			attr_dev(div0, "class", "pl0 measure center");
			add_location(div0, file$e, 105, 2, 2359);
			attr_dev(div1, "class", "pa3 pa2-ns");
			add_location(div1, file$e, 103, 0, 2331);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, div1, anchor);
			append_dev(div1, div0);
			if_blocks[current_block_type_index].m(div0, null);
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
				if_block.m(div0, null);
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
			if (detaching) detach_dev(div1);
			if_blocks[current_block_type_index].d();
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

function getSelectListLabels(lists) {
	const labels = new Set();

	lists.forEach(item => {
		if (item.info.labels.length > 0) {
			item.info.labels.forEach(label => {
				labels.add(label);
			});
		}
	});

	if (labels.size === 0) {
		return [];
	}

	return ["Any label", ...labels];
}

function filterListsByFilters(lists, find, filterByLabel) {
	let filtered = lists.filter(item => {
		if (find == "all") {
			return true;
		}

		return find == item.info.type;
	});

	filtered = filtered.filter(item => {
		if (filterByLabel == "Any label") {
			return true;
		}

		return item.info.labels.includes(filterByLabel);
	});

	return filtered;
}

function instance$f($$self, $$props, $$invalidate) {
	let $ListsByMeStore;
	let $error;
	let $loading;
	validate_store(myLists, "ListsByMeStore");
	component_subscribe($$self, myLists, $$value => $$invalidate(11, $ListsByMeStore = $$value));
	const loading = myLists.loading;
	validate_store(loading, "loading");
	component_subscribe($$self, loading, value => $$invalidate(6, $loading = value));
	const error = myLists.error;
	validate_store(error, "error");
	component_subscribe($$self, error, value => $$invalidate(5, $error = value));
	myLists.get();
	let find = "all";
	let filterByLabel = "Any label";
	let lists = $ListsByMeStore;
	const foundListTypes = [...new Set($ListsByMeStore.map(item => item.info.type))];

	const defaultListTypes = [
		{ key: "all", description: "Any list type" },
		{ key: "v1", description: "free text" },
		{ key: "v2", description: "From -> To" },
		{
			key: "v4",
			description: "A url and some text"
		},
		{
			key: "v3",
			description: "Concept2 rowing machine log"
		}
	];

	function getSelectListTypes(lists) {
		const foundListTypes = [...new Set(lists.map(item => item.info.type))];
		const listTypes = [];

		// Add the all option
		listTypes.push(defaultListTypes[0]);

		const filtered = new Set(defaultListTypes.filter(item => foundListTypes.includes(item.key)));

		filtered.forEach(e => {
			listTypes.push(e);
		});

		return listTypes;
	}

	function hasLabels(listLabels) {
		return !!listLabels.length;
	}

	function reset() {
		$$invalidate(0, find = "all");
		$$invalidate(1, filterByLabel = "Any label");
	}

	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_find> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_find", $$slots, []);

	function select_change_handler() {
		find = select_value(this);
		$$invalidate(0, find);
		(((($$invalidate(3, listTypes), $$invalidate(2, filterLists)), $$invalidate(11, $ListsByMeStore)), $$invalidate(0, find)), $$invalidate(1, filterByLabel));
	}

	function select_change_handler_1() {
		filterByLabel = select_value(this);
		$$invalidate(1, filterByLabel);
		(((($$invalidate(4, listLabels), $$invalidate(2, filterLists)), $$invalidate(11, $ListsByMeStore)), $$invalidate(0, find)), $$invalidate(1, filterByLabel));
	}

	$$self.$capture_state = () => ({
		ErrorBox: Error_box,
		ListItem: List_find_item,
		ListsByMeStore: myLists,
		loading,
		error,
		find,
		filterByLabel,
		lists,
		foundListTypes,
		defaultListTypes,
		getSelectListTypes,
		getSelectListLabels,
		filterListsByFilters,
		hasLabels,
		reset,
		$ListsByMeStore,
		filterLists,
		listTypes,
		listLabels,
		$error,
		$loading
	});

	$$self.$inject_state = $$props => {
		if ("find" in $$props) $$invalidate(0, find = $$props.find);
		if ("filterByLabel" in $$props) $$invalidate(1, filterByLabel = $$props.filterByLabel);
		if ("lists" in $$props) lists = $$props.lists;
		if ("filterLists" in $$props) $$invalidate(2, filterLists = $$props.filterLists);
		if ("listTypes" in $$props) $$invalidate(3, listTypes = $$props.listTypes);
		if ("listLabels" in $$props) $$invalidate(4, listLabels = $$props.listLabels);
	};

	let filterLists;
	let listTypes;
	let listLabels;

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	$$self.$$.update = () => {
		if ($$self.$$.dirty & /*$ListsByMeStore, find, filterByLabel*/ 2051) {
			 $$invalidate(2, filterLists = filterListsByFilters($ListsByMeStore, find, filterByLabel));
		}

		if ($$self.$$.dirty & /*filterLists*/ 4) {
			 $$invalidate(3, listTypes = getSelectListTypes(filterLists));
		}

		if ($$self.$$.dirty & /*filterLists*/ 4) {
			 $$invalidate(4, listLabels = getSelectListLabels(filterLists));
		}
	};

	return [
		find,
		filterByLabel,
		filterLists,
		listTypes,
		listLabels,
		$error,
		$loading,
		loading,
		error,
		hasLabels,
		reset,
		$ListsByMeStore,
		lists,
		foundListTypes,
		defaultListTypes,
		getSelectListTypes,
		select_change_handler,
		select_change_handler_1
	];
}

class List_find extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$f, create_fragment$f, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_find",
			options,
			id: create_fragment$f.name
		});
	}
}

/* src/editor/components/list.view.data.item.v1.svelte generated by Svelte v3.20.1 */

const file$f = "src/editor/components/list.view.data.item.v1.svelte";

function create_fragment$g(ctx) {
	let article;
	let div;
	let h1;
	let t;

	const block = {
		c: function create() {
			article = element("article");
			div = element("div");
			h1 = element("h1");
			t = text(/*item*/ ctx[0]);
			attr_dev(h1, "class", "f6 f5-ns fw5 lh-title black-60 mv0");
			add_location(h1, file$f, 6, 4, 128);
			attr_dev(div, "class", "dtc v-mid pl0");
			add_location(div, file$f, 5, 2, 96);
			attr_dev(article, "class", "dt w-100 bb b--black-05 pb2 mt2");
			add_location(article, file$f, 4, 0, 44);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, article, anchor);
			append_dev(article, div);
			append_dev(div, h1);
			append_dev(h1, t);
		},
		p: function update(ctx, [dirty]) {
			if (dirty & /*item*/ 1) set_data_dev(t, /*item*/ ctx[0]);
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(article);
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
	let { item = "" } = $$props;
	const writable_props = ["item"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_view_data_item_v1> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_view_data_item_v1", $$slots, []);

	$$self.$set = $$props => {
		if ("item" in $$props) $$invalidate(0, item = $$props.item);
	};

	$$self.$capture_state = () => ({ item });

	$$self.$inject_state = $$props => {
		if ("item" in $$props) $$invalidate(0, item = $$props.item);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [item];
}

class List_view_data_item_v1 extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$g, create_fragment$g, safe_not_equal, { item: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_view_data_item_v1",
			options,
			id: create_fragment$g.name
		});
	}

	get item() {
		throw new Error("<List_view_data_item_v1>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set item(value) {
		throw new Error("<List_view_data_item_v1>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list.view.data.item.v2.svelte generated by Svelte v3.20.1 */

const file$g = "src/editor/components/list.view.data.item.v2.svelte";

function create_fragment$h(ctx) {
	let article;
	let div;
	let h1;
	let t0_value = /*item*/ ctx[0].from + "";
	let t0;
	let t1;
	let h2;
	let t2_value = /*item*/ ctx[0].to + "";
	let t2;

	const block = {
		c: function create() {
			article = element("article");
			div = element("div");
			h1 = element("h1");
			t0 = text(t0_value);
			t1 = space();
			h2 = element("h2");
			t2 = text(t2_value);
			attr_dev(h1, "class", "f6 f5-ns fw6 lh-title black mv0");
			add_location(h1, file$g, 9, 4, 156);
			attr_dev(h2, "class", "f6 fw4 mt0 mb0 black-60");
			add_location(h2, file$g, 10, 4, 221);
			attr_dev(div, "class", "dtc v-mid pl0");
			add_location(div, file$g, 8, 2, 124);
			attr_dev(article, "class", "dt w-100 bb b--black-05 pb2 mt2");
			add_location(article, file$g, 7, 0, 72);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, article, anchor);
			append_dev(article, div);
			append_dev(div, h1);
			append_dev(h1, t0);
			append_dev(div, t1);
			append_dev(div, h2);
			append_dev(h2, t2);
		},
		p: function update(ctx, [dirty]) {
			if (dirty & /*item*/ 1 && t0_value !== (t0_value = /*item*/ ctx[0].from + "")) set_data_dev(t0, t0_value);
			if (dirty & /*item*/ 1 && t2_value !== (t2_value = /*item*/ ctx[0].to + "")) set_data_dev(t2, t2_value);
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(article);
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
	let { item = { from: "", to: "" } } = $$props;
	const writable_props = ["item"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_view_data_item_v2> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_view_data_item_v2", $$slots, []);

	$$self.$set = $$props => {
		if ("item" in $$props) $$invalidate(0, item = $$props.item);
	};

	$$self.$capture_state = () => ({ item });

	$$self.$inject_state = $$props => {
		if ("item" in $$props) $$invalidate(0, item = $$props.item);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [item];
}

class List_view_data_item_v2 extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$h, create_fragment$h, safe_not_equal, { item: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_view_data_item_v2",
			options,
			id: create_fragment$h.name
		});
	}

	get item() {
		throw new Error("<List_view_data_item_v2>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set item(value) {
		throw new Error("<List_view_data_item_v2>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list.view.data.item.v3.svelte generated by Svelte v3.20.1 */

const file$h = "src/editor/components/list.view.data.item.v3.svelte";

function create_fragment$i(ctx) {
	let article;
	let div;
	let h1;
	let t;

	const block = {
		c: function create() {
			article = element("article");
			div = element("div");
			h1 = element("h1");
			t = text(/*description*/ ctx[0]);
			attr_dev(h1, "class", "f6 f5-ns fw6 lh-title black mv0");
			add_location(h1, file$h, 12, 4, 299);
			attr_dev(div, "class", "dtc v-mid pl0");
			add_location(div, file$h, 11, 2, 267);
			attr_dev(article, "class", "dt w-100 bb b--black-05 pb2 mt2");
			add_location(article, file$h, 10, 0, 215);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, article, anchor);
			append_dev(article, div);
			append_dev(div, h1);
			append_dev(h1, t);
		},
		p: function update(ctx, [dirty]) {
			if (dirty & /*description*/ 1) set_data_dev(t, /*description*/ ctx[0]);
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(article);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$i.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$i($$self, $$props, $$invalidate) {
	let { item = {
		when: "",
		overall: {
			time: "",
			distance: "",
			p500: "",
			spm: ""
		},
		splits: []
	} } = $$props;

	const writable_props = ["item"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_view_data_item_v3> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_view_data_item_v3", $$slots, []);

	$$self.$set = $$props => {
		if ("item" in $$props) $$invalidate(1, item = $$props.item);
	};

	$$self.$capture_state = () => ({ item, description });

	$$self.$inject_state = $$props => {
		if ("item" in $$props) $$invalidate(1, item = $$props.item);
		if ("description" in $$props) $$invalidate(0, description = $$props.description);
	};

	let description;

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	$$self.$$.update = () => {
		if ($$self.$$.dirty & /*item*/ 2) {
			 $$invalidate(0, description = `${item.overall.distance} meters in ${item.overall.time}`);
		}
	};

	return [description, item];
}

class List_view_data_item_v3 extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$i, create_fragment$i, safe_not_equal, { item: 1 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_view_data_item_v3",
			options,
			id: create_fragment$i.name
		});
	}

	get item() {
		throw new Error("<List_view_data_item_v3>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set item(value) {
		throw new Error("<List_view_data_item_v3>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list.view.data.item.v4.svelte generated by Svelte v3.20.1 */

const file$i = "src/editor/components/list.view.data.item.v4.svelte";

function create_fragment$j(ctx) {
	let article;
	let div;
	let h1;
	let t0_value = /*item*/ ctx[0].content + "";
	let t0;
	let t1;
	let a;
	let t2_value = /*item*/ ctx[0].url + "";
	let t2;
	let a_href_value;

	const block = {
		c: function create() {
			article = element("article");
			div = element("div");
			h1 = element("h1");
			t0 = text(t0_value);
			t1 = space();
			a = element("a");
			t2 = text(t2_value);
			attr_dev(h1, "class", "f6 f5-ns fw6 lh-title black mv0");
			add_location(h1, file$i, 9, 4, 160);
			attr_dev(a, "class", "f6 fw4 mt0 mb0 black-40");
			attr_dev(a, "href", a_href_value = /*item*/ ctx[0].url);
			attr_dev(a, "target", "_blank");
			add_location(a, file$i, 10, 4, 228);
			attr_dev(div, "class", "dtc v-mid pl0");
			add_location(div, file$i, 8, 2, 128);
			attr_dev(article, "class", "dt w-100 bb b--black-05 pb2 mt2");
			add_location(article, file$i, 7, 0, 76);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, article, anchor);
			append_dev(article, div);
			append_dev(div, h1);
			append_dev(h1, t0);
			append_dev(div, t1);
			append_dev(div, a);
			append_dev(a, t2);
		},
		p: function update(ctx, [dirty]) {
			if (dirty & /*item*/ 1 && t0_value !== (t0_value = /*item*/ ctx[0].content + "")) set_data_dev(t0, t0_value);
			if (dirty & /*item*/ 1 && t2_value !== (t2_value = /*item*/ ctx[0].url + "")) set_data_dev(t2, t2_value);

			if (dirty & /*item*/ 1 && a_href_value !== (a_href_value = /*item*/ ctx[0].url)) {
				attr_dev(a, "href", a_href_value);
			}
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(article);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$j.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$j($$self, $$props, $$invalidate) {
	let { item = { content: "", url: "" } } = $$props;
	const writable_props = ["item"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_view_data_item_v4> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_view_data_item_v4", $$slots, []);

	$$self.$set = $$props => {
		if ("item" in $$props) $$invalidate(0, item = $$props.item);
	};

	$$self.$capture_state = () => ({ item });

	$$self.$inject_state = $$props => {
		if ("item" in $$props) $$invalidate(0, item = $$props.item);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [item];
}

class List_view_data_item_v4 extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$j, create_fragment$j, safe_not_equal, { item: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_view_data_item_v4",
			options,
			id: create_fragment$j.name
		});
	}

	get item() {
		throw new Error("<List_view_data_item_v4>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set item(value) {
		throw new Error("<List_view_data_item_v4>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list.view.svelte generated by Svelte v3.20.1 */
const file$j = "src/editor/components/list.view.svelte";

function get_each_context$2(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[12] = list[i];
	child_ctx[13] = list;
	child_ctx[14] = i;
	return child_ctx;
}

function get_each_context_1$1(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[12] = list[i];
	return child_ctx;
}

// (61:4) {#if labels.length > 0}
function create_if_block_1$4(ctx) {
	let div;
	let ul;
	let each_value_1 = /*labels*/ ctx[4];
	validate_each_argument(each_value_1);
	let each_blocks = [];

	for (let i = 0; i < each_value_1.length; i += 1) {
		each_blocks[i] = create_each_block_1$1(get_each_context_1$1(ctx, each_value_1, i));
	}

	const block = {
		c: function create() {
			div = element("div");
			ul = element("ul");

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			attr_dev(ul, "class", "list pl0");
			add_location(ul, file$j, 62, 8, 1442);
			attr_dev(div, "class", "nicebox");
			add_location(div, file$j, 61, 6, 1412);
		},
		m: function mount(target, anchor) {
			insert_dev(target, div, anchor);
			append_dev(div, ul);

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(ul, null);
			}
		},
		p: function update(ctx, dirty) {
			if (dirty & /*labels*/ 16) {
				each_value_1 = /*labels*/ ctx[4];
				validate_each_argument(each_value_1);
				let i;

				for (i = 0; i < each_value_1.length; i += 1) {
					const child_ctx = get_each_context_1$1(ctx, each_value_1, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
					} else {
						each_blocks[i] = create_each_block_1$1(child_ctx);
						each_blocks[i].c();
						each_blocks[i].m(ul, null);
					}
				}

				for (; i < each_blocks.length; i += 1) {
					each_blocks[i].d(1);
				}

				each_blocks.length = each_value_1.length;
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div);
			destroy_each(each_blocks, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_1$4.name,
		type: "if",
		source: "(61:4) {#if labels.length > 0}",
		ctx
	});

	return block;
}

// (65:10) {#each labels as item}
function create_each_block_1$1(ctx) {
	let li;
	let span;
	let t0_value = /*item*/ ctx[12] + "";
	let t0;
	let t1;

	const block = {
		c: function create() {
			li = element("li");
			span = element("span");
			t0 = text(t0_value);
			t1 = space();
			attr_dev(span, "href", "#");
			attr_dev(span, "class", "f6 f5-ns b db pa2 link dark-gray ba b--black-20");
			add_location(span, file$j, 66, 14, 1553);
			attr_dev(li, "class", "dib mr1 mb2 pl0");
			add_location(li, file$j, 65, 12, 1510);
		},
		m: function mount(target, anchor) {
			insert_dev(target, li, anchor);
			append_dev(li, span);
			append_dev(span, t0);
			append_dev(li, t1);
		},
		p: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(li);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block_1$1.name,
		type: "each",
		source: "(65:10) {#each labels as item}",
		ctx
	});

	return block;
}

// (79:4) {#if data.length > 0}
function create_if_block$8(ctx) {
	let div;
	let current;
	let each_value = /*data*/ ctx[0];
	validate_each_argument(each_value);
	let each_blocks = [];

	for (let i = 0; i < each_value.length; i += 1) {
		each_blocks[i] = create_each_block$2(get_each_context$2(ctx, each_value, i));
	}

	const out = i => transition_out(each_blocks[i], 1, 1, () => {
		each_blocks[i] = null;
	});

	const block = {
		c: function create() {
			div = element("div");

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			attr_dev(div, "class", "nicebox");
			add_location(div, file$j, 79, 6, 1823);
		},
		m: function mount(target, anchor) {
			insert_dev(target, div, anchor);

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(div, null);
			}

			current = true;
		},
		p: function update(ctx, dirty) {
			if (dirty & /*renderItem, data*/ 33) {
				each_value = /*data*/ ctx[0];
				validate_each_argument(each_value);
				let i;

				for (i = 0; i < each_value.length; i += 1) {
					const child_ctx = get_each_context$2(ctx, each_value, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
						transition_in(each_blocks[i], 1);
					} else {
						each_blocks[i] = create_each_block$2(child_ctx);
						each_blocks[i].c();
						transition_in(each_blocks[i], 1);
						each_blocks[i].m(div, null);
					}
				}

				group_outros();

				for (i = each_value.length; i < each_blocks.length; i += 1) {
					out(i);
				}

				check_outros();
			}
		},
		i: function intro(local) {
			if (current) return;

			for (let i = 0; i < each_value.length; i += 1) {
				transition_in(each_blocks[i]);
			}

			current = true;
		},
		o: function outro(local) {
			each_blocks = each_blocks.filter(Boolean);

			for (let i = 0; i < each_blocks.length; i += 1) {
				transition_out(each_blocks[i]);
			}

			current = false;
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div);
			destroy_each(each_blocks, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$8.name,
		type: "if",
		source: "(79:4) {#if data.length > 0}",
		ctx
	});

	return block;
}

// (81:8) {#each data as item}
function create_each_block$2(ctx) {
	let updating_item;
	let switch_instance_anchor;
	let current;

	function switch_instance_item_binding(value) {
		/*switch_instance_item_binding*/ ctx[11].call(null, value, /*item*/ ctx[12], /*each_value*/ ctx[13], /*item_index*/ ctx[14]);
	}

	var switch_value = /*renderItem*/ ctx[5];

	function switch_props(ctx) {
		let switch_instance_props = {};

		if (/*item*/ ctx[12] !== void 0) {
			switch_instance_props.item = /*item*/ ctx[12];
		}

		return {
			props: switch_instance_props,
			$$inline: true
		};
	}

	if (switch_value) {
		var switch_instance = new switch_value(switch_props(ctx));
		binding_callbacks.push(() => bind(switch_instance, "item", switch_instance_item_binding));
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
		p: function update(new_ctx, dirty) {
			ctx = new_ctx;
			const switch_instance_changes = {};

			if (!updating_item && dirty & /*data*/ 1) {
				updating_item = true;
				switch_instance_changes.item = /*item*/ ctx[12];
				add_flush_callback(() => updating_item = false);
			}

			if (switch_value !== (switch_value = /*renderItem*/ ctx[5])) {
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
					binding_callbacks.push(() => bind(switch_instance, "item", switch_instance_item_binding));
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
		id: create_each_block$2.name,
		type: "each",
		source: "(81:8) {#each data as item}",
		ctx
	});

	return block;
}

function create_fragment$k(ctx) {
	let div3;
	let div2;
	let div0;
	let button0;
	let t1;
	let button1;
	let t2;
	let t3_value = /*aList*/ ctx[1].info.shared_with + "";
	let t3;
	let t4;
	let t5;
	let div1;
	let h1;
	let t7;
	let p;
	let t8;
	let t9;
	let t10;
	let current;
	let dispose;
	let if_block0 = /*labels*/ ctx[4].length > 0 && create_if_block_1$4(ctx);
	let if_block1 = /*data*/ ctx[0].length > 0 && create_if_block$8(ctx);

	const block = {
		c: function create() {
			div3 = element("div");
			div2 = element("div");
			div0 = element("div");
			button0 = element("button");
			button0.textContent = "Edit list";
			t1 = space();
			button1 = element("button");
			t2 = text("View (");
			t3 = text(t3_value);
			t4 = text(")");
			t5 = space();
			div1 = element("div");
			h1 = element("h1");
			h1.textContent = `${/*title*/ ctx[3]}`;
			t7 = space();
			p = element("p");
			t8 = text(/*uuid*/ ctx[2]);
			t9 = space();
			if (if_block0) if_block0.c();
			t10 = space();
			if (if_block1) if_block1.c();
			add_location(button0, file$j, 50, 6, 1182);
			add_location(button1, file$j, 51, 6, 1233);
			add_location(div0, file$j, 49, 4, 1170);
			add_location(h1, file$j, 55, 6, 1328);
			add_location(p, file$j, 56, 6, 1351);
			add_location(div1, file$j, 54, 4, 1316);
			attr_dev(div2, "class", "pl0 measure center");
			add_location(div2, file$j, 47, 2, 1132);
			attr_dev(div3, "class", "pa3 pa5-ns");
			add_location(div3, file$j, 46, 0, 1105);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div3, anchor);
			append_dev(div3, div2);
			append_dev(div2, div0);
			append_dev(div0, button0);
			append_dev(div0, t1);
			append_dev(div0, button1);
			append_dev(button1, t2);
			append_dev(button1, t3);
			append_dev(button1, t4);
			append_dev(div2, t5);
			append_dev(div2, div1);
			append_dev(div1, h1);
			append_dev(div1, t7);
			append_dev(div1, p);
			append_dev(p, t8);
			append_dev(div2, t9);
			if (if_block0) if_block0.m(div2, null);
			append_dev(div2, t10);
			if (if_block1) if_block1.m(div2, null);
			current = true;
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(button0, "click", /*edit*/ ctx[6], false, false, false),
				listen_dev(button1, "click", /*view*/ ctx[7], false, false, false)
			];
		},
		p: function update(ctx, [dirty]) {
			if ((!current || dirty & /*aList*/ 2) && t3_value !== (t3_value = /*aList*/ ctx[1].info.shared_with + "")) set_data_dev(t3, t3_value);
			if (!current || dirty & /*uuid*/ 4) set_data_dev(t8, /*uuid*/ ctx[2]);
			if (/*labels*/ ctx[4].length > 0) if_block0.p(ctx, dirty);

			if (/*data*/ ctx[0].length > 0) {
				if (if_block1) {
					if_block1.p(ctx, dirty);
					transition_in(if_block1, 1);
				} else {
					if_block1 = create_if_block$8(ctx);
					if_block1.c();
					transition_in(if_block1, 1);
					if_block1.m(div2, null);
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
			transition_in(if_block1);
			current = true;
		},
		o: function outro(local) {
			transition_out(if_block1);
			current = false;
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div3);
			if (if_block0) if_block0.d();
			if (if_block1) if_block1.d();
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$k.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$k($$self, $$props, $$invalidate) {
	let { aList = {} } = $$props;
	let { uuid = "" } = $$props;
	let { info = { title: "", labels: [] } } = $$props;
	let { data = [] } = $$props;
	let title = info.title;
	let labels = info.labels;
	let listType = info.type;

	let items = {
		v1: List_view_data_item_v1,
		v2: List_view_data_item_v2,
		v3: List_view_data_item_v3,
		v4: List_view_data_item_v4
	};

	let renderItem = items[listType];

	function edit() {
		// This is really useful to know.
		// I dont want to store the copy.
		// TODO consider moving this into the store!
		const edit = JSON.parse(JSON.stringify(aList));

		listsEdits.add(edit);
		paths.list.edit(uuid);
	}

	function view() {
		window.open("https://learnalist.net/alists/" + aList.uuid + ".html", "_blank");
	}

	const writable_props = ["aList", "uuid", "info", "data"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_view> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_view", $$slots, []);

	function switch_instance_item_binding(value, item, each_value, item_index) {
		each_value[item_index] = value;
		$$invalidate(0, data);
	}

	$$self.$set = $$props => {
		if ("aList" in $$props) $$invalidate(1, aList = $$props.aList);
		if ("uuid" in $$props) $$invalidate(2, uuid = $$props.uuid);
		if ("info" in $$props) $$invalidate(8, info = $$props.info);
		if ("data" in $$props) $$invalidate(0, data = $$props.data);
	};

	$$self.$capture_state = () => ({
		goto: paths,
		cache,
		listsEdits,
		ItemV1: List_view_data_item_v1,
		ItemV2: List_view_data_item_v2,
		ItemV3: List_view_data_item_v3,
		ItemV4: List_view_data_item_v4,
		aList,
		uuid,
		info,
		data,
		title,
		labels,
		listType,
		items,
		renderItem,
		edit,
		view
	});

	$$self.$inject_state = $$props => {
		if ("aList" in $$props) $$invalidate(1, aList = $$props.aList);
		if ("uuid" in $$props) $$invalidate(2, uuid = $$props.uuid);
		if ("info" in $$props) $$invalidate(8, info = $$props.info);
		if ("data" in $$props) $$invalidate(0, data = $$props.data);
		if ("title" in $$props) $$invalidate(3, title = $$props.title);
		if ("labels" in $$props) $$invalidate(4, labels = $$props.labels);
		if ("listType" in $$props) listType = $$props.listType;
		if ("items" in $$props) items = $$props.items;
		if ("renderItem" in $$props) $$invalidate(5, renderItem = $$props.renderItem);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [
		data,
		aList,
		uuid,
		title,
		labels,
		renderItem,
		edit,
		view,
		info,
		listType,
		items,
		switch_instance_item_binding
	];
}

class List_view extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$k, create_fragment$k, safe_not_equal, { aList: 1, uuid: 2, info: 8, data: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_view",
			options,
			id: create_fragment$k.name
		});
	}

	get aList() {
		throw new Error("<List_view>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set aList(value) {
		throw new Error("<List_view>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	get uuid() {
		throw new Error("<List_view>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set uuid(value) {
		throw new Error("<List_view>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	get info() {
		throw new Error("<List_view>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set info(value) {
		throw new Error("<List_view>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	get data() {
		throw new Error("<List_view>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set data(value) {
		throw new Error("<List_view>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/routes/list_view.svelte generated by Svelte v3.20.1 */

const { console: console_1$2 } = globals;
const file$k = "src/editor/routes/list_view.svelte";

// (14:0) {:else}
function create_else_block$3(ctx) {
	let p;

	const block = {
		c: function create() {
			p = element("p");
			p.textContent = "Not found";
			add_location(p, file$k, 14, 2, 356);
		},
		m: function mount(target, anchor) {
			insert_dev(target, p, anchor);
		},
		p: noop,
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(p);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_else_block$3.name,
		type: "else",
		source: "(14:0) {:else}",
		ctx
	});

	return block;
}

// (12:0) {#if show}
function create_if_block$9(ctx) {
	let current;
	const listview_spread_levels = [{ aList: /*aList*/ ctx[1] }, /*aList*/ ctx[1]];
	let listview_props = {};

	for (let i = 0; i < listview_spread_levels.length; i += 1) {
		listview_props = assign(listview_props, listview_spread_levels[i]);
	}

	const listview = new List_view({ props: listview_props, $$inline: true });

	const block = {
		c: function create() {
			create_component(listview.$$.fragment);
		},
		m: function mount(target, anchor) {
			mount_component(listview, target, anchor);
			current = true;
		},
		p: function update(ctx, dirty) {
			const listview_changes = (dirty & /*aList*/ 2)
			? get_spread_update(listview_spread_levels, [{ aList: /*aList*/ ctx[1] }, get_spread_object(/*aList*/ ctx[1])])
			: {};

			listview.$set(listview_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(listview.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(listview.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(listview, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$9.name,
		type: "if",
		source: "(12:0) {#if show}",
		ctx
	});

	return block;
}

function create_fragment$l(ctx) {
	let current_block_type_index;
	let if_block;
	let if_block_anchor;
	let current;
	const if_block_creators = [create_if_block$9, create_else_block$3];
	const if_blocks = [];

	function select_block_type(ctx, dirty) {
		if (/*show*/ ctx[0]) return 0;
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
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
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
		id: create_fragment$l.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$l($$self, $$props, $$invalidate) {
	let { params = {} } = $$props;
	const aList = myLists.find(params.uuid);
	console.log(aList);
	const writable_props = ["params"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$2.warn(`<List_view> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_view", $$slots, []);

	$$self.$set = $$props => {
		if ("params" in $$props) $$invalidate(2, params = $$props.params);
	};

	$$self.$capture_state = () => ({
		cache,
		myLists,
		ListView: List_view,
		params,
		aList,
		show
	});

	$$self.$inject_state = $$props => {
		if ("params" in $$props) $$invalidate(2, params = $$props.params);
		if ("show" in $$props) $$invalidate(0, show = $$props.show);
	};

	let show;

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	 $$invalidate(0, show = aList && aList.info && aList.data);
	return [show, aList, params];
}

class List_view$1 extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$l, create_fragment$l, safe_not_equal, { params: 2 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_view",
			options,
			id: create_fragment$l.name
		});
	}

	get params() {
		throw new Error("<List_view>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set params(value) {
		throw new Error("<List_view>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list_edit_title.svelte generated by Svelte v3.20.1 */

const file$l = "src/editor/components/list_edit_title.svelte";

function add_css$1() {
	var style = element("style");
	style.id = "svelte-i2pbaw-style";
	style.textContent = ".container.svelte-i2pbaw{display:flex}input.svelte-i2pbaw{display:flex;flex-grow:1}\n/*# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibGlzdF9lZGl0X3RpdGxlLnN2ZWx0ZSIsInNvdXJjZXMiOlsibGlzdF9lZGl0X3RpdGxlLnN2ZWx0ZSJdLCJzb3VyY2VzQ29udGVudCI6WyI8c3R5bGU+XG4gIC5jb250YWluZXIge1xuICAgIGRpc3BsYXk6IGZsZXg7XG4gIH1cblxuICBpbnB1dCB7XG4gICAgZGlzcGxheTogZmxleDtcbiAgICBmbGV4LWdyb3c6IDE7IC8qIFNldCB0aGUgbWlkZGxlIGVsZW1lbnQgdG8gZ3JvdyBhbmQgc3RyZXRjaCAqL1xuICB9XG5cbi8qIyBzb3VyY2VNYXBwaW5nVVJMPWRhdGE6YXBwbGljYXRpb24vanNvbjtiYXNlNjQsZXlKMlpYSnphVzl1SWpvekxDSnpiM1Z5WTJWeklqcGJJbk55WXk5bFpHbDBiM0l2WTI5dGNHOXVaVzUwY3k5c2FYTjBYMlZrYVhSZmRHbDBiR1V1YzNabGJIUmxJbDBzSW01aGJXVnpJanBiWFN3aWJXRndjR2x1WjNNaU9pSTdSVUZEUlR0SlFVTkZMR0ZCUVdFN1JVRkRaanM3UlVGRlFUdEpRVU5GTEdGQlFXRTdTVUZEWWl4WlFVRlpMRVZCUVVVc0swTkJRU3RETzBWQlF5OUVJaXdpWm1sc1pTSTZJbk55WXk5bFpHbDBiM0l2WTI5dGNHOXVaVzUwY3k5c2FYTjBYMlZrYVhSZmRHbDBiR1V1YzNabGJIUmxJaXdpYzI5MWNtTmxjME52Ym5SbGJuUWlPbHNpWEc0Z0lDNWpiMjUwWVdsdVpYSWdlMXh1SUNBZ0lHUnBjM0JzWVhrNklHWnNaWGc3WEc0Z0lIMWNibHh1SUNCcGJuQjFkQ0I3WEc0Z0lDQWdaR2x6Y0d4aGVUb2dabXhsZUR0Y2JpQWdJQ0JtYkdWNExXZHliM2M2SURFN0lDOHFJRk5sZENCMGFHVWdiV2xrWkd4bElHVnNaVzFsYm5RZ2RHOGdaM0p2ZHlCaGJtUWdjM1J5WlhSamFDQXFMMXh1SUNCOVhHNGlYWDA9ICovPC9zdHlsZT5cblxuPHNjcmlwdD5cbiAgZXhwb3J0IGxldCB0aXRsZTtcbjwvc2NyaXB0PlxuXG48ZGl2IGNsYXNzPVwiY29udGFpbmVyXCI+XG4gIDxpbnB1dCBwbGFjZWhvbGRlcj1cIlRpdGxlXCIgYmluZDp2YWx1ZT1cInt0aXRsZX1cIiAvPlxuPC9kaXY+XG4iXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQ0UsVUFBVSxjQUFDLENBQUMsQUFDVixPQUFPLENBQUUsSUFBSSxBQUNmLENBQUMsQUFFRCxLQUFLLGNBQUMsQ0FBQyxBQUNMLE9BQU8sQ0FBRSxJQUFJLENBQ2IsU0FBUyxDQUFFLENBQUMsQUFDZCxDQUFDIn0= */";
	append_dev(document.head, style);
}

function create_fragment$m(ctx) {
	let div;
	let input;
	let dispose;

	const block = {
		c: function create() {
			div = element("div");
			input = element("input");
			attr_dev(input, "placeholder", "Title");
			attr_dev(input, "class", "svelte-i2pbaw");
			add_location(input, file$l, 17, 2, 793);
			attr_dev(div, "class", "container svelte-i2pbaw");
			add_location(div, file$l, 16, 0, 767);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div, anchor);
			append_dev(div, input);
			set_input_value(input, /*title*/ ctx[0]);
			if (remount) dispose();
			dispose = listen_dev(input, "input", /*input_input_handler*/ ctx[1]);
		},
		p: function update(ctx, [dirty]) {
			if (dirty & /*title*/ 1 && input.value !== /*title*/ ctx[0]) {
				set_input_value(input, /*title*/ ctx[0]);
			}
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(div);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$m.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$m($$self, $$props, $$invalidate) {
	let { title } = $$props;
	const writable_props = ["title"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_edit_title> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_edit_title", $$slots, []);

	function input_input_handler() {
		title = this.value;
		$$invalidate(0, title);
	}

	$$self.$set = $$props => {
		if ("title" in $$props) $$invalidate(0, title = $$props.title);
	};

	$$self.$capture_state = () => ({ title });

	$$self.$inject_state = $$props => {
		if ("title" in $$props) $$invalidate(0, title = $$props.title);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [title, input_input_handler];
}

class List_edit_title extends SvelteComponentDev {
	constructor(options) {
		super(options);
		if (!document.getElementById("svelte-i2pbaw-style")) add_css$1();
		init(this, options, instance$m, create_fragment$m, safe_not_equal, { title: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_edit_title",
			options,
			id: create_fragment$m.name
		});

		const { ctx } = this.$$;
		const props = options.props || {};

		if (/*title*/ ctx[0] === undefined && !("title" in props)) {
			console.warn("<List_edit_title> was created without expected prop 'title'");
		}
	}

	get title() {
		throw new Error("<List_edit_title>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set title(value) {
		throw new Error("<List_edit_title>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
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

/* src/editor/components/list_edit_data_v1.svelte generated by Svelte v3.20.1 */
const file$m = "src/editor/components/list_edit_data_v1.svelte";

function add_css$2() {
	var style = element("style");
	style.id = "svelte-1havqk5-style";
	style.textContent = "input.svelte-1havqk5.svelte-1havqk5:disabled{background:#ffcccc;color:#333}.item-container.svelte-1havqk5.svelte-1havqk5{display:flex}.item-container.svelte-1havqk5 .item.svelte-1havqk5{}.item-container.svelte-1havqk5 .item-left.svelte-1havqk5{flex-grow:1;margin-right:0.5em}\n/*# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibGlzdF9lZGl0X2RhdGFfdjEuc3ZlbHRlIiwic291cmNlcyI6WyJsaXN0X2VkaXRfZGF0YV92MS5zdmVsdGUiXSwic291cmNlc0NvbnRlbnQiOlsiPHN0eWxlPlxuICBpbnB1dDpkaXNhYmxlZCB7XG4gICAgYmFja2dyb3VuZDogI2ZmY2NjYztcbiAgICBjb2xvcjogIzMzMztcbiAgfVxuXG4gIC5jb250YWluZXIge1xuICAgIGRpc3BsYXk6IGZsZXg7XG4gICAganVzdGlmeS1jb250ZW50OiBzcGFjZS1iZXR3ZWVuO1xuICAgIGZsZXgtZGlyZWN0aW9uOiBjb2x1bW47XG4gIH1cblxuICAuaXRlbS1jb250YWluZXIge1xuICAgIGRpc3BsYXk6IGZsZXg7XG4gIH1cblxuICAuaXRlbS1jb250YWluZXIgLml0ZW0ge1xuICB9XG5cbiAgLml0ZW0tY29udGFpbmVyIC5pdGVtLWxlZnQge1xuICAgIGZsZXgtZ3JvdzogMTsgLyogU2V0IHRoZSBtaWRkbGUgZWxlbWVudCB0byBncm93IGFuZCBzdHJldGNoICovXG4gICAgbWFyZ2luLXJpZ2h0OiAwLjVlbTtcbiAgfVxuXG4vKiMgc291cmNlTWFwcGluZ1VSTD1kYXRhOmFwcGxpY2F0aW9uL2pzb247YmFzZTY0LGV5SjJaWEp6YVc5dUlqb3pMQ0p6YjNWeVkyVnpJanBiSW5OeVl5OWxaR2wwYjNJdlkyOXRjRzl1Wlc1MGN5OXNhWE4wWDJWa2FYUmZaR0YwWVY5Mk1TNXpkbVZzZEdVaVhTd2libUZ0WlhNaU9sdGRMQ0p0WVhCd2FXNW5jeUk2SWp0RlFVTkZPMGxCUTBVc2JVSkJRVzFDTzBsQlEyNUNMRmRCUVZjN1JVRkRZanM3UlVGRlFUdEpRVU5GTEdGQlFXRTdTVUZEWWl3NFFrRkJPRUk3U1VGRE9VSXNjMEpCUVhOQ08wVkJRM2hDT3p0RlFVVkJPMGxCUTBVc1lVRkJZVHRGUVVObU96dEZRVVZCTzBWQlEwRTdPMFZCUlVFN1NVRkRSU3haUVVGWkxFVkJRVVVzSzBOQlFTdERPMGxCUXpkRUxHMUNRVUZ0UWp0RlFVTnlRaUlzSW1acGJHVWlPaUp6Y21NdlpXUnBkRzl5TDJOdmJYQnZibVZ1ZEhNdmJHbHpkRjlsWkdsMFgyUmhkR0ZmZGpFdWMzWmxiSFJsSWl3aWMyOTFjbU5sYzBOdmJuUmxiblFpT2xzaVhHNGdJR2x1Y0hWME9tUnBjMkZpYkdWa0lIdGNiaUFnSUNCaVlXTnJaM0p2ZFc1a09pQWpabVpqWTJOak8xeHVJQ0FnSUdOdmJHOXlPaUFqTXpNek8xeHVJQ0I5WEc1Y2JpQWdMbU52Ym5SaGFXNWxjaUI3WEc0Z0lDQWdaR2x6Y0d4aGVUb2dabXhsZUR0Y2JpQWdJQ0JxZFhOMGFXWjVMV052Ym5SbGJuUTZJSE53WVdObExXSmxkSGRsWlc0N1hHNGdJQ0FnWm14bGVDMWthWEpsWTNScGIyNDZJR052YkhWdGJqdGNiaUFnZlZ4dVhHNGdJQzVwZEdWdExXTnZiblJoYVc1bGNpQjdYRzRnSUNBZ1pHbHpjR3hoZVRvZ1pteGxlRHRjYmlBZ2ZWeHVYRzRnSUM1cGRHVnRMV052Ym5SaGFXNWxjaUF1YVhSbGJTQjdYRzRnSUgxY2JseHVJQ0F1YVhSbGJTMWpiMjUwWVdsdVpYSWdMbWwwWlcwdGJHVm1kQ0I3WEc0Z0lDQWdabXhsZUMxbmNtOTNPaUF4T3lBdktpQlRaWFFnZEdobElHMXBaR1JzWlNCbGJHVnRaVzUwSUhSdklHZHliM2NnWVc1a0lITjBjbVYwWTJnZ0tpOWNiaUFnSUNCdFlYSm5hVzR0Y21sbmFIUTZJREF1TldWdE8xeHVJQ0I5WEc0aVhYMD0gKi88L3N0eWxlPlxuXG48c2NyaXB0PlxuICBpbXBvcnQgeyBjb3B5T2JqZWN0LCBpc0RldmljZU1vYmlsZSB9IGZyb20gXCIuLi9saWIvaGVscGVyLmpzXCI7XG4gIGltcG9ydCB7IHRhcCB9IGZyb20gXCJAc3ZlbHRlanMvZ2VzdHVyZXNcIjtcbiAgaW1wb3J0IHsgYWZ0ZXJVcGRhdGUgfSBmcm9tIFwic3ZlbHRlXCI7XG5cbiAgY29uc3QgcG9zc2libGVDb21tYW5kcyA9IHtcbiAgICBub3RoaW5nOiBcIlwiLFxuICAgIG5ld0l0ZW06IFwiV2hlbiBhbiBpdGVtIGlzIGFkZGVkXCJcbiAgfTtcblxuICBjb25zdCBpc01vYmlsZSA9IGlzRGV2aWNlTW9iaWxlKCk7XG4gIGNvbnN0IG9yZGVySGVscGVyVGV4dCA9ICFpc01vYmlsZSA/IFwiZHJhZyBhbmQgZHJvcCB0byBzd2FwXCIgOiBcInRhcCB0byBzd2FwXCI7XG5cbiAgY29uc3QgbmV3Um93ID0gXCJcIjtcbiAgY29uc3QgX3N3YXBJdGVtcyA9IHtcbiAgICBmcm9tOiAtMSxcbiAgICBmcm9tRWxlbWVudDogbnVsbCxcbiAgICB0bzogLTEsXG4gICAgdG9FbGVtZW50OiBudWxsXG4gIH07XG5cbiAgbGV0IGl0ZW1zQ29udGFpbmVyO1xuICBsZXQgbGFzdENtZCA9IHBvc3NpYmxlQ29tbWFuZHMubm90aGluZztcblxuICBsZXQgZW5hYmxlU29ydGFibGUgPSBmYWxzZTtcbiAgZXhwb3J0IGxldCBsaXN0RGF0YTtcbiAgbGV0IHN3YXBJdGVtcyA9IGNvcHlPYmplY3QoX3N3YXBJdGVtcyk7XG5cbiAgYWZ0ZXJVcGRhdGUoKCkgPT4ge1xuICAgIGlmIChsYXN0Q21kID09PSBwb3NzaWJsZUNvbW1hbmRzLm5ld0l0ZW0pIHtcbiAgICAgIC8vIFRoaXMgb25seSB3b3JrcyBmb3IgVjEgZWxlbWVudHNcbiAgICAgIGxldCBub2RlcyA9IGl0ZW1zQ29udGFpbmVyLnF1ZXJ5U2VsZWN0b3JBbGwoXCIuaXRlbS1jb250YWluZXJcIik7XG4gICAgICBub2Rlc1tub2Rlcy5sZW5ndGggLSAxXS5xdWVyeVNlbGVjdG9yKFwiaW5wdXQ6Zmlyc3QtY2hpbGRcIikuZm9jdXMoKTtcbiAgICAgIGxhc3RDbWQgPSBwb3NzaWJsZUNvbW1hbmRzLm5vdGhpbmc7XG4gICAgfVxuICB9KTtcblxuICBmdW5jdGlvbiBhZGQoKSB7XG4gICAgbGlzdERhdGEgPSBsaXN0RGF0YS5jb25jYXQoY29weU9iamVjdChuZXdSb3cpKTtcbiAgICBsYXN0Q21kID0gcG9zc2libGVDb21tYW5kcy5uZXdJdGVtO1xuICB9XG5cbiAgZnVuY3Rpb24gcmVtb3ZlKGxpc3RJdGVtKSB7XG4gICAgbGlzdERhdGEgPSBsaXN0RGF0YS5maWx0ZXIodCA9PiB0ICE9PSBsaXN0SXRlbSk7XG4gICAgaWYgKCFsaXN0RGF0YS5sZW5ndGgpIHtcbiAgICAgIGxpc3REYXRhID0gW2NvcHlPYmplY3QobmV3Um93KV07XG4gICAgfVxuICB9XG5cbiAgZnVuY3Rpb24gcmVtb3ZlQWxsKCkge1xuICAgIGxpc3REYXRhID0gW2NvcHlPYmplY3QobmV3Um93KV07XG4gIH1cblxuICBmdW5jdGlvbiB0b2dnbGVTb3J0YWJsZShldikge1xuICAgIGlmIChsaXN0RGF0YS5sZW5ndGggPD0gMSkge1xuICAgICAgYWxlcnQoXCJub3RoaW5nIHRvIHN3YXBcIik7XG4gICAgICByZXR1cm47XG4gICAgfVxuXG4gICAgZW5hYmxlU29ydGFibGUgPSBlbmFibGVTb3J0YWJsZSA/IGZhbHNlIDogdHJ1ZTtcbiAgICBpZiAoZW5hYmxlU29ydGFibGUpIHtcbiAgICAgIC8vIFJlc2V0IHN3YXBJdGVtc1xuICAgICAgc3dhcEl0ZW1zID0gY29weU9iamVjdChfc3dhcEl0ZW1zKTtcbiAgICB9XG4gIH1cblxuICBmdW5jdGlvbiBkcmFnc3RhcnQoZXYpIHtcbiAgICBzd2FwSXRlbXMgPSBjb3B5T2JqZWN0KF9zd2FwSXRlbXMpO1xuICAgIHN3YXBJdGVtcy5mcm9tID0gZXYudGFyZ2V0LmdldEF0dHJpYnV0ZShcImRhdGEtaW5kZXhcIik7XG4gIH1cblxuICBmdW5jdGlvbiBkcmFnb3Zlcihldikge1xuICAgIGV2LnByZXZlbnREZWZhdWx0KCk7XG4gIH1cblxuICBmdW5jdGlvbiBkcm9wKGV2KSB7XG4gICAgZXYucHJldmVudERlZmF1bHQoKTtcbiAgICBzd2FwSXRlbXMudG8gPSBldi50YXJnZXQuZ2V0QXR0cmlidXRlKFwiZGF0YS1pbmRleFwiKTtcblxuICAgIC8vIFdlIG1pZ2h0IGxhbmQgb24gdGhlIGNoaWxkcmVuLCBsb29rIHVwIGZvciB0aGUgZHJhZ2dhYmxlIGF0dHJpYnV0ZVxuICAgIGlmIChzd2FwSXRlbXMudG8gPT0gbnVsbCkge1xuICAgICAgc3dhcEl0ZW1zLnRvID0gZXYudGFyZ2V0XG4gICAgICAgIC5jbG9zZXN0KFwiW2RyYWdnYWJsZV1cIilcbiAgICAgICAgLmdldEF0dHJpYnV0ZShcImRhdGEtaW5kZXhcIik7XG4gICAgfVxuXG4gICAgbGV0IGEgPSBsaXN0RGF0YVtzd2FwSXRlbXMuZnJvbV07XG4gICAgbGV0IGIgPSBsaXN0RGF0YVtzd2FwSXRlbXMudG9dO1xuICAgIGxpc3REYXRhW3N3YXBJdGVtcy5mcm9tXSA9IGI7XG4gICAgbGlzdERhdGFbc3dhcEl0ZW1zLnRvXSA9IGE7XG4gIH1cblxuICBmdW5jdGlvbiB0YXBIYW5kbGVyKGV2KSB7XG4gICAgZXYucHJldmVudERlZmF1bHQoKTtcblxuICAgIGxldCBpbmRleCA9IGV2LnRhcmdldC5nZXRBdHRyaWJ1dGUoXCJkYXRhLWluZGV4XCIpO1xuXG4gICAgaWYgKGluZGV4ID09PSBudWxsKSB7XG4gICAgICBzd2FwSXRlbXMgPSBjb3B5T2JqZWN0KF9zd2FwSXRlbXMpO1xuICAgICAgcmV0dXJuO1xuICAgIH1cblxuICAgIGlmIChzd2FwSXRlbXMuZnJvbSA9PT0gLTEpIHtcbiAgICAgIHN3YXBJdGVtcy5mcm9tRWxlbWVudCA9IGV2LnRhcmdldDtcbiAgICAgIHN3YXBJdGVtcy5mcm9tRWxlbWVudC5zdHlsZVtcImJvcmRlci1sZWZ0XCJdID0gXCJzb2xpZCBncmVlblwiO1xuICAgICAgc3dhcEl0ZW1zLmZyb20gPSBpbmRleDtcbiAgICAgIHJldHVybjtcbiAgICB9XG5cbiAgICBpZiAoc3dhcEl0ZW1zLmZyb20gPT09IGluZGV4KSB7XG4gICAgICBzd2FwSXRlbXMuZnJvbUVsZW1lbnQuc3R5bGUuYm9yZGVyID0gXCJcIjtcbiAgICAgIHN3YXBJdGVtcyA9IGNvcHlPYmplY3QoX3N3YXBJdGVtcyk7XG4gICAgICByZXR1cm47XG4gICAgfVxuXG4gICAgc3dhcEl0ZW1zLnRvID0gaW5kZXg7XG4gICAgbGV0IGEgPSBsaXN0RGF0YVtzd2FwSXRlbXMuZnJvbV07XG4gICAgbGV0IGIgPSBsaXN0RGF0YVtzd2FwSXRlbXMudG9dO1xuICAgIGxpc3REYXRhW3N3YXBJdGVtcy5mcm9tXSA9IGI7XG4gICAgbGlzdERhdGFbc3dhcEl0ZW1zLnRvXSA9IGE7XG5cbiAgICBzd2FwSXRlbXMuZnJvbUVsZW1lbnQuc3R5bGUuYm9yZGVyID0gXCJcIjtcbiAgICBzd2FwSXRlbXMuZnJvbUVsZW1lbnQuc3R5bGVbXCJib3JkZXItcmFkaXVzXCJdID0gXCIwcHhcIjtcbiAgICBzd2FwSXRlbXMgPSBjb3B5T2JqZWN0KF9zd2FwSXRlbXMpO1xuICB9XG48L3NjcmlwdD5cblxuPGgxPkl0ZW1zPC9oMT5cblxuPGRpdiBiaW5kOnRoaXM9XCJ7aXRlbXNDb250YWluZXJ9XCI+XG4gIHsjaWYgIWVuYWJsZVNvcnRhYmxlfVxuICAgIHsjZWFjaCBsaXN0RGF0YSBhcyBsaXN0SXRlbX1cbiAgICAgIDxkaXYgY2xhc3M9XCJpdGVtLWNvbnRhaW5lclwiPlxuICAgICAgICA8aW5wdXQgYmluZDp2YWx1ZT1cIntsaXN0SXRlbX1cIiBjbGFzcz1cIml0ZW0gaXRlbS1sZWZ0XCIgLz5cbiAgICAgICAgPGJ1dHRvbiBvbjpjbGljaz1cInsoKSA9PiByZW1vdmUobGlzdEl0ZW0pfVwiIGNsYXNzPVwiaXRlbVwiPng8L2J1dHRvbj5cbiAgICAgIDwvZGl2PlxuICAgIHsvZWFjaH1cblxuICAgIDxidXR0b24gb246Y2xpY2s9XCJ7YWRkfVwiPk5ldzwvYnV0dG9uPlxuXG4gICAgPGJ1dHRvbiBvbjpjbGljaz1cIntyZW1vdmVBbGx9XCI+UmVtb3ZlIGFsbDwvYnV0dG9uPlxuXG4gICAgPGJ1dHRvbiBvbjpjbGljaz1cInt0b2dnbGVTb3J0YWJsZX1cIj5DaGFuZ2UgT3JkZXI8L2J1dHRvbj5cbiAgey9pZn1cblxuICB7I2lmIGVuYWJsZVNvcnRhYmxlfVxuICAgIHsjZWFjaCBsaXN0RGF0YSBhcyBsaXN0SXRlbSwgcG9zfVxuICAgICAgPGRpdlxuICAgICAgICBkcmFnZ2FibGU9XCJ0cnVlXCJcbiAgICAgICAgY2xhc3M9XCJkcm9wem9uZSBpdGVtLWNvbnRhaW5lclwiXG4gICAgICAgIGRhdGEtaW5kZXg9XCJ7cG9zfVwiXG4gICAgICAgIG9uOmRyYWdzdGFydD1cIntkcmFnc3RhcnR9XCJcbiAgICAgICAgb246ZHJhZ292ZXI9XCJ7ZHJhZ292ZXJ9XCJcbiAgICAgICAgb246ZHJvcD1cIntkcm9wfVwiXG4gICAgICAgIHVzZTp0YXBcbiAgICAgICAgb246dGFwPVwie3RhcEhhbmRsZXJ9XCJcbiAgICAgID5cbiAgICAgICAgPGlucHV0IGNsYXNzPVwiaXRlbSBpdGVtLWxlZnRcIiB2YWx1ZT1cIntsaXN0SXRlbX1cIiBkaXNhYmxlZCAvPlxuICAgICAgPC9kaXY+XG4gICAgey9lYWNofVxuXG4gICAgPGJ1dHRvbiBvbjpjbGljaz1cInt0b2dnbGVTb3J0YWJsZX1cIj5cbiAgICAgIEZpbmlzaGVkIG9yZGVyaW5nPyAoe29yZGVySGVscGVyVGV4dH0pXG4gICAgPC9idXR0b24+XG4gIHsvaWZ9XG48L2Rpdj5cbiJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFDRSxtQ0FBSyxTQUFTLEFBQUMsQ0FBQyxBQUNkLFVBQVUsQ0FBRSxPQUFPLENBQ25CLEtBQUssQ0FBRSxJQUFJLEFBQ2IsQ0FBQyxBQVFELGVBQWUsOEJBQUMsQ0FBQyxBQUNmLE9BQU8sQ0FBRSxJQUFJLEFBQ2YsQ0FBQyxBQUVELDhCQUFlLENBQUMsS0FBSyxlQUFDLENBQUMsQUFDdkIsQ0FBQyxBQUVELDhCQUFlLENBQUMsVUFBVSxlQUFDLENBQUMsQUFDMUIsU0FBUyxDQUFFLENBQUMsQ0FDWixZQUFZLENBQUUsS0FBSyxBQUNyQixDQUFDIn0= */";
	append_dev(document.head, style);
}

function get_each_context$3(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[19] = list[i];
	child_ctx[21] = i;
	return child_ctx;
}

function get_each_context_1$2(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[19] = list[i];
	child_ctx[22] = list;
	child_ctx[23] = i;
	return child_ctx;
}

// (157:2) {#if !enableSortable}
function create_if_block_1$5(ctx) {
	let t0;
	let button0;
	let t2;
	let button1;
	let t4;
	let button2;
	let dispose;
	let each_value_1 = /*listData*/ ctx[0];
	validate_each_argument(each_value_1);
	let each_blocks = [];

	for (let i = 0; i < each_value_1.length; i += 1) {
		each_blocks[i] = create_each_block_1$2(get_each_context_1$2(ctx, each_value_1, i));
	}

	const block = {
		c: function create() {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			t0 = space();
			button0 = element("button");
			button0.textContent = "New";
			t2 = space();
			button1 = element("button");
			button1.textContent = "Remove all";
			t4 = space();
			button2 = element("button");
			button2.textContent = "Change Order";
			add_location(button0, file$m, 164, 4, 4815);
			add_location(button1, file$m, 166, 4, 4858);
			add_location(button2, file$m, 168, 4, 4914);
		},
		m: function mount(target, anchor, remount) {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(target, anchor);
			}

			insert_dev(target, t0, anchor);
			insert_dev(target, button0, anchor);
			insert_dev(target, t2, anchor);
			insert_dev(target, button1, anchor);
			insert_dev(target, t4, anchor);
			insert_dev(target, button2, anchor);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(button0, "click", /*add*/ ctx[4], false, false, false),
				listen_dev(button1, "click", /*removeAll*/ ctx[6], false, false, false),
				listen_dev(button2, "click", /*toggleSortable*/ ctx[7], false, false, false)
			];
		},
		p: function update(ctx, dirty) {
			if (dirty & /*remove, listData*/ 33) {
				each_value_1 = /*listData*/ ctx[0];
				validate_each_argument(each_value_1);
				let i;

				for (i = 0; i < each_value_1.length; i += 1) {
					const child_ctx = get_each_context_1$2(ctx, each_value_1, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
					} else {
						each_blocks[i] = create_each_block_1$2(child_ctx);
						each_blocks[i].c();
						each_blocks[i].m(t0.parentNode, t0);
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
			if (detaching) detach_dev(t0);
			if (detaching) detach_dev(button0);
			if (detaching) detach_dev(t2);
			if (detaching) detach_dev(button1);
			if (detaching) detach_dev(t4);
			if (detaching) detach_dev(button2);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_1$5.name,
		type: "if",
		source: "(157:2) {#if !enableSortable}",
		ctx
	});

	return block;
}

// (158:4) {#each listData as listItem}
function create_each_block_1$2(ctx) {
	let div;
	let input;
	let t0;
	let button;
	let dispose;

	function input_input_handler() {
		/*input_input_handler*/ ctx[16].call(input, /*listItem*/ ctx[19], /*each_value_1*/ ctx[22], /*listItem_index*/ ctx[23]);
	}

	function click_handler(...args) {
		return /*click_handler*/ ctx[17](/*listItem*/ ctx[19], ...args);
	}

	const block = {
		c: function create() {
			div = element("div");
			input = element("input");
			t0 = space();
			button = element("button");
			button.textContent = "x";
			attr_dev(input, "class", "item item-left svelte-1havqk5");
			add_location(input, file$m, 159, 8, 4652);
			attr_dev(button, "class", "item svelte-1havqk5");
			add_location(button, file$m, 160, 8, 4717);
			attr_dev(div, "class", "item-container svelte-1havqk5");
			add_location(div, file$m, 158, 6, 4615);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div, anchor);
			append_dev(div, input);
			set_input_value(input, /*listItem*/ ctx[19]);
			append_dev(div, t0);
			append_dev(div, button);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(input, "input", input_input_handler),
				listen_dev(button, "click", click_handler, false, false, false)
			];
		},
		p: function update(new_ctx, dirty) {
			ctx = new_ctx;

			if (dirty & /*listData*/ 1 && input.value !== /*listItem*/ ctx[19]) {
				set_input_value(input, /*listItem*/ ctx[19]);
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block_1$2.name,
		type: "each",
		source: "(158:4) {#each listData as listItem}",
		ctx
	});

	return block;
}

// (172:2) {#if enableSortable}
function create_if_block$a(ctx) {
	let t0;
	let button;
	let dispose;
	let each_value = /*listData*/ ctx[0];
	validate_each_argument(each_value);
	let each_blocks = [];

	for (let i = 0; i < each_value.length; i += 1) {
		each_blocks[i] = create_each_block$3(get_each_context$3(ctx, each_value, i));
	}

	const block = {
		c: function create() {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			t0 = space();
			button = element("button");

			button.textContent = `
      Finished ordering? (${/*orderHelperText*/ ctx[3]})
    `;

			add_location(button, file$m, 187, 4, 5391);
		},
		m: function mount(target, anchor, remount) {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(target, anchor);
			}

			insert_dev(target, t0, anchor);
			insert_dev(target, button, anchor);
			if (remount) dispose();
			dispose = listen_dev(button, "click", /*toggleSortable*/ ctx[7], false, false, false);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*dragstart, dragover, drop, tapHandler, listData*/ 1793) {
				each_value = /*listData*/ ctx[0];
				validate_each_argument(each_value);
				let i;

				for (i = 0; i < each_value.length; i += 1) {
					const child_ctx = get_each_context$3(ctx, each_value, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
					} else {
						each_blocks[i] = create_each_block$3(child_ctx);
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
			if (detaching) detach_dev(button);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$a.name,
		type: "if",
		source: "(172:2) {#if enableSortable}",
		ctx
	});

	return block;
}

// (173:4) {#each listData as listItem, pos}
function create_each_block$3(ctx) {
	let div;
	let input;
	let input_value_value;
	let div_data_index_value;
	let tap_action;
	let dispose;

	const block = {
		c: function create() {
			div = element("div");
			input = element("input");
			attr_dev(input, "class", "item item-left svelte-1havqk5");
			input.value = input_value_value = /*listItem*/ ctx[19];
			input.disabled = true;
			add_location(input, file$m, 183, 8, 5300);
			attr_dev(div, "draggable", "true");
			attr_dev(div, "class", "dropzone item-container svelte-1havqk5");
			attr_dev(div, "data-index", div_data_index_value = /*pos*/ ctx[21]);
			add_location(div, file$m, 173, 6, 5048);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div, anchor);
			append_dev(div, input);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(div, "dragstart", /*dragstart*/ ctx[8], false, false, false),
				listen_dev(div, "dragover", dragover, false, false, false),
				listen_dev(div, "drop", /*drop*/ ctx[9], false, false, false),
				action_destroyer(tap_action = tap.call(null, div)),
				listen_dev(div, "tap", /*tapHandler*/ ctx[10], false, false, false)
			];
		},
		p: function update(ctx, dirty) {
			if (dirty & /*listData*/ 1 && input_value_value !== (input_value_value = /*listItem*/ ctx[19]) && input.value !== input_value_value) {
				prop_dev(input, "value", input_value_value);
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block$3.name,
		type: "each",
		source: "(173:4) {#each listData as listItem, pos}",
		ctx
	});

	return block;
}

function create_fragment$n(ctx) {
	let h1;
	let t1;
	let div;
	let t2;
	let if_block0 = !/*enableSortable*/ ctx[2] && create_if_block_1$5(ctx);
	let if_block1 = /*enableSortable*/ ctx[2] && create_if_block$a(ctx);

	const block = {
		c: function create() {
			h1 = element("h1");
			h1.textContent = "Items";
			t1 = space();
			div = element("div");
			if (if_block0) if_block0.c();
			t2 = space();
			if (if_block1) if_block1.c();
			add_location(h1, file$m, 153, 0, 4501);
			add_location(div, file$m, 155, 0, 4517);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, h1, anchor);
			insert_dev(target, t1, anchor);
			insert_dev(target, div, anchor);
			if (if_block0) if_block0.m(div, null);
			append_dev(div, t2);
			if (if_block1) if_block1.m(div, null);
			/*div_binding*/ ctx[18](div);
		},
		p: function update(ctx, [dirty]) {
			if (!/*enableSortable*/ ctx[2]) {
				if (if_block0) {
					if_block0.p(ctx, dirty);
				} else {
					if_block0 = create_if_block_1$5(ctx);
					if_block0.c();
					if_block0.m(div, t2);
				}
			} else if (if_block0) {
				if_block0.d(1);
				if_block0 = null;
			}

			if (/*enableSortable*/ ctx[2]) {
				if (if_block1) {
					if_block1.p(ctx, dirty);
				} else {
					if_block1 = create_if_block$a(ctx);
					if_block1.c();
					if_block1.m(div, null);
				}
			} else if (if_block1) {
				if_block1.d(1);
				if_block1 = null;
			}
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(h1);
			if (detaching) detach_dev(t1);
			if (detaching) detach_dev(div);
			if (if_block0) if_block0.d();
			if (if_block1) if_block1.d();
			/*div_binding*/ ctx[18](null);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$n.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

const newRow = "";

function dragover(ev) {
	ev.preventDefault();
}

function instance$n($$self, $$props, $$invalidate) {
	const possibleCommands = {
		nothing: "",
		newItem: "When an item is added"
	};

	const isMobile = isDeviceMobile();
	const orderHelperText = !isMobile ? "drag and drop to swap" : "tap to swap";

	const _swapItems = {
		from: -1,
		fromElement: null,
		to: -1,
		toElement: null
	};

	let itemsContainer;
	let lastCmd = possibleCommands.nothing;
	let enableSortable = false;
	let { listData } = $$props;
	let swapItems = copyObject(_swapItems);

	afterUpdate(() => {
		if (lastCmd === possibleCommands.newItem) {
			// This only works for V1 elements
			let nodes = itemsContainer.querySelectorAll(".item-container");

			nodes[nodes.length - 1].querySelector("input:first-child").focus();
			lastCmd = possibleCommands.nothing;
		}
	});

	function add() {
		$$invalidate(0, listData = listData.concat(copyObject(newRow)));
		lastCmd = possibleCommands.newItem;
	}

	function remove(listItem) {
		$$invalidate(0, listData = listData.filter(t => t !== listItem));

		if (!listData.length) {
			$$invalidate(0, listData = [copyObject(newRow)]);
		}
	}

	function removeAll() {
		$$invalidate(0, listData = [copyObject(newRow)]);
	}

	function toggleSortable(ev) {
		if (listData.length <= 1) {
			alert("nothing to swap");
			return;
		}

		$$invalidate(2, enableSortable = enableSortable ? false : true);

		if (enableSortable) {
			// Reset swapItems
			swapItems = copyObject(_swapItems);
		}
	}

	function dragstart(ev) {
		swapItems = copyObject(_swapItems);
		swapItems.from = ev.target.getAttribute("data-index");
	}

	function drop(ev) {
		ev.preventDefault();
		swapItems.to = ev.target.getAttribute("data-index");

		// We might land on the children, look up for the draggable attribute
		if (swapItems.to == null) {
			swapItems.to = ev.target.closest("[draggable]").getAttribute("data-index");
		}

		let a = listData[swapItems.from];
		let b = listData[swapItems.to];
		$$invalidate(0, listData[swapItems.from] = b, listData);
		$$invalidate(0, listData[swapItems.to] = a, listData);
	}

	function tapHandler(ev) {
		ev.preventDefault();
		let index = ev.target.getAttribute("data-index");

		if (index === null) {
			swapItems = copyObject(_swapItems);
			return;
		}

		if (swapItems.from === -1) {
			swapItems.fromElement = ev.target;
			swapItems.fromElement.style["border-left"] = "solid green";
			swapItems.from = index;
			return;
		}

		if (swapItems.from === index) {
			swapItems.fromElement.style.border = "";
			swapItems = copyObject(_swapItems);
			return;
		}

		swapItems.to = index;
		let a = listData[swapItems.from];
		let b = listData[swapItems.to];
		$$invalidate(0, listData[swapItems.from] = b, listData);
		$$invalidate(0, listData[swapItems.to] = a, listData);
		swapItems.fromElement.style.border = "";
		swapItems.fromElement.style["border-radius"] = "0px";
		swapItems = copyObject(_swapItems);
	}

	const writable_props = ["listData"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_edit_data_v1> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_edit_data_v1", $$slots, []);

	function input_input_handler(listItem, each_value_1, listItem_index) {
		each_value_1[listItem_index] = this.value;
		$$invalidate(0, listData);
	}

	const click_handler = listItem => remove(listItem);

	function div_binding($$value) {
		binding_callbacks[$$value ? "unshift" : "push"](() => {
			$$invalidate(1, itemsContainer = $$value);
		});
	}

	$$self.$set = $$props => {
		if ("listData" in $$props) $$invalidate(0, listData = $$props.listData);
	};

	$$self.$capture_state = () => ({
		copyObject,
		isDeviceMobile,
		tap,
		afterUpdate,
		possibleCommands,
		isMobile,
		orderHelperText,
		newRow,
		_swapItems,
		itemsContainer,
		lastCmd,
		enableSortable,
		listData,
		swapItems,
		add,
		remove,
		removeAll,
		toggleSortable,
		dragstart,
		dragover,
		drop,
		tapHandler
	});

	$$self.$inject_state = $$props => {
		if ("itemsContainer" in $$props) $$invalidate(1, itemsContainer = $$props.itemsContainer);
		if ("lastCmd" in $$props) lastCmd = $$props.lastCmd;
		if ("enableSortable" in $$props) $$invalidate(2, enableSortable = $$props.enableSortable);
		if ("listData" in $$props) $$invalidate(0, listData = $$props.listData);
		if ("swapItems" in $$props) swapItems = $$props.swapItems;
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [
		listData,
		itemsContainer,
		enableSortable,
		orderHelperText,
		add,
		remove,
		removeAll,
		toggleSortable,
		dragstart,
		drop,
		tapHandler,
		lastCmd,
		swapItems,
		possibleCommands,
		isMobile,
		_swapItems,
		input_input_handler,
		click_handler,
		div_binding
	];
}

class List_edit_data_v1 extends SvelteComponentDev {
	constructor(options) {
		super(options);
		if (!document.getElementById("svelte-1havqk5-style")) add_css$2();
		init(this, options, instance$n, create_fragment$n, safe_not_equal, { listData: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_edit_data_v1",
			options,
			id: create_fragment$n.name
		});

		const { ctx } = this.$$;
		const props = options.props || {};

		if (/*listData*/ ctx[0] === undefined && !("listData" in props)) {
			console.warn("<List_edit_data_v1> was created without expected prop 'listData'");
		}
	}

	get listData() {
		throw new Error("<List_edit_data_v1>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set listData(value) {
		throw new Error("<List_edit_data_v1>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list_edit_data_v2.svelte generated by Svelte v3.20.1 */
const file$n = "src/editor/components/list_edit_data_v2.svelte";

function add_css$3() {
	var style = element("style");
	style.id = "svelte-vda4an-style";
	style.textContent = "input.svelte-vda4an.svelte-vda4an:disabled{background:#ffcccc;color:#333}.item-container.svelte-vda4an.svelte-vda4an{display:flex}.item-container.svelte-vda4an .item.svelte-vda4an{}.item-container.svelte-vda4an .item-left.svelte-vda4an{flex-grow:1;margin-right:0.5em}\n/*# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibGlzdF9lZGl0X2RhdGFfdjIuc3ZlbHRlIiwic291cmNlcyI6WyJsaXN0X2VkaXRfZGF0YV92Mi5zdmVsdGUiXSwic291cmNlc0NvbnRlbnQiOlsiPHN0eWxlPlxuICBpbnB1dDpkaXNhYmxlZCB7XG4gICAgYmFja2dyb3VuZDogI2ZmY2NjYztcbiAgICBjb2xvcjogIzMzMztcbiAgfVxuXG4gIC5jb250YWluZXIge1xuICAgIGRpc3BsYXk6IGZsZXg7XG4gICAganVzdGlmeS1jb250ZW50OiBzcGFjZS1iZXR3ZWVuO1xuICAgIGZsZXgtZGlyZWN0aW9uOiBjb2x1bW47XG4gIH1cblxuICAuaXRlbS1jb250YWluZXIge1xuICAgIGRpc3BsYXk6IGZsZXg7XG4gIH1cblxuICAuaXRlbS1jb250YWluZXIgLml0ZW0ge1xuICB9XG5cbiAgLml0ZW0tY29udGFpbmVyIC5pdGVtLWxlZnQge1xuICAgIGZsZXgtZ3JvdzogMTsgLyogU2V0IHRoZSBtaWRkbGUgZWxlbWVudCB0byBncm93IGFuZCBzdHJldGNoICovXG4gICAgbWFyZ2luLXJpZ2h0OiAwLjVlbTtcbiAgfVxuXG4vKiMgc291cmNlTWFwcGluZ1VSTD1kYXRhOmFwcGxpY2F0aW9uL2pzb247YmFzZTY0LGV5SjJaWEp6YVc5dUlqb3pMQ0p6YjNWeVkyVnpJanBiSW5OeVl5OWxaR2wwYjNJdlkyOXRjRzl1Wlc1MGN5OXNhWE4wWDJWa2FYUmZaR0YwWVY5Mk1pNXpkbVZzZEdVaVhTd2libUZ0WlhNaU9sdGRMQ0p0WVhCd2FXNW5jeUk2SWp0RlFVTkZPMGxCUTBVc2JVSkJRVzFDTzBsQlEyNUNMRmRCUVZjN1JVRkRZanM3UlVGRlFUdEpRVU5GTEdGQlFXRTdTVUZEWWl3NFFrRkJPRUk3U1VGRE9VSXNjMEpCUVhOQ08wVkJRM2hDT3p0RlFVVkJPMGxCUTBVc1lVRkJZVHRGUVVObU96dEZRVVZCTzBWQlEwRTdPMFZCUlVFN1NVRkRSU3haUVVGWkxFVkJRVVVzSzBOQlFTdERPMGxCUXpkRUxHMUNRVUZ0UWp0RlFVTnlRaUlzSW1acGJHVWlPaUp6Y21NdlpXUnBkRzl5TDJOdmJYQnZibVZ1ZEhNdmJHbHpkRjlsWkdsMFgyUmhkR0ZmZGpJdWMzWmxiSFJsSWl3aWMyOTFjbU5sYzBOdmJuUmxiblFpT2xzaVhHNGdJR2x1Y0hWME9tUnBjMkZpYkdWa0lIdGNiaUFnSUNCaVlXTnJaM0p2ZFc1a09pQWpabVpqWTJOak8xeHVJQ0FnSUdOdmJHOXlPaUFqTXpNek8xeHVJQ0I5WEc1Y2JpQWdMbU52Ym5SaGFXNWxjaUI3WEc0Z0lDQWdaR2x6Y0d4aGVUb2dabXhsZUR0Y2JpQWdJQ0JxZFhOMGFXWjVMV052Ym5SbGJuUTZJSE53WVdObExXSmxkSGRsWlc0N1hHNGdJQ0FnWm14bGVDMWthWEpsWTNScGIyNDZJR052YkhWdGJqdGNiaUFnZlZ4dVhHNGdJQzVwZEdWdExXTnZiblJoYVc1bGNpQjdYRzRnSUNBZ1pHbHpjR3hoZVRvZ1pteGxlRHRjYmlBZ2ZWeHVYRzRnSUM1cGRHVnRMV052Ym5SaGFXNWxjaUF1YVhSbGJTQjdYRzRnSUgxY2JseHVJQ0F1YVhSbGJTMWpiMjUwWVdsdVpYSWdMbWwwWlcwdGJHVm1kQ0I3WEc0Z0lDQWdabXhsZUMxbmNtOTNPaUF4T3lBdktpQlRaWFFnZEdobElHMXBaR1JzWlNCbGJHVnRaVzUwSUhSdklHZHliM2NnWVc1a0lITjBjbVYwWTJnZ0tpOWNiaUFnSUNCdFlYSm5hVzR0Y21sbmFIUTZJREF1TldWdE8xeHVJQ0I5WEc0aVhYMD0gKi88L3N0eWxlPlxuXG48c2NyaXB0PlxuICBpbXBvcnQgeyBjb3B5T2JqZWN0LCBpc0RldmljZU1vYmlsZSB9IGZyb20gXCIuLi9saWIvaGVscGVyLmpzXCI7XG4gIGltcG9ydCB7IHRhcCB9IGZyb20gXCJAc3ZlbHRlanMvZ2VzdHVyZXNcIjtcbiAgaW1wb3J0IHsgYWZ0ZXJVcGRhdGUgfSBmcm9tIFwic3ZlbHRlXCI7XG5cbiAgY29uc3QgcG9zc2libGVDb21tYW5kcyA9IHtcbiAgICBub3RoaW5nOiBcIlwiLFxuICAgIG5ld0l0ZW06IFwiV2hlbiBhbiBpdGVtIGlzIGFkZGVkXCJcbiAgfTtcblxuICBjb25zdCBpc01vYmlsZSA9IGlzRGV2aWNlTW9iaWxlKCk7XG4gIGNvbnN0IG9yZGVySGVscGVyVGV4dCA9ICFpc01vYmlsZSA/IFwiZHJhZyBhbmQgZHJvcCB0byBzd2FwXCIgOiBcInRhcCB0byBzd2FwXCI7XG5cbiAgY29uc3QgbmV3Um93ID0ge1xuICAgIGZyb206IFwiXCIsXG4gICAgdG86IFwiXCJcbiAgfTtcbiAgY29uc3QgX3N3YXBJdGVtcyA9IHtcbiAgICBmcm9tOiAtMSxcbiAgICBmcm9tRWxlbWVudDogbnVsbCxcbiAgICB0bzogLTEsXG4gICAgdG9FbGVtZW50OiBudWxsXG4gIH07XG5cbiAgbGV0IGl0ZW1zQ29udGFpbmVyO1xuICBsZXQgbGFzdENtZCA9IHBvc3NpYmxlQ29tbWFuZHMubm90aGluZztcblxuICBsZXQgZW5hYmxlU29ydGFibGUgPSBmYWxzZTtcbiAgZXhwb3J0IGxldCBsaXN0RGF0YTtcbiAgbGV0IHN3YXBJdGVtcyA9IGNvcHlPYmplY3QoX3N3YXBJdGVtcyk7XG5cbiAgYWZ0ZXJVcGRhdGUoKCkgPT4ge1xuICAgIGlmIChsYXN0Q21kID09PSBwb3NzaWJsZUNvbW1hbmRzLm5ld0l0ZW0pIHtcbiAgICAgIC8vIFRoaXMgb25seSB3b3JrcyBmb3IgVjEgZWxlbWVudHNcbiAgICAgIGxldCBub2RlcyA9IGl0ZW1zQ29udGFpbmVyLnF1ZXJ5U2VsZWN0b3JBbGwoXCIuaXRlbS1jb250YWluZXJcIik7XG4gICAgICBub2Rlc1tub2Rlcy5sZW5ndGggLSAxXS5xdWVyeVNlbGVjdG9yKFwiaW5wdXQ6Zmlyc3QtY2hpbGRcIikuZm9jdXMoKTtcbiAgICAgIGxhc3RDbWQgPSBwb3NzaWJsZUNvbW1hbmRzLm5vdGhpbmc7XG4gICAgfVxuICB9KTtcblxuICBmdW5jdGlvbiBhZGQoKSB7XG4gICAgbGlzdERhdGEgPSBsaXN0RGF0YS5jb25jYXQoY29weU9iamVjdChuZXdSb3cpKTtcbiAgICBsYXN0Q21kID0gcG9zc2libGVDb21tYW5kcy5uZXdJdGVtO1xuICB9XG5cbiAgZnVuY3Rpb24gcmVtb3ZlKGxpc3RJdGVtKSB7XG4gICAgbGlzdERhdGEgPSBsaXN0RGF0YS5maWx0ZXIodCA9PiB0ICE9PSBsaXN0SXRlbSk7XG4gICAgaWYgKCFsaXN0RGF0YS5sZW5ndGgpIHtcbiAgICAgIGxpc3REYXRhID0gW2NvcHlPYmplY3QobmV3Um93KV07XG4gICAgfVxuICB9XG5cbiAgZnVuY3Rpb24gcmVtb3ZlQWxsKCkge1xuICAgIGxpc3REYXRhID0gW2NvcHlPYmplY3QobmV3Um93KV07XG4gIH1cblxuICBmdW5jdGlvbiB0b2dnbGVTb3J0YWJsZShldikge1xuICAgIGlmIChsaXN0RGF0YS5sZW5ndGggPD0gMSkge1xuICAgICAgYWxlcnQoXCJub3RoaW5nIHRvIHN3YXBcIik7XG4gICAgICByZXR1cm47XG4gICAgfVxuXG4gICAgZW5hYmxlU29ydGFibGUgPSBlbmFibGVTb3J0YWJsZSA/IGZhbHNlIDogdHJ1ZTtcbiAgICBpZiAoZW5hYmxlU29ydGFibGUpIHtcbiAgICAgIC8vIFJlc2V0IHN3YXBJdGVtc1xuICAgICAgc3dhcEl0ZW1zID0gY29weU9iamVjdChfc3dhcEl0ZW1zKTtcbiAgICB9XG4gIH1cblxuICBmdW5jdGlvbiBkcmFnc3RhcnQoZXYpIHtcbiAgICBzd2FwSXRlbXMgPSBjb3B5T2JqZWN0KF9zd2FwSXRlbXMpO1xuICAgIHN3YXBJdGVtcy5mcm9tID0gZXYudGFyZ2V0LmdldEF0dHJpYnV0ZShcImRhdGEtaW5kZXhcIik7XG4gIH1cblxuICBmdW5jdGlvbiBkcmFnb3Zlcihldikge1xuICAgIGV2LnByZXZlbnREZWZhdWx0KCk7XG4gIH1cblxuICBmdW5jdGlvbiBkcm9wKGV2KSB7XG4gICAgZXYucHJldmVudERlZmF1bHQoKTtcbiAgICBzd2FwSXRlbXMudG8gPSBldi50YXJnZXQuZ2V0QXR0cmlidXRlKFwiZGF0YS1pbmRleFwiKTtcblxuICAgIC8vIFdlIG1pZ2h0IGxhbmQgb24gdGhlIGNoaWxkcmVuLCBsb29rIHVwIGZvciB0aGUgZHJhZ2dhYmxlIGF0dHJpYnV0ZVxuICAgIGlmIChzd2FwSXRlbXMudG8gPT0gbnVsbCkge1xuICAgICAgc3dhcEl0ZW1zLnRvID0gZXYudGFyZ2V0XG4gICAgICAgIC5jbG9zZXN0KFwiW2RyYWdnYWJsZV1cIilcbiAgICAgICAgLmdldEF0dHJpYnV0ZShcImRhdGEtaW5kZXhcIik7XG4gICAgfVxuXG4gICAgbGV0IGEgPSBsaXN0RGF0YVtzd2FwSXRlbXMuZnJvbV07XG4gICAgbGV0IGIgPSBsaXN0RGF0YVtzd2FwSXRlbXMudG9dO1xuICAgIGxpc3REYXRhW3N3YXBJdGVtcy5mcm9tXSA9IGI7XG4gICAgbGlzdERhdGFbc3dhcEl0ZW1zLnRvXSA9IGE7XG4gIH1cblxuICBmdW5jdGlvbiB0YXBIYW5kbGVyKGV2KSB7XG4gICAgZXYucHJldmVudERlZmF1bHQoKTtcblxuICAgIGxldCBpbmRleCA9IGV2LnRhcmdldC5nZXRBdHRyaWJ1dGUoXCJkYXRhLWluZGV4XCIpO1xuXG4gICAgaWYgKGluZGV4ID09PSBudWxsKSB7XG4gICAgICBzd2FwSXRlbXMgPSBjb3B5T2JqZWN0KF9zd2FwSXRlbXMpO1xuICAgICAgcmV0dXJuO1xuICAgIH1cblxuICAgIGlmIChzd2FwSXRlbXMuZnJvbSA9PT0gLTEpIHtcbiAgICAgIHN3YXBJdGVtcy5mcm9tRWxlbWVudCA9IGV2LnRhcmdldDtcbiAgICAgIHN3YXBJdGVtcy5mcm9tRWxlbWVudC5zdHlsZVtcImJvcmRlci1sZWZ0XCJdID0gXCJzb2xpZCBncmVlblwiO1xuICAgICAgc3dhcEl0ZW1zLmZyb20gPSBpbmRleDtcbiAgICAgIHJldHVybjtcbiAgICB9XG5cbiAgICBpZiAoc3dhcEl0ZW1zLmZyb20gPT09IGluZGV4KSB7XG4gICAgICBzd2FwSXRlbXMuZnJvbUVsZW1lbnQuc3R5bGUuYm9yZGVyID0gXCJcIjtcbiAgICAgIHN3YXBJdGVtcyA9IGNvcHlPYmplY3QoX3N3YXBJdGVtcyk7XG4gICAgICByZXR1cm47XG4gICAgfVxuXG4gICAgc3dhcEl0ZW1zLnRvID0gaW5kZXg7XG4gICAgbGV0IGEgPSBsaXN0RGF0YVtzd2FwSXRlbXMuZnJvbV07XG4gICAgbGV0IGIgPSBsaXN0RGF0YVtzd2FwSXRlbXMudG9dO1xuICAgIGxpc3REYXRhW3N3YXBJdGVtcy5mcm9tXSA9IGI7XG4gICAgbGlzdERhdGFbc3dhcEl0ZW1zLnRvXSA9IGE7XG5cbiAgICBzd2FwSXRlbXMuZnJvbUVsZW1lbnQuc3R5bGUuYm9yZGVyID0gXCJcIjtcbiAgICBzd2FwSXRlbXMuZnJvbUVsZW1lbnQuc3R5bGVbXCJib3JkZXItcmFkaXVzXCJdID0gXCIwcHhcIjtcbiAgICBzd2FwSXRlbXMgPSBjb3B5T2JqZWN0KF9zd2FwSXRlbXMpO1xuICB9XG48L3NjcmlwdD5cblxuPGgxPkl0ZW1zPC9oMT5cblxuPGRpdiBiaW5kOnRoaXM9XCJ7aXRlbXNDb250YWluZXJ9XCI+XG4gIHsjaWYgIWVuYWJsZVNvcnRhYmxlfVxuICAgIHsjZWFjaCBsaXN0RGF0YSBhcyBsaXN0SXRlbX1cbiAgICAgIDxkaXYgY2xhc3M9XCJpdGVtLWNvbnRhaW5lciBwdjIgYmIgYi0tYmxhY2stMDVcIj5cbiAgICAgICAgPGRpdiBjbGFzcz1cImZsZXggZmxleC1jb2x1bW4gaXRlbS1sZWZ0XCI+XG4gICAgICAgICAgPGlucHV0XG4gICAgICAgICAgICBwbGFjZWhvbGRlcj1cImZyb21cIlxuICAgICAgICAgICAgYmluZDp2YWx1ZT1cIntsaXN0SXRlbS5mcm9tfVwiXG4gICAgICAgICAgICBjbGFzcz1cIml0ZW0gaXRlbS1sZWZ0XCJcbiAgICAgICAgICAvPlxuICAgICAgICAgIDxpbnB1dFxuICAgICAgICAgICAgcGxhY2Vob2xkZXI9XCJ0b1wiXG4gICAgICAgICAgICBiaW5kOnZhbHVlPVwie2xpc3RJdGVtLnRvfVwiXG4gICAgICAgICAgICBjbGFzcz1cIml0ZW0gaXRlbS1sZWZ0XCJcbiAgICAgICAgICAvPlxuICAgICAgICA8L2Rpdj5cbiAgICAgICAgPGRpdiBjbGFzcz1cImZsZXggZmxleC1jb2x1bW5cIj5cbiAgICAgICAgICA8YnV0dG9uIG9uOmNsaWNrPVwieygpID0+IHJlbW92ZShsaXN0SXRlbSl9XCIgY2xhc3M9XCJpdGVtXCI+eDwvYnV0dG9uPlxuICAgICAgICA8L2Rpdj5cbiAgICAgIDwvZGl2PlxuICAgIHsvZWFjaH1cblxuICAgIDxidXR0b24gb246Y2xpY2s9XCJ7YWRkfVwiPk5ldzwvYnV0dG9uPlxuXG4gICAgPGJ1dHRvbiBvbjpjbGljaz1cIntyZW1vdmVBbGx9XCI+UmVtb3ZlIGFsbDwvYnV0dG9uPlxuXG4gICAgPGJ1dHRvbiBvbjpjbGljaz1cInt0b2dnbGVTb3J0YWJsZX1cIj5DaGFuZ2UgT3JkZXI8L2J1dHRvbj5cbiAgey9pZn1cblxuICB7I2lmIGVuYWJsZVNvcnRhYmxlfVxuICAgIHsjZWFjaCBsaXN0RGF0YSBhcyBsaXN0SXRlbSwgcG9zfVxuICAgICAgPGRpdlxuICAgICAgICBkcmFnZ2FibGU9XCJ0cnVlXCJcbiAgICAgICAgY2xhc3M9XCJkcm9wem9uZSBpdGVtLWNvbnRhaW5lclwiXG4gICAgICAgIGRhdGEtaW5kZXg9XCJ7cG9zfVwiXG4gICAgICAgIG9uOmRyYWdzdGFydD1cIntkcmFnc3RhcnR9XCJcbiAgICAgICAgb246ZHJhZ292ZXI9XCJ7ZHJhZ292ZXJ9XCJcbiAgICAgICAgb246ZHJvcD1cIntkcm9wfVwiXG4gICAgICAgIHVzZTp0YXBcbiAgICAgICAgb246dGFwPVwie3RhcEhhbmRsZXJ9XCJcbiAgICAgID5cbiAgICAgICAgPGlucHV0XG4gICAgICAgICAgcGxhY2Vob2xkZXI9XCJmcm9tXCJcbiAgICAgICAgICBjbGFzcz1cIml0ZW0gaXRlbS1sZWZ0XCJcbiAgICAgICAgICB2YWx1ZT1cIntsaXN0SXRlbS5mcm9tfVwiXG4gICAgICAgICAgZGlzYWJsZWRcbiAgICAgICAgLz5cbiAgICAgICAgPGlucHV0XG4gICAgICAgICAgcGxhY2Vob2xkZXI9XCJ0b1wiXG4gICAgICAgICAgY2xhc3M9XCJpdGVtIGl0ZW0tbGVmdFwiXG4gICAgICAgICAgdmFsdWU9XCJ7bGlzdEl0ZW0udG99XCJcbiAgICAgICAgICBkaXNhYmxlZFxuICAgICAgICAvPlxuICAgICAgPC9kaXY+XG4gICAgey9lYWNofVxuXG4gICAgPGJ1dHRvbiBvbjpjbGljaz1cInt0b2dnbGVTb3J0YWJsZX1cIj5cbiAgICAgIEZpbmlzaGVkIG9yZGVyaW5nPyAoe29yZGVySGVscGVyVGV4dH0pXG4gICAgPC9idXR0b24+XG4gIHsvaWZ9XG48L2Rpdj5cbiJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFDRSxpQ0FBSyxTQUFTLEFBQUMsQ0FBQyxBQUNkLFVBQVUsQ0FBRSxPQUFPLENBQ25CLEtBQUssQ0FBRSxJQUFJLEFBQ2IsQ0FBQyxBQVFELGVBQWUsNEJBQUMsQ0FBQyxBQUNmLE9BQU8sQ0FBRSxJQUFJLEFBQ2YsQ0FBQyxBQUVELDZCQUFlLENBQUMsS0FBSyxjQUFDLENBQUMsQUFDdkIsQ0FBQyxBQUVELDZCQUFlLENBQUMsVUFBVSxjQUFDLENBQUMsQUFDMUIsU0FBUyxDQUFFLENBQUMsQ0FDWixZQUFZLENBQUUsS0FBSyxBQUNyQixDQUFDIn0= */";
	append_dev(document.head, style);
}

function get_each_context$4(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[21] = list[i];
	child_ctx[23] = i;
	return child_ctx;
}

function get_each_context_1$3(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[21] = list[i];
	child_ctx[24] = list;
	child_ctx[25] = i;
	return child_ctx;
}

// (160:2) {#if !enableSortable}
function create_if_block_1$6(ctx) {
	let t0;
	let button0;
	let t2;
	let button1;
	let t4;
	let button2;
	let dispose;
	let each_value_1 = /*listData*/ ctx[0];
	validate_each_argument(each_value_1);
	let each_blocks = [];

	for (let i = 0; i < each_value_1.length; i += 1) {
		each_blocks[i] = create_each_block_1$3(get_each_context_1$3(ctx, each_value_1, i));
	}

	const block = {
		c: function create() {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			t0 = space();
			button0 = element("button");
			button0.textContent = "New";
			t2 = space();
			button1 = element("button");
			button1.textContent = "Remove all";
			t4 = space();
			button2 = element("button");
			button2.textContent = "Change Order";
			add_location(button0, file$n, 180, 4, 5187);
			add_location(button1, file$n, 182, 4, 5230);
			add_location(button2, file$n, 184, 4, 5286);
		},
		m: function mount(target, anchor, remount) {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(target, anchor);
			}

			insert_dev(target, t0, anchor);
			insert_dev(target, button0, anchor);
			insert_dev(target, t2, anchor);
			insert_dev(target, button1, anchor);
			insert_dev(target, t4, anchor);
			insert_dev(target, button2, anchor);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(button0, "click", /*add*/ ctx[4], false, false, false),
				listen_dev(button1, "click", /*removeAll*/ ctx[6], false, false, false),
				listen_dev(button2, "click", /*toggleSortable*/ ctx[7], false, false, false)
			];
		},
		p: function update(ctx, dirty) {
			if (dirty & /*remove, listData*/ 33) {
				each_value_1 = /*listData*/ ctx[0];
				validate_each_argument(each_value_1);
				let i;

				for (i = 0; i < each_value_1.length; i += 1) {
					const child_ctx = get_each_context_1$3(ctx, each_value_1, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
					} else {
						each_blocks[i] = create_each_block_1$3(child_ctx);
						each_blocks[i].c();
						each_blocks[i].m(t0.parentNode, t0);
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
			if (detaching) detach_dev(t0);
			if (detaching) detach_dev(button0);
			if (detaching) detach_dev(t2);
			if (detaching) detach_dev(button1);
			if (detaching) detach_dev(t4);
			if (detaching) detach_dev(button2);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_1$6.name,
		type: "if",
		source: "(160:2) {#if !enableSortable}",
		ctx
	});

	return block;
}

// (161:4) {#each listData as listItem}
function create_each_block_1$3(ctx) {
	let div2;
	let div0;
	let input0;
	let t0;
	let input1;
	let t1;
	let div1;
	let button;
	let dispose;

	function input0_input_handler() {
		/*input0_input_handler*/ ctx[17].call(input0, /*listItem*/ ctx[21]);
	}

	function input1_input_handler() {
		/*input1_input_handler*/ ctx[18].call(input1, /*listItem*/ ctx[21]);
	}

	function click_handler(...args) {
		return /*click_handler*/ ctx[19](/*listItem*/ ctx[21], ...args);
	}

	const block = {
		c: function create() {
			div2 = element("div");
			div0 = element("div");
			input0 = element("input");
			t0 = space();
			input1 = element("input");
			t1 = space();
			div1 = element("div");
			button = element("button");
			button.textContent = "x";
			attr_dev(input0, "placeholder", "from");
			attr_dev(input0, "class", "item item-left svelte-vda4an");
			add_location(input0, file$n, 163, 10, 4750);
			attr_dev(input1, "placeholder", "to");
			attr_dev(input1, "class", "item item-left svelte-vda4an");
			add_location(input1, file$n, 168, 10, 4887);
			attr_dev(div0, "class", "flex flex-column item-left svelte-vda4an");
			add_location(div0, file$n, 162, 8, 4699);
			attr_dev(button, "class", "item svelte-vda4an");
			add_location(button, file$n, 175, 10, 5074);
			attr_dev(div1, "class", "flex flex-column");
			add_location(div1, file$n, 174, 8, 5033);
			attr_dev(div2, "class", "item-container pv2 bb b--black-05 svelte-vda4an");
			add_location(div2, file$n, 161, 6, 4643);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div2, anchor);
			append_dev(div2, div0);
			append_dev(div0, input0);
			set_input_value(input0, /*listItem*/ ctx[21].from);
			append_dev(div0, t0);
			append_dev(div0, input1);
			set_input_value(input1, /*listItem*/ ctx[21].to);
			append_dev(div2, t1);
			append_dev(div2, div1);
			append_dev(div1, button);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(input0, "input", input0_input_handler),
				listen_dev(input1, "input", input1_input_handler),
				listen_dev(button, "click", click_handler, false, false, false)
			];
		},
		p: function update(new_ctx, dirty) {
			ctx = new_ctx;

			if (dirty & /*listData*/ 1 && input0.value !== /*listItem*/ ctx[21].from) {
				set_input_value(input0, /*listItem*/ ctx[21].from);
			}

			if (dirty & /*listData*/ 1 && input1.value !== /*listItem*/ ctx[21].to) {
				set_input_value(input1, /*listItem*/ ctx[21].to);
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div2);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block_1$3.name,
		type: "each",
		source: "(161:4) {#each listData as listItem}",
		ctx
	});

	return block;
}

// (188:2) {#if enableSortable}
function create_if_block$b(ctx) {
	let t0;
	let button;
	let dispose;
	let each_value = /*listData*/ ctx[0];
	validate_each_argument(each_value);
	let each_blocks = [];

	for (let i = 0; i < each_value.length; i += 1) {
		each_blocks[i] = create_each_block$4(get_each_context$4(ctx, each_value, i));
	}

	const block = {
		c: function create() {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			t0 = space();
			button = element("button");

			button.textContent = `
      Finished ordering? (${/*orderHelperText*/ ctx[3]})
    `;

			add_location(button, file$n, 214, 4, 5972);
		},
		m: function mount(target, anchor, remount) {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(target, anchor);
			}

			insert_dev(target, t0, anchor);
			insert_dev(target, button, anchor);
			if (remount) dispose();
			dispose = listen_dev(button, "click", /*toggleSortable*/ ctx[7], false, false, false);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*dragstart, dragover, drop, tapHandler, listData*/ 1793) {
				each_value = /*listData*/ ctx[0];
				validate_each_argument(each_value);
				let i;

				for (i = 0; i < each_value.length; i += 1) {
					const child_ctx = get_each_context$4(ctx, each_value, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
					} else {
						each_blocks[i] = create_each_block$4(child_ctx);
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
			if (detaching) detach_dev(button);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$b.name,
		type: "if",
		source: "(188:2) {#if enableSortable}",
		ctx
	});

	return block;
}

// (189:4) {#each listData as listItem, pos}
function create_each_block$4(ctx) {
	let div;
	let input0;
	let input0_value_value;
	let t;
	let input1;
	let input1_value_value;
	let div_data_index_value;
	let tap_action;
	let dispose;

	const block = {
		c: function create() {
			div = element("div");
			input0 = element("input");
			t = space();
			input1 = element("input");
			attr_dev(input0, "placeholder", "from");
			attr_dev(input0, "class", "item item-left svelte-vda4an");
			input0.value = input0_value_value = /*listItem*/ ctx[21].from;
			input0.disabled = true;
			add_location(input0, file$n, 199, 8, 5672);
			attr_dev(input1, "placeholder", "to");
			attr_dev(input1, "class", "item item-left svelte-vda4an");
			input1.value = input1_value_value = /*listItem*/ ctx[21].to;
			input1.disabled = true;
			add_location(input1, file$n, 205, 8, 5813);
			attr_dev(div, "draggable", "true");
			attr_dev(div, "class", "dropzone item-container svelte-vda4an");
			attr_dev(div, "data-index", div_data_index_value = /*pos*/ ctx[23]);
			add_location(div, file$n, 189, 6, 5420);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div, anchor);
			append_dev(div, input0);
			append_dev(div, t);
			append_dev(div, input1);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(div, "dragstart", /*dragstart*/ ctx[8], false, false, false),
				listen_dev(div, "dragover", dragover$1, false, false, false),
				listen_dev(div, "drop", /*drop*/ ctx[9], false, false, false),
				action_destroyer(tap_action = tap.call(null, div)),
				listen_dev(div, "tap", /*tapHandler*/ ctx[10], false, false, false)
			];
		},
		p: function update(ctx, dirty) {
			if (dirty & /*listData*/ 1 && input0_value_value !== (input0_value_value = /*listItem*/ ctx[21].from) && input0.value !== input0_value_value) {
				prop_dev(input0, "value", input0_value_value);
			}

			if (dirty & /*listData*/ 1 && input1_value_value !== (input1_value_value = /*listItem*/ ctx[21].to) && input1.value !== input1_value_value) {
				prop_dev(input1, "value", input1_value_value);
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block$4.name,
		type: "each",
		source: "(189:4) {#each listData as listItem, pos}",
		ctx
	});

	return block;
}

function create_fragment$o(ctx) {
	let h1;
	let t1;
	let div;
	let t2;
	let if_block0 = !/*enableSortable*/ ctx[2] && create_if_block_1$6(ctx);
	let if_block1 = /*enableSortable*/ ctx[2] && create_if_block$b(ctx);

	const block = {
		c: function create() {
			h1 = element("h1");
			h1.textContent = "Items";
			t1 = space();
			div = element("div");
			if (if_block0) if_block0.c();
			t2 = space();
			if (if_block1) if_block1.c();
			add_location(h1, file$n, 156, 0, 4529);
			add_location(div, file$n, 158, 0, 4545);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, h1, anchor);
			insert_dev(target, t1, anchor);
			insert_dev(target, div, anchor);
			if (if_block0) if_block0.m(div, null);
			append_dev(div, t2);
			if (if_block1) if_block1.m(div, null);
			/*div_binding*/ ctx[20](div);
		},
		p: function update(ctx, [dirty]) {
			if (!/*enableSortable*/ ctx[2]) {
				if (if_block0) {
					if_block0.p(ctx, dirty);
				} else {
					if_block0 = create_if_block_1$6(ctx);
					if_block0.c();
					if_block0.m(div, t2);
				}
			} else if (if_block0) {
				if_block0.d(1);
				if_block0 = null;
			}

			if (/*enableSortable*/ ctx[2]) {
				if (if_block1) {
					if_block1.p(ctx, dirty);
				} else {
					if_block1 = create_if_block$b(ctx);
					if_block1.c();
					if_block1.m(div, null);
				}
			} else if (if_block1) {
				if_block1.d(1);
				if_block1 = null;
			}
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(h1);
			if (detaching) detach_dev(t1);
			if (detaching) detach_dev(div);
			if (if_block0) if_block0.d();
			if (if_block1) if_block1.d();
			/*div_binding*/ ctx[20](null);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$o.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function dragover$1(ev) {
	ev.preventDefault();
}

function instance$o($$self, $$props, $$invalidate) {
	const possibleCommands = {
		nothing: "",
		newItem: "When an item is added"
	};

	const isMobile = isDeviceMobile();
	const orderHelperText = !isMobile ? "drag and drop to swap" : "tap to swap";
	const newRow = { from: "", to: "" };

	const _swapItems = {
		from: -1,
		fromElement: null,
		to: -1,
		toElement: null
	};

	let itemsContainer;
	let lastCmd = possibleCommands.nothing;
	let enableSortable = false;
	let { listData } = $$props;
	let swapItems = copyObject(_swapItems);

	afterUpdate(() => {
		if (lastCmd === possibleCommands.newItem) {
			// This only works for V1 elements
			let nodes = itemsContainer.querySelectorAll(".item-container");

			nodes[nodes.length - 1].querySelector("input:first-child").focus();
			lastCmd = possibleCommands.nothing;
		}
	});

	function add() {
		$$invalidate(0, listData = listData.concat(copyObject(newRow)));
		lastCmd = possibleCommands.newItem;
	}

	function remove(listItem) {
		$$invalidate(0, listData = listData.filter(t => t !== listItem));

		if (!listData.length) {
			$$invalidate(0, listData = [copyObject(newRow)]);
		}
	}

	function removeAll() {
		$$invalidate(0, listData = [copyObject(newRow)]);
	}

	function toggleSortable(ev) {
		if (listData.length <= 1) {
			alert("nothing to swap");
			return;
		}

		$$invalidate(2, enableSortable = enableSortable ? false : true);

		if (enableSortable) {
			// Reset swapItems
			swapItems = copyObject(_swapItems);
		}
	}

	function dragstart(ev) {
		swapItems = copyObject(_swapItems);
		swapItems.from = ev.target.getAttribute("data-index");
	}

	function drop(ev) {
		ev.preventDefault();
		swapItems.to = ev.target.getAttribute("data-index");

		// We might land on the children, look up for the draggable attribute
		if (swapItems.to == null) {
			swapItems.to = ev.target.closest("[draggable]").getAttribute("data-index");
		}

		let a = listData[swapItems.from];
		let b = listData[swapItems.to];
		$$invalidate(0, listData[swapItems.from] = b, listData);
		$$invalidate(0, listData[swapItems.to] = a, listData);
	}

	function tapHandler(ev) {
		ev.preventDefault();
		let index = ev.target.getAttribute("data-index");

		if (index === null) {
			swapItems = copyObject(_swapItems);
			return;
		}

		if (swapItems.from === -1) {
			swapItems.fromElement = ev.target;
			swapItems.fromElement.style["border-left"] = "solid green";
			swapItems.from = index;
			return;
		}

		if (swapItems.from === index) {
			swapItems.fromElement.style.border = "";
			swapItems = copyObject(_swapItems);
			return;
		}

		swapItems.to = index;
		let a = listData[swapItems.from];
		let b = listData[swapItems.to];
		$$invalidate(0, listData[swapItems.from] = b, listData);
		$$invalidate(0, listData[swapItems.to] = a, listData);
		swapItems.fromElement.style.border = "";
		swapItems.fromElement.style["border-radius"] = "0px";
		swapItems = copyObject(_swapItems);
	}

	const writable_props = ["listData"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_edit_data_v2> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_edit_data_v2", $$slots, []);

	function input0_input_handler(listItem) {
		listItem.from = this.value;
		$$invalidate(0, listData);
	}

	function input1_input_handler(listItem) {
		listItem.to = this.value;
		$$invalidate(0, listData);
	}

	const click_handler = listItem => remove(listItem);

	function div_binding($$value) {
		binding_callbacks[$$value ? "unshift" : "push"](() => {
			$$invalidate(1, itemsContainer = $$value);
		});
	}

	$$self.$set = $$props => {
		if ("listData" in $$props) $$invalidate(0, listData = $$props.listData);
	};

	$$self.$capture_state = () => ({
		copyObject,
		isDeviceMobile,
		tap,
		afterUpdate,
		possibleCommands,
		isMobile,
		orderHelperText,
		newRow,
		_swapItems,
		itemsContainer,
		lastCmd,
		enableSortable,
		listData,
		swapItems,
		add,
		remove,
		removeAll,
		toggleSortable,
		dragstart,
		dragover: dragover$1,
		drop,
		tapHandler
	});

	$$self.$inject_state = $$props => {
		if ("itemsContainer" in $$props) $$invalidate(1, itemsContainer = $$props.itemsContainer);
		if ("lastCmd" in $$props) lastCmd = $$props.lastCmd;
		if ("enableSortable" in $$props) $$invalidate(2, enableSortable = $$props.enableSortable);
		if ("listData" in $$props) $$invalidate(0, listData = $$props.listData);
		if ("swapItems" in $$props) swapItems = $$props.swapItems;
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [
		listData,
		itemsContainer,
		enableSortable,
		orderHelperText,
		add,
		remove,
		removeAll,
		toggleSortable,
		dragstart,
		drop,
		tapHandler,
		lastCmd,
		swapItems,
		possibleCommands,
		isMobile,
		newRow,
		_swapItems,
		input0_input_handler,
		input1_input_handler,
		click_handler,
		div_binding
	];
}

class List_edit_data_v2 extends SvelteComponentDev {
	constructor(options) {
		super(options);
		if (!document.getElementById("svelte-vda4an-style")) add_css$3();
		init(this, options, instance$o, create_fragment$o, safe_not_equal, { listData: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_edit_data_v2",
			options,
			id: create_fragment$o.name
		});

		const { ctx } = this.$$;
		const props = options.props || {};

		if (/*listData*/ ctx[0] === undefined && !("listData" in props)) {
			console.warn("<List_edit_data_v2> was created without expected prop 'listData'");
		}
	}

	get listData() {
		throw new Error("<List_edit_data_v2>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set listData(value) {
		throw new Error("<List_edit_data_v2>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list_edit_data_v3_split.svelte generated by Svelte v3.20.1 */
const file$o = "src/editor/components/list_edit_data_v3_split.svelte";

function add_css$4() {
	var style = element("style");
	style.id = "svelte-33y8ka-style";
	style.textContent = "input.svelte-33y8ka:disabled{background:#ffcccc;color:#333}\n/*# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibGlzdF9lZGl0X2RhdGFfdjNfc3BsaXQuc3ZlbHRlIiwic291cmNlcyI6WyJsaXN0X2VkaXRfZGF0YV92M19zcGxpdC5zdmVsdGUiXSwic291cmNlc0NvbnRlbnQiOlsiPHN0eWxlPlxuICBpbnB1dDpkaXNhYmxlZCB7XG4gICAgYmFja2dyb3VuZDogI2ZmY2NjYztcbiAgICBjb2xvcjogIzMzMztcbiAgfVxuXG4vKiMgc291cmNlTWFwcGluZ1VSTD1kYXRhOmFwcGxpY2F0aW9uL2pzb247YmFzZTY0LGV5SjJaWEp6YVc5dUlqb3pMQ0p6YjNWeVkyVnpJanBiSW5OeVl5OWxaR2wwYjNJdlkyOXRjRzl1Wlc1MGN5OXNhWE4wWDJWa2FYUmZaR0YwWVY5Mk0xOXpjR3hwZEM1emRtVnNkR1VpWFN3aWJtRnRaWE1pT2x0ZExDSnRZWEJ3YVc1bmN5STZJanRGUVVORk8wbEJRMFVzYlVKQlFXMUNPMGxCUTI1Q0xGZEJRVmM3UlVGRFlpSXNJbVpwYkdVaU9pSnpjbU12WldScGRHOXlMMk52YlhCdmJtVnVkSE12YkdsemRGOWxaR2wwWDJSaGRHRmZkak5mYzNCc2FYUXVjM1psYkhSbElpd2ljMjkxY21ObGMwTnZiblJsYm5RaU9sc2lYRzRnSUdsdWNIVjBPbVJwYzJGaWJHVmtJSHRjYmlBZ0lDQmlZV05yWjNKdmRXNWtPaUFqWm1aalkyTmpPMXh1SUNBZ0lHTnZiRzl5T2lBak16TXpPMXh1SUNCOVhHNGlYWDA9ICovPC9zdHlsZT5cblxuPHNjcmlwdD5cbiAgaW1wb3J0IHsgY3JlYXRlRXZlbnREaXNwYXRjaGVyIH0gZnJvbSBcInN2ZWx0ZVwiO1xuICBleHBvcnQgbGV0IGRpc2FibGVkID0gdW5kZWZpbmVkO1xuICBleHBvcnQgbGV0IGluZGV4O1xuICBleHBvcnQgbGV0IHNwbGl0SW5kZXg7XG4gIGV4cG9ydCBsZXQgc3BsaXQgPSB7XG4gICAgdGltZTogXCJcIixcbiAgICBkaXN0YW5jZTogMCxcbiAgICBwNTAwOiBcIlwiLFxuICAgIHNwbTogMFxuICB9O1xuICBjb25zdCBkaXNwYXRjaCA9IGNyZWF0ZUV2ZW50RGlzcGF0Y2hlcigpO1xuICAvLyBUT0RPIGR1cGxpY2F0ZVxuICBjb25zdCBjbGljayA9ICgpID0+IHtcbiAgICBzcGxpdCA9IHtcbiAgICAgIHRpbWU6IFwiXCIsXG4gICAgICBkaXN0YW5jZTogMCxcbiAgICAgIHA1MDA6IFwiXCIsXG4gICAgICBzcG06IDBcbiAgICB9O1xuICAgIGRpc3BhdGNoKFwiY2xpY2tcIiwge1xuICAgICAgaW5kZXgsXG4gICAgICBzcGxpdEluZGV4XG4gICAgfSk7XG4gIH07XG48L3NjcmlwdD5cblxuPGRpdiBjbGFzcz1cImZsZXggcHYwXCI+XG4gIDxkaXYgY2xhc3M9XCJ3LTI1IHBhMCBtcjJcIj5cbiAgICA8aW5wdXRcbiAgICAgIHBsYWNlaG9sZGVyPVwidGltZVwiXG4gICAgICB7ZGlzYWJsZWR9XG4gICAgICBiaW5kOnZhbHVlPVwie3NwbGl0LnRpbWV9XCJcbiAgICAgIGNsYXNzPVwidy0xMDBcIlxuICAgIC8+XG4gIDwvZGl2PlxuICA8ZGl2IGNsYXNzPVwidy0yNSBwYTAgbXIyXCI+XG4gICAgPGlucHV0XG4gICAgICBwbGFjZWhvbGRlcj1cImRpc3RhbmNlXCJcbiAgICAgIHR5cGU9XCJudW1iZXJcIlxuICAgICAge2Rpc2FibGVkfVxuICAgICAgYmluZDp2YWx1ZT1cIntzcGxpdC5kaXN0YW5jZX1cIlxuICAgICAgY2xhc3M9XCJ3LTEwMFwiXG4gICAgLz5cbiAgPC9kaXY+XG4gIDxkaXYgY2xhc3M9XCJ3LTI1IHBhMCBtcjJcIj5cbiAgICA8aW5wdXRcbiAgICAgIHBsYWNlaG9sZGVyPVwiLzUwMG1cIlxuICAgICAge2Rpc2FibGVkfVxuICAgICAgYmluZDp2YWx1ZT1cIntzcGxpdC5wNTAwfVwiXG4gICAgICBjbGFzcz1cInctMTAwXCJcbiAgICAvPlxuICA8L2Rpdj5cbiAgPGRpdiBjbGFzcz1cInctMjUgcGEwIG1yMlwiPlxuICAgIDxpbnB1dFxuICAgICAgcGxhY2Vob2xkZXI9XCJzcG1cIlxuICAgICAgdHlwZT1cIm51bWJlclwiXG4gICAgICB7ZGlzYWJsZWR9XG4gICAgICBiaW5kOnZhbHVlPVwie3NwbGl0LnNwbX1cIlxuICAgICAgY2xhc3M9XCJ3LTEwMFwiXG4gICAgLz5cbiAgPC9kaXY+XG4gIDxkaXYgY2xhc3M9XCJwYTBcIj5cbiAgICB7I2lmIGRpc2FibGVkID09PSB1bmRlZmluZWR9XG4gICAgICA8YnV0dG9uIG9uOmNsaWNrPVwie2NsaWNrfVwiIGNsYXNzPVwiaXRlbVwiPng8L2J1dHRvbj5cbiAgICB7OmVsc2V9XG4gICAgICA8c3Bhbj4mbmJzcDs8L3NwYW4+XG4gICAgey9pZn1cbiAgPC9kaXY+XG48L2Rpdj5cbiJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFDRSxtQkFBSyxTQUFTLEFBQUMsQ0FBQyxBQUNkLFVBQVUsQ0FBRSxPQUFPLENBQ25CLEtBQUssQ0FBRSxJQUFJLEFBQ2IsQ0FBQyJ9 */";
	append_dev(document.head, style);
}

// (74:4) {:else}
function create_else_block$4(ctx) {
	let span;

	const block = {
		c: function create() {
			span = element("span");
			span.textContent = " ";
			add_location(span, file$o, 74, 6, 1801);
		},
		m: function mount(target, anchor) {
			insert_dev(target, span, anchor);
		},
		p: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(span);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_else_block$4.name,
		type: "else",
		source: "(74:4) {:else}",
		ctx
	});

	return block;
}

// (72:4) {#if disabled === undefined}
function create_if_block$c(ctx) {
	let button;
	let dispose;

	const block = {
		c: function create() {
			button = element("button");
			button.textContent = "x";
			attr_dev(button, "class", "item");
			add_location(button, file$o, 72, 6, 1732);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, button, anchor);
			if (remount) dispose();
			dispose = listen_dev(button, "click", /*click*/ ctx[2], false, false, false);
		},
		p: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(button);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$c.name,
		type: "if",
		source: "(72:4) {#if disabled === undefined}",
		ctx
	});

	return block;
}

function create_fragment$p(ctx) {
	let div5;
	let div0;
	let input0;
	let t0;
	let div1;
	let input1;
	let input1_updating = false;
	let t1;
	let div2;
	let input2;
	let t2;
	let div3;
	let input3;
	let input3_updating = false;
	let t3;
	let div4;
	let dispose;

	function input1_input_handler() {
		input1_updating = true;
		/*input1_input_handler*/ ctx[7].call(input1);
	}

	function input3_input_handler() {
		input3_updating = true;
		/*input3_input_handler*/ ctx[9].call(input3);
	}

	function select_block_type(ctx, dirty) {
		if (/*disabled*/ ctx[1] === undefined) return create_if_block$c;
		return create_else_block$4;
	}

	let current_block_type = select_block_type(ctx);
	let if_block = current_block_type(ctx);

	const block = {
		c: function create() {
			div5 = element("div");
			div0 = element("div");
			input0 = element("input");
			t0 = space();
			div1 = element("div");
			input1 = element("input");
			t1 = space();
			div2 = element("div");
			input2 = element("input");
			t2 = space();
			div3 = element("div");
			input3 = element("input");
			t3 = space();
			div4 = element("div");
			if_block.c();
			attr_dev(input0, "placeholder", "time");
			input0.disabled = /*disabled*/ ctx[1];
			attr_dev(input0, "class", "w-100 svelte-33y8ka");
			add_location(input0, file$o, 37, 4, 1059);
			attr_dev(div0, "class", "w-25 pa0 mr2");
			add_location(div0, file$o, 36, 2, 1028);
			attr_dev(input1, "placeholder", "distance");
			attr_dev(input1, "type", "number");
			input1.disabled = /*disabled*/ ctx[1];
			attr_dev(input1, "class", "w-100 svelte-33y8ka");
			add_location(input1, file$o, 45, 4, 1209);
			attr_dev(div1, "class", "w-25 pa0 mr2");
			add_location(div1, file$o, 44, 2, 1178);
			attr_dev(input2, "placeholder", "/500m");
			input2.disabled = /*disabled*/ ctx[1];
			attr_dev(input2, "class", "w-100 svelte-33y8ka");
			add_location(input2, file$o, 54, 4, 1387);
			attr_dev(div2, "class", "w-25 pa0 mr2");
			add_location(div2, file$o, 53, 2, 1356);
			attr_dev(input3, "placeholder", "spm");
			attr_dev(input3, "type", "number");
			input3.disabled = /*disabled*/ ctx[1];
			attr_dev(input3, "class", "w-100 svelte-33y8ka");
			add_location(input3, file$o, 62, 4, 1538);
			attr_dev(div3, "class", "w-25 pa0 mr2");
			add_location(div3, file$o, 61, 2, 1507);
			attr_dev(div4, "class", "pa0");
			add_location(div4, file$o, 70, 2, 1675);
			attr_dev(div5, "class", "flex pv0");
			add_location(div5, file$o, 35, 0, 1003);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div5, anchor);
			append_dev(div5, div0);
			append_dev(div0, input0);
			set_input_value(input0, /*split*/ ctx[0].time);
			append_dev(div5, t0);
			append_dev(div5, div1);
			append_dev(div1, input1);
			set_input_value(input1, /*split*/ ctx[0].distance);
			append_dev(div5, t1);
			append_dev(div5, div2);
			append_dev(div2, input2);
			set_input_value(input2, /*split*/ ctx[0].p500);
			append_dev(div5, t2);
			append_dev(div5, div3);
			append_dev(div3, input3);
			set_input_value(input3, /*split*/ ctx[0].spm);
			append_dev(div5, t3);
			append_dev(div5, div4);
			if_block.m(div4, null);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(input0, "input", /*input0_input_handler*/ ctx[6]),
				listen_dev(input1, "input", input1_input_handler),
				listen_dev(input2, "input", /*input2_input_handler*/ ctx[8]),
				listen_dev(input3, "input", input3_input_handler)
			];
		},
		p: function update(ctx, [dirty]) {
			if (dirty & /*disabled*/ 2) {
				prop_dev(input0, "disabled", /*disabled*/ ctx[1]);
			}

			if (dirty & /*split*/ 1 && input0.value !== /*split*/ ctx[0].time) {
				set_input_value(input0, /*split*/ ctx[0].time);
			}

			if (dirty & /*disabled*/ 2) {
				prop_dev(input1, "disabled", /*disabled*/ ctx[1]);
			}

			if (!input1_updating && dirty & /*split*/ 1) {
				set_input_value(input1, /*split*/ ctx[0].distance);
			}

			input1_updating = false;

			if (dirty & /*disabled*/ 2) {
				prop_dev(input2, "disabled", /*disabled*/ ctx[1]);
			}

			if (dirty & /*split*/ 1 && input2.value !== /*split*/ ctx[0].p500) {
				set_input_value(input2, /*split*/ ctx[0].p500);
			}

			if (dirty & /*disabled*/ 2) {
				prop_dev(input3, "disabled", /*disabled*/ ctx[1]);
			}

			if (!input3_updating && dirty & /*split*/ 1) {
				set_input_value(input3, /*split*/ ctx[0].spm);
			}

			input3_updating = false;

			if (current_block_type === (current_block_type = select_block_type(ctx)) && if_block) {
				if_block.p(ctx, dirty);
			} else {
				if_block.d(1);
				if_block = current_block_type(ctx);

				if (if_block) {
					if_block.c();
					if_block.m(div4, null);
				}
			}
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(div5);
			if_block.d();
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$p.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$p($$self, $$props, $$invalidate) {
	let { disabled = undefined } = $$props;
	let { index } = $$props;
	let { splitIndex } = $$props;
	let { split = { time: "", distance: 0, p500: "", spm: 0 } } = $$props;
	const dispatch = createEventDispatcher();

	// TODO duplicate
	const click = () => {
		$$invalidate(0, split = { time: "", distance: 0, p500: "", spm: 0 });
		dispatch("click", { index, splitIndex });
	};

	const writable_props = ["disabled", "index", "splitIndex", "split"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_edit_data_v3_split> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_edit_data_v3_split", $$slots, []);

	function input0_input_handler() {
		split.time = this.value;
		$$invalidate(0, split);
	}

	function input1_input_handler() {
		split.distance = to_number(this.value);
		$$invalidate(0, split);
	}

	function input2_input_handler() {
		split.p500 = this.value;
		$$invalidate(0, split);
	}

	function input3_input_handler() {
		split.spm = to_number(this.value);
		$$invalidate(0, split);
	}

	$$self.$set = $$props => {
		if ("disabled" in $$props) $$invalidate(1, disabled = $$props.disabled);
		if ("index" in $$props) $$invalidate(3, index = $$props.index);
		if ("splitIndex" in $$props) $$invalidate(4, splitIndex = $$props.splitIndex);
		if ("split" in $$props) $$invalidate(0, split = $$props.split);
	};

	$$self.$capture_state = () => ({
		createEventDispatcher,
		disabled,
		index,
		splitIndex,
		split,
		dispatch,
		click
	});

	$$self.$inject_state = $$props => {
		if ("disabled" in $$props) $$invalidate(1, disabled = $$props.disabled);
		if ("index" in $$props) $$invalidate(3, index = $$props.index);
		if ("splitIndex" in $$props) $$invalidate(4, splitIndex = $$props.splitIndex);
		if ("split" in $$props) $$invalidate(0, split = $$props.split);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [
		split,
		disabled,
		click,
		index,
		splitIndex,
		dispatch,
		input0_input_handler,
		input1_input_handler,
		input2_input_handler,
		input3_input_handler
	];
}

class List_edit_data_v3_split extends SvelteComponentDev {
	constructor(options) {
		super(options);
		if (!document.getElementById("svelte-33y8ka-style")) add_css$4();

		init(this, options, instance$p, create_fragment$p, safe_not_equal, {
			disabled: 1,
			index: 3,
			splitIndex: 4,
			split: 0
		});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_edit_data_v3_split",
			options,
			id: create_fragment$p.name
		});

		const { ctx } = this.$$;
		const props = options.props || {};

		if (/*index*/ ctx[3] === undefined && !("index" in props)) {
			console.warn("<List_edit_data_v3_split> was created without expected prop 'index'");
		}

		if (/*splitIndex*/ ctx[4] === undefined && !("splitIndex" in props)) {
			console.warn("<List_edit_data_v3_split> was created without expected prop 'splitIndex'");
		}
	}

	get disabled() {
		throw new Error("<List_edit_data_v3_split>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set disabled(value) {
		throw new Error("<List_edit_data_v3_split>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	get index() {
		throw new Error("<List_edit_data_v3_split>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set index(value) {
		throw new Error("<List_edit_data_v3_split>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	get splitIndex() {
		throw new Error("<List_edit_data_v3_split>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set splitIndex(value) {
		throw new Error("<List_edit_data_v3_split>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	get split() {
		throw new Error("<List_edit_data_v3_split>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set split(value) {
		throw new Error("<List_edit_data_v3_split>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list_edit_data_v3_record.svelte generated by Svelte v3.20.1 */
const file$p = "src/editor/components/list_edit_data_v3_record.svelte";

function add_css$5() {
	var style = element("style");
	style.id = "svelte-1dchow9-style";
	style.textContent = "input.svelte-1dchow9:disabled{background:#ffcccc;color:#333}.item-container.svelte-1dchow9{display:flex}.nodrag.svelte-1dchow9{-webkit-touch-callout:none;-webkit-user-select:none;-moz-user-select:none;-ms-user-select:none;user-select:none}\n/*# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibGlzdF9lZGl0X2RhdGFfdjNfcmVjb3JkLnN2ZWx0ZSIsInNvdXJjZXMiOlsibGlzdF9lZGl0X2RhdGFfdjNfcmVjb3JkLnN2ZWx0ZSJdLCJzb3VyY2VzQ29udGVudCI6WyI8c3R5bGU+XG4gIGlucHV0OmRpc2FibGVkIHtcbiAgICBiYWNrZ3JvdW5kOiAjZmZjY2NjO1xuICAgIGNvbG9yOiAjMzMzO1xuICB9XG5cbiAgLml0ZW0tY29udGFpbmVyIHtcbiAgICBkaXNwbGF5OiBmbGV4O1xuICB9XG5cbiAgLm5vZHJhZyB7XG4gICAgLXdlYmtpdC10b3VjaC1jYWxsb3V0OiBub25lO1xuICAgIC13ZWJraXQtdXNlci1zZWxlY3Q6IG5vbmU7XG4gICAgLW1vei11c2VyLXNlbGVjdDogbm9uZTtcbiAgICAtbXMtdXNlci1zZWxlY3Q6IG5vbmU7XG4gICAgdXNlci1zZWxlY3Q6IG5vbmU7XG4gIH1cblxuLyojIHNvdXJjZU1hcHBpbmdVUkw9ZGF0YTphcHBsaWNhdGlvbi9qc29uO2Jhc2U2NCxleUoyWlhKemFXOXVJam96TENKemIzVnlZMlZ6SWpwYkluTnlZeTlsWkdsMGIzSXZZMjl0Y0c5dVpXNTBjeTlzYVhOMFgyVmthWFJmWkdGMFlWOTJNMTl5WldOdmNtUXVjM1psYkhSbElsMHNJbTVoYldWeklqcGJYU3dpYldGd2NHbHVaM01pT2lJN1JVRkRSVHRKUVVORkxHMUNRVUZ0UWp0SlFVTnVRaXhYUVVGWE8wVkJRMkk3TzBWQlJVRTdTVUZEUlN4aFFVRmhPMFZCUTJZN08wVkJSVUU3U1VGRFJTd3lRa0ZCTWtJN1NVRkRNMElzZVVKQlFYbENPMGxCUlhwQ0xITkNRVUZ6UWp0SlFVTjBRaXh4UWtGQmNVSTdTVUZEY2tJc2FVSkJRV2xDTzBWQlEyNUNJaXdpWm1sc1pTSTZJbk55WXk5bFpHbDBiM0l2WTI5dGNHOXVaVzUwY3k5c2FYTjBYMlZrYVhSZlpHRjBZVjkyTTE5eVpXTnZjbVF1YzNabGJIUmxJaXdpYzI5MWNtTmxjME52Ym5SbGJuUWlPbHNpWEc0Z0lHbHVjSFYwT21ScGMyRmliR1ZrSUh0Y2JpQWdJQ0JpWVdOclozSnZkVzVrT2lBalptWmpZMk5qTzF4dUlDQWdJR052Ykc5eU9pQWpNek16TzF4dUlDQjlYRzVjYmlBZ0xtbDBaVzB0WTI5dWRHRnBibVZ5SUh0Y2JpQWdJQ0JrYVhOd2JHRjVPaUJtYkdWNE8xeHVJQ0I5WEc1Y2JpQWdMbTV2WkhKaFp5QjdYRzRnSUNBZ0xYZGxZbXRwZEMxMGIzVmphQzFqWVd4c2IzVjBPaUJ1YjI1bE8xeHVJQ0FnSUMxM1pXSnJhWFF0ZFhObGNpMXpaV3hsWTNRNklHNXZibVU3WEc0Z0lDQWdMV3RvZEcxc0xYVnpaWEl0YzJWc1pXTjBPaUJ1YjI1bE8xeHVJQ0FnSUMxdGIzb3RkWE5sY2kxelpXeGxZM1E2SUc1dmJtVTdYRzRnSUNBZ0xXMXpMWFZ6WlhJdGMyVnNaV04wT2lCdWIyNWxPMXh1SUNBZ0lIVnpaWEl0YzJWc1pXTjBPaUJ1YjI1bE8xeHVJQ0I5WEc0aVhYMD0gKi88L3N0eWxlPlxuXG48c2NyaXB0PlxuICBpbXBvcnQgeyBjb3B5T2JqZWN0LCBpc0RldmljZU1vYmlsZSB9IGZyb20gXCIuLi9saWIvaGVscGVyLmpzXCI7XG4gIGltcG9ydCB7IHRhcCB9IGZyb20gXCJAc3ZlbHRlanMvZ2VzdHVyZXNcIjtcbiAgaW1wb3J0IHsgYWZ0ZXJVcGRhdGUgfSBmcm9tIFwic3ZlbHRlXCI7XG5cbiAgaW1wb3J0IFNwbGl0IGZyb20gXCIuL2xpc3RfZWRpdF9kYXRhX3YzX3NwbGl0LnN2ZWx0ZVwiO1xuXG4gIGV4cG9ydCBsZXQgaW5kZXg7XG4gIGV4cG9ydCBsZXQgcmVjb3JkO1xuICBleHBvcnQgbGV0IGRpc2FibGVkID0gdW5kZWZpbmVkO1xuXG4gIGltcG9ydCB7IGNyZWF0ZUV2ZW50RGlzcGF0Y2hlciB9IGZyb20gXCJzdmVsdGVcIjtcblxuICBjb25zdCBkaXNwYXRjaCA9IGNyZWF0ZUV2ZW50RGlzcGF0Y2hlcigpO1xuXG4gIGNvbnN0IG5ld1JvdyA9IHsgdGltZTogXCJcIiwgZGlzdGFuY2U6IDAsIHA1MDA6IFwiXCIsIHNwbTogMCB9O1xuXG4gIGZ1bmN0aW9uIGFkZFNwbGl0KCkge1xuICAgIHJlY29yZC5zcGxpdHMucHVzaChjb3B5T2JqZWN0KG5ld1JvdykpO1xuICAgIHJlY29yZCA9IHJlY29yZDtcbiAgfVxuXG4gIGZ1bmN0aW9uIHJlbW92ZVNwbGl0KGV2ZW50KSB7XG4gICAgY29uc3Qgc3BsaXRJbmRleCA9IGV2ZW50LmRldGFpbC5zcGxpdEluZGV4O1xuICAgIHJlY29yZC5zcGxpdHMuc3BsaWNlKHNwbGl0SW5kZXgsIDEpO1xuXG4gICAgcmVjb3JkID0gcmVjb3JkO1xuICB9XG5cbiAgZnVuY3Rpb24gcmVtb3ZlKCkge1xuICAgIGRpc3BhdGNoKFwicmVtb3ZlUmVjb3JkXCIsIGluZGV4KTtcbiAgfVxuXG4gIGZ1bmN0aW9uIGRpc2FibGVNZSgpIHtcbiAgICByZXR1cm4gdW5kZWZpbmVkO1xuICB9XG48L3NjcmlwdD5cblxuPGRpdiBjbGFzcz1cIml0ZW0tY29udGFpbmVyIHB2MlwiPlxuICA8ZGl2IGNsYXNzPVwiZmxleCBmbCB3LTEwMFwiPlxuICAgIDxkaXYgY2xhc3M9XCJwYTAgdy0xMDAgbXIyXCI+XG4gICAgICA8aW5wdXQgcGxhY2Vob2xkZXI9XCJ3aGVuXCIgYmluZDp2YWx1ZT1cIntyZWNvcmQud2hlbn1cIiB7ZGlzYWJsZWR9IC8+XG4gICAgPC9kaXY+XG5cbiAgICA8ZGl2IGNsYXNzPVwicGEwXCI+XG4gICAgICB7I2lmIGRpc2FibGVkID09PSB1bmRlZmluZWR9XG4gICAgICAgIDxidXR0b24gb246Y2xpY2s9XCJ7cmVtb3ZlfVwiIGNsYXNzPVwiaXRlbVwiPng8L2J1dHRvbj5cbiAgICAgIHs6ZWxzZX1cbiAgICAgICAgPHNwYW4+Jm5ic3A7PC9zcGFuPlxuICAgICAgey9pZn1cbiAgICA8L2Rpdj5cbiAgPC9kaXY+XG48L2Rpdj5cblxuPGRpdiBjbGFzcz1cIml0ZW0tY29udGFpbmVyIHB2MlwiPlxuICA8ZGl2IGNsYXNzPVwiZmxleCBmbGV4LWNvbHVtbiBmbCB3LTEwMFwiPlxuICAgIDxkaXYgY2xhc3M9XCJmbGV4IHB2MFwiPlxuICAgICAgPGRpdiBjbGFzcz1cInctMjUgcGExIG1yMlwiPlxuICAgICAgICA8c3Bhbj50aW1lPC9zcGFuPlxuICAgICAgPC9kaXY+XG4gICAgICA8ZGl2IGNsYXNzPVwidy0yNSBwYTEgbXIyXCI+XG4gICAgICAgIDxzcGFuPm1ldGVyczwvc3Bhbj5cbiAgICAgIDwvZGl2PlxuICAgICAgPGRpdiBjbGFzcz1cInctMjUgcGExIG1yMlwiPlxuICAgICAgICA8c3Bhbj4vNTAwbTwvc3Bhbj5cbiAgICAgIDwvZGl2PlxuXG4gICAgICA8ZGl2IGNsYXNzPVwidy0yNSBwYTEgbXIyXCI+XG4gICAgICAgIDxzcGFuPnMvbTwvc3Bhbj5cbiAgICAgIDwvZGl2PlxuICAgICAgPGRpdiBjbGFzcz1cInBhMFwiPlxuICAgICAgICA8c3BhbiBjbGFzcz1cIml0ZW0gcGExXCI+Jm5ic3A7PC9zcGFuPlxuICAgICAgPC9kaXY+XG4gICAgPC9kaXY+XG4gIDwvZGl2PlxuPC9kaXY+XG48IS0tIE9WRVJBTEw6c3RhcnQgLS0+XG48ZGl2IGNsYXNzPVwiaXRlbS1jb250YWluZXIgcHYxIG5vZHJhZ1wiPlxuICA8ZGl2IGNsYXNzPVwiZmxleCBmbGV4LWNvbHVtbiBwdjIgZmwgdy0xMDAgYncxIGJiIGJ0IGItLW1vb24tZ3JheVwiPlxuICAgIDxTcGxpdCB7ZGlzYWJsZWR9IGJpbmQ6c3BsaXQ9XCJ7cmVjb3JkLm92ZXJhbGx9XCIgLz5cbiAgPC9kaXY+XG48L2Rpdj5cbjwhLS0gT1ZFUkFMTDpmaW5pc2ggLS0+XG5cbnsjZWFjaCByZWNvcmQuc3BsaXRzIGFzIHNwbGl0LCBzcGxpdEluZGV4fVxuICA8ZGl2IGNsYXNzPVwiaXRlbS1jb250YWluZXIgcHYxXCI+XG4gICAgPGRpdiBjbGFzcz1cImZsZXggZmxleC1jb2x1bW4gZmwgdy0xMDBcIj5cbiAgICAgIDxTcGxpdCB7ZGlzYWJsZWR9IHtzcGxpdEluZGV4fSBiaW5kOnNwbGl0IG9uOmNsaWNrPVwie3JlbW92ZVNwbGl0fVwiIC8+XG4gICAgPC9kaXY+XG4gIDwvZGl2Plxuey9lYWNofVxuXG57I2lmIGRpc2FibGVkID09PSB1bmRlZmluZWR9XG4gIDxkaXYgY2xhc3M9XCJmbGV4IHB2MVwiPlxuICAgIDxidXR0b24gY2xhc3M9XCJtcjEgcGgxXCIgb246Y2xpY2s9XCJ7KCkgPT4gYWRkU3BsaXQoKX1cIj5BZGQgU3BsaXQ8L2J1dHRvbj5cbiAgPC9kaXY+XG57L2lmfVxuIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUNFLG9CQUFLLFNBQVMsQUFBQyxDQUFDLEFBQ2QsVUFBVSxDQUFFLE9BQU8sQ0FDbkIsS0FBSyxDQUFFLElBQUksQUFDYixDQUFDLEFBRUQsZUFBZSxlQUFDLENBQUMsQUFDZixPQUFPLENBQUUsSUFBSSxBQUNmLENBQUMsQUFFRCxPQUFPLGVBQUMsQ0FBQyxBQUNQLHFCQUFxQixDQUFFLElBQUksQ0FDM0IsbUJBQW1CLENBQUUsSUFBSSxDQUN6QixnQkFBZ0IsQ0FBRSxJQUFJLENBQ3RCLGVBQWUsQ0FBRSxJQUFJLENBQ3JCLFdBQVcsQ0FBRSxJQUFJLEFBQ25CLENBQUMifQ== */";
	append_dev(document.head, style);
}

function get_each_context$5(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[12] = list[i];
	child_ctx[13] = list;
	child_ctx[14] = i;
	return child_ctx;
}

// (68:6) {:else}
function create_else_block$5(ctx) {
	let span;

	const block = {
		c: function create() {
			span = element("span");
			span.textContent = " ";
			add_location(span, file$p, 68, 8, 2331);
		},
		m: function mount(target, anchor) {
			insert_dev(target, span, anchor);
		},
		p: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(span);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_else_block$5.name,
		type: "else",
		source: "(68:6) {:else}",
		ctx
	});

	return block;
}

// (66:6) {#if disabled === undefined}
function create_if_block_1$7(ctx) {
	let button;
	let dispose;

	const block = {
		c: function create() {
			button = element("button");
			button.textContent = "x";
			attr_dev(button, "class", "item");
			add_location(button, file$p, 66, 8, 2257);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, button, anchor);
			if (remount) dispose();
			dispose = listen_dev(button, "click", /*remove*/ ctx[4], false, false, false);
		},
		p: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(button);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_1$7.name,
		type: "if",
		source: "(66:6) {#if disabled === undefined}",
		ctx
	});

	return block;
}

// (105:0) {#each record.splits as split, splitIndex}
function create_each_block$5(ctx) {
	let div1;
	let div0;
	let updating_split;
	let current;

	function split_split_binding_1(value) {
		/*split_split_binding_1*/ ctx[10].call(null, value, /*split*/ ctx[12], /*each_value*/ ctx[13], /*splitIndex*/ ctx[14]);
	}

	let split_props = {
		disabled: /*disabled*/ ctx[1],
		splitIndex: /*splitIndex*/ ctx[14]
	};

	if (/*split*/ ctx[12] !== void 0) {
		split_props.split = /*split*/ ctx[12];
	}

	const split = new List_edit_data_v3_split({ props: split_props, $$inline: true });
	binding_callbacks.push(() => bind(split, "split", split_split_binding_1));
	split.$on("click", /*removeSplit*/ ctx[3]);

	const block = {
		c: function create() {
			div1 = element("div");
			div0 = element("div");
			create_component(split.$$.fragment);
			attr_dev(div0, "class", "flex flex-column fl w-100");
			add_location(div0, file$p, 106, 4, 3203);
			attr_dev(div1, "class", "item-container pv1 svelte-1dchow9");
			add_location(div1, file$p, 105, 2, 3166);
		},
		m: function mount(target, anchor) {
			insert_dev(target, div1, anchor);
			append_dev(div1, div0);
			mount_component(split, div0, null);
			current = true;
		},
		p: function update(new_ctx, dirty) {
			ctx = new_ctx;
			const split_changes = {};
			if (dirty & /*disabled*/ 2) split_changes.disabled = /*disabled*/ ctx[1];

			if (!updating_split && dirty & /*record*/ 1) {
				updating_split = true;
				split_changes.split = /*split*/ ctx[12];
				add_flush_callback(() => updating_split = false);
			}

			split.$set(split_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(split.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(split.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div1);
			destroy_component(split);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block$5.name,
		type: "each",
		source: "(105:0) {#each record.splits as split, splitIndex}",
		ctx
	});

	return block;
}

// (113:0) {#if disabled === undefined}
function create_if_block$d(ctx) {
	let div;
	let button;
	let dispose;

	const block = {
		c: function create() {
			div = element("div");
			button = element("button");
			button.textContent = "Add Split";
			attr_dev(button, "class", "mr1 ph1");
			add_location(button, file$p, 114, 4, 3406);
			attr_dev(div, "class", "flex pv1");
			add_location(div, file$p, 113, 2, 3379);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div, anchor);
			append_dev(div, button);
			if (remount) dispose();
			dispose = listen_dev(button, "click", /*click_handler*/ ctx[11], false, false, false);
		},
		p: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(div);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$d.name,
		type: "if",
		source: "(113:0) {#if disabled === undefined}",
		ctx
	});

	return block;
}

function create_fragment$q(ctx) {
	let div3;
	let div2;
	let div0;
	let input;
	let t0;
	let div1;
	let t1;
	let div11;
	let div10;
	let div9;
	let div4;
	let span0;
	let t3;
	let div5;
	let span1;
	let t5;
	let div6;
	let span2;
	let t7;
	let div7;
	let span3;
	let t9;
	let div8;
	let span4;
	let t11;
	let div13;
	let div12;
	let updating_split;
	let t12;
	let t13;
	let if_block1_anchor;
	let current;
	let dispose;

	function select_block_type(ctx, dirty) {
		if (/*disabled*/ ctx[1] === undefined) return create_if_block_1$7;
		return create_else_block$5;
	}

	let current_block_type = select_block_type(ctx);
	let if_block0 = current_block_type(ctx);

	function split_split_binding(value) {
		/*split_split_binding*/ ctx[9].call(null, value);
	}

	let split_props = { disabled: /*disabled*/ ctx[1] };

	if (/*record*/ ctx[0].overall !== void 0) {
		split_props.split = /*record*/ ctx[0].overall;
	}

	const split = new List_edit_data_v3_split({ props: split_props, $$inline: true });
	binding_callbacks.push(() => bind(split, "split", split_split_binding));
	let each_value = /*record*/ ctx[0].splits;
	validate_each_argument(each_value);
	let each_blocks = [];

	for (let i = 0; i < each_value.length; i += 1) {
		each_blocks[i] = create_each_block$5(get_each_context$5(ctx, each_value, i));
	}

	const out = i => transition_out(each_blocks[i], 1, 1, () => {
		each_blocks[i] = null;
	});

	let if_block1 = /*disabled*/ ctx[1] === undefined && create_if_block$d(ctx);

	const block = {
		c: function create() {
			div3 = element("div");
			div2 = element("div");
			div0 = element("div");
			input = element("input");
			t0 = space();
			div1 = element("div");
			if_block0.c();
			t1 = space();
			div11 = element("div");
			div10 = element("div");
			div9 = element("div");
			div4 = element("div");
			span0 = element("span");
			span0.textContent = "time";
			t3 = space();
			div5 = element("div");
			span1 = element("span");
			span1.textContent = "meters";
			t5 = space();
			div6 = element("div");
			span2 = element("span");
			span2.textContent = "/500m";
			t7 = space();
			div7 = element("div");
			span3 = element("span");
			span3.textContent = "s/m";
			t9 = space();
			div8 = element("div");
			span4 = element("span");
			span4.textContent = " ";
			t11 = space();
			div13 = element("div");
			div12 = element("div");
			create_component(split.$$.fragment);
			t12 = space();

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			t13 = space();
			if (if_block1) if_block1.c();
			if_block1_anchor = empty();
			attr_dev(input, "placeholder", "when");
			input.disabled = /*disabled*/ ctx[1];
			attr_dev(input, "class", "svelte-1dchow9");
			add_location(input, file$p, 61, 6, 2113);
			attr_dev(div0, "class", "pa0 w-100 mr2");
			add_location(div0, file$p, 60, 4, 2079);
			attr_dev(div1, "class", "pa0");
			add_location(div1, file$p, 64, 4, 2196);
			attr_dev(div2, "class", "flex fl w-100");
			add_location(div2, file$p, 59, 2, 2047);
			attr_dev(div3, "class", "item-container pv2 svelte-1dchow9");
			add_location(div3, file$p, 58, 0, 2012);
			add_location(span0, file$p, 78, 8, 2534);
			attr_dev(div4, "class", "w-25 pa1 mr2");
			add_location(div4, file$p, 77, 6, 2499);
			add_location(span1, file$p, 81, 8, 2606);
			attr_dev(div5, "class", "w-25 pa1 mr2");
			add_location(div5, file$p, 80, 6, 2571);
			add_location(span2, file$p, 84, 8, 2680);
			attr_dev(div6, "class", "w-25 pa1 mr2");
			add_location(div6, file$p, 83, 6, 2645);
			add_location(span3, file$p, 88, 8, 2754);
			attr_dev(div7, "class", "w-25 pa1 mr2");
			add_location(div7, file$p, 87, 6, 2719);
			attr_dev(span4, "class", "item pa1");
			add_location(span4, file$p, 91, 8, 2816);
			attr_dev(div8, "class", "pa0");
			add_location(div8, file$p, 90, 6, 2790);
			attr_dev(div9, "class", "flex pv0");
			add_location(div9, file$p, 76, 4, 2470);
			attr_dev(div10, "class", "flex flex-column fl w-100");
			add_location(div10, file$p, 75, 2, 2426);
			attr_dev(div11, "class", "item-container pv2 svelte-1dchow9");
			add_location(div11, file$p, 74, 0, 2391);
			attr_dev(div12, "class", "flex flex-column pv2 fl w-100 bw1 bb bt b--moon-gray");
			add_location(div12, file$p, 98, 2, 2958);
			attr_dev(div13, "class", "item-container pv1 nodrag svelte-1dchow9");
			add_location(div13, file$p, 97, 0, 2916);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div3, anchor);
			append_dev(div3, div2);
			append_dev(div2, div0);
			append_dev(div0, input);
			set_input_value(input, /*record*/ ctx[0].when);
			append_dev(div2, t0);
			append_dev(div2, div1);
			if_block0.m(div1, null);
			insert_dev(target, t1, anchor);
			insert_dev(target, div11, anchor);
			append_dev(div11, div10);
			append_dev(div10, div9);
			append_dev(div9, div4);
			append_dev(div4, span0);
			append_dev(div9, t3);
			append_dev(div9, div5);
			append_dev(div5, span1);
			append_dev(div9, t5);
			append_dev(div9, div6);
			append_dev(div6, span2);
			append_dev(div9, t7);
			append_dev(div9, div7);
			append_dev(div7, span3);
			append_dev(div9, t9);
			append_dev(div9, div8);
			append_dev(div8, span4);
			insert_dev(target, t11, anchor);
			insert_dev(target, div13, anchor);
			append_dev(div13, div12);
			mount_component(split, div12, null);
			insert_dev(target, t12, anchor);

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(target, anchor);
			}

			insert_dev(target, t13, anchor);
			if (if_block1) if_block1.m(target, anchor);
			insert_dev(target, if_block1_anchor, anchor);
			current = true;
			if (remount) dispose();
			dispose = listen_dev(input, "input", /*input_input_handler*/ ctx[8]);
		},
		p: function update(ctx, [dirty]) {
			if (!current || dirty & /*disabled*/ 2) {
				prop_dev(input, "disabled", /*disabled*/ ctx[1]);
			}

			if (dirty & /*record*/ 1 && input.value !== /*record*/ ctx[0].when) {
				set_input_value(input, /*record*/ ctx[0].when);
			}

			if (current_block_type === (current_block_type = select_block_type(ctx)) && if_block0) {
				if_block0.p(ctx, dirty);
			} else {
				if_block0.d(1);
				if_block0 = current_block_type(ctx);

				if (if_block0) {
					if_block0.c();
					if_block0.m(div1, null);
				}
			}

			const split_changes = {};
			if (dirty & /*disabled*/ 2) split_changes.disabled = /*disabled*/ ctx[1];

			if (!updating_split && dirty & /*record*/ 1) {
				updating_split = true;
				split_changes.split = /*record*/ ctx[0].overall;
				add_flush_callback(() => updating_split = false);
			}

			split.$set(split_changes);

			if (dirty & /*disabled, record, removeSplit*/ 11) {
				each_value = /*record*/ ctx[0].splits;
				validate_each_argument(each_value);
				let i;

				for (i = 0; i < each_value.length; i += 1) {
					const child_ctx = get_each_context$5(ctx, each_value, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
						transition_in(each_blocks[i], 1);
					} else {
						each_blocks[i] = create_each_block$5(child_ctx);
						each_blocks[i].c();
						transition_in(each_blocks[i], 1);
						each_blocks[i].m(t13.parentNode, t13);
					}
				}

				group_outros();

				for (i = each_value.length; i < each_blocks.length; i += 1) {
					out(i);
				}

				check_outros();
			}

			if (/*disabled*/ ctx[1] === undefined) {
				if (if_block1) {
					if_block1.p(ctx, dirty);
				} else {
					if_block1 = create_if_block$d(ctx);
					if_block1.c();
					if_block1.m(if_block1_anchor.parentNode, if_block1_anchor);
				}
			} else if (if_block1) {
				if_block1.d(1);
				if_block1 = null;
			}
		},
		i: function intro(local) {
			if (current) return;
			transition_in(split.$$.fragment, local);

			for (let i = 0; i < each_value.length; i += 1) {
				transition_in(each_blocks[i]);
			}

			current = true;
		},
		o: function outro(local) {
			transition_out(split.$$.fragment, local);
			each_blocks = each_blocks.filter(Boolean);

			for (let i = 0; i < each_blocks.length; i += 1) {
				transition_out(each_blocks[i]);
			}

			current = false;
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div3);
			if_block0.d();
			if (detaching) detach_dev(t1);
			if (detaching) detach_dev(div11);
			if (detaching) detach_dev(t11);
			if (detaching) detach_dev(div13);
			destroy_component(split);
			if (detaching) detach_dev(t12);
			destroy_each(each_blocks, detaching);
			if (detaching) detach_dev(t13);
			if (if_block1) if_block1.d(detaching);
			if (detaching) detach_dev(if_block1_anchor);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$q.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function disableMe() {
	return undefined;
}

function instance$q($$self, $$props, $$invalidate) {
	let { index } = $$props;
	let { record } = $$props;
	let { disabled = undefined } = $$props;
	const dispatch = createEventDispatcher();
	const newRow = { time: "", distance: 0, p500: "", spm: 0 };

	function addSplit() {
		record.splits.push(copyObject(newRow));
		$$invalidate(0, record);
	}

	function removeSplit(event) {
		const splitIndex = event.detail.splitIndex;
		record.splits.splice(splitIndex, 1);
		$$invalidate(0, record);
	}

	function remove() {
		dispatch("removeRecord", index);
	}

	const writable_props = ["index", "record", "disabled"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_edit_data_v3_record> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_edit_data_v3_record", $$slots, []);

	function input_input_handler() {
		record.when = this.value;
		$$invalidate(0, record);
	}

	function split_split_binding(value) {
		record.overall = value;
		$$invalidate(0, record);
	}

	function split_split_binding_1(value, split, each_value, splitIndex) {
		each_value[splitIndex] = value;
		$$invalidate(0, record);
	}

	const click_handler = () => addSplit();

	$$self.$set = $$props => {
		if ("index" in $$props) $$invalidate(5, index = $$props.index);
		if ("record" in $$props) $$invalidate(0, record = $$props.record);
		if ("disabled" in $$props) $$invalidate(1, disabled = $$props.disabled);
	};

	$$self.$capture_state = () => ({
		copyObject,
		isDeviceMobile,
		tap,
		afterUpdate,
		Split: List_edit_data_v3_split,
		index,
		record,
		disabled,
		createEventDispatcher,
		dispatch,
		newRow,
		addSplit,
		removeSplit,
		remove,
		disableMe
	});

	$$self.$inject_state = $$props => {
		if ("index" in $$props) $$invalidate(5, index = $$props.index);
		if ("record" in $$props) $$invalidate(0, record = $$props.record);
		if ("disabled" in $$props) $$invalidate(1, disabled = $$props.disabled);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [
		record,
		disabled,
		addSplit,
		removeSplit,
		remove,
		index,
		dispatch,
		newRow,
		input_input_handler,
		split_split_binding,
		split_split_binding_1,
		click_handler
	];
}

class List_edit_data_v3_record extends SvelteComponentDev {
	constructor(options) {
		super(options);
		if (!document.getElementById("svelte-1dchow9-style")) add_css$5();
		init(this, options, instance$q, create_fragment$q, safe_not_equal, { index: 5, record: 0, disabled: 1 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_edit_data_v3_record",
			options,
			id: create_fragment$q.name
		});

		const { ctx } = this.$$;
		const props = options.props || {};

		if (/*index*/ ctx[5] === undefined && !("index" in props)) {
			console.warn("<List_edit_data_v3_record> was created without expected prop 'index'");
		}

		if (/*record*/ ctx[0] === undefined && !("record" in props)) {
			console.warn("<List_edit_data_v3_record> was created without expected prop 'record'");
		}
	}

	get index() {
		throw new Error("<List_edit_data_v3_record>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set index(value) {
		throw new Error("<List_edit_data_v3_record>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	get record() {
		throw new Error("<List_edit_data_v3_record>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set record(value) {
		throw new Error("<List_edit_data_v3_record>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	get disabled() {
		throw new Error("<List_edit_data_v3_record>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set disabled(value) {
		throw new Error("<List_edit_data_v3_record>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list_edit_data_v3.svelte generated by Svelte v3.20.1 */
const file$q = "src/editor/components/list_edit_data_v3.svelte";

function add_css$6() {
	var style = element("style");
	style.id = "svelte-1fdkwjz-style";
	style.textContent = "{}\n/*# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibGlzdF9lZGl0X2RhdGFfdjMuc3ZlbHRlIiwic291cmNlcyI6WyJsaXN0X2VkaXRfZGF0YV92My5zdmVsdGUiXSwic291cmNlc0NvbnRlbnQiOlsiPHN0eWxlPlxuICAuY29udGFpbmVyIHtcbiAgICBkaXNwbGF5OiBmbGV4O1xuICAgIGp1c3RpZnktY29udGVudDogc3BhY2UtYmV0d2VlbjtcbiAgICBmbGV4LWRpcmVjdGlvbjogY29sdW1uO1xuICB9XG5cbiAgLml0ZW0tY29udGFpbmVyIHtcbiAgICBkaXNwbGF5OiBmbGV4O1xuICB9XG5cbiAgLml0ZW0tY29udGFpbmVyIC5pdGVtIHtcbiAgfVxuXG4gIC5pdGVtLWNvbnRhaW5lciAuaXRlbS1sZWZ0IHtcbiAgICBmbGV4LWdyb3c6IDE7IC8qIFNldCB0aGUgbWlkZGxlIGVsZW1lbnQgdG8gZ3JvdyBhbmQgc3RyZXRjaCAqL1xuICAgIG1hcmdpbi1yaWdodDogMC41ZW07XG4gIH1cblxuLyojIHNvdXJjZU1hcHBpbmdVUkw9ZGF0YTphcHBsaWNhdGlvbi9qc29uO2Jhc2U2NCxleUoyWlhKemFXOXVJam96TENKemIzVnlZMlZ6SWpwYkluTnlZeTlsWkdsMGIzSXZZMjl0Y0c5dVpXNTBjeTlzYVhOMFgyVmthWFJmWkdGMFlWOTJNeTV6ZG1Wc2RHVWlYU3dpYm1GdFpYTWlPbHRkTENKdFlYQndhVzVuY3lJNklqdEZRVU5GTzBsQlEwVXNZVUZCWVR0SlFVTmlMRGhDUVVFNFFqdEpRVU01UWl4elFrRkJjMEk3UlVGRGVFSTdPMFZCUlVFN1NVRkRSU3hoUVVGaE8wVkJRMlk3TzBWQlJVRTdSVUZEUVRzN1JVRkZRVHRKUVVORkxGbEJRVmtzUlVGQlJTd3JRMEZCSzBNN1NVRkROMFFzYlVKQlFXMUNPMFZCUTNKQ0lpd2labWxzWlNJNkluTnlZeTlsWkdsMGIzSXZZMjl0Y0c5dVpXNTBjeTlzYVhOMFgyVmthWFJmWkdGMFlWOTJNeTV6ZG1Wc2RHVWlMQ0p6YjNWeVkyVnpRMjl1ZEdWdWRDSTZXeUpjYmlBZ0xtTnZiblJoYVc1bGNpQjdYRzRnSUNBZ1pHbHpjR3hoZVRvZ1pteGxlRHRjYmlBZ0lDQnFkWE4wYVdaNUxXTnZiblJsYm5RNklITndZV05sTFdKbGRIZGxaVzQ3WEc0Z0lDQWdabXhsZUMxa2FYSmxZM1JwYjI0NklHTnZiSFZ0Ymp0Y2JpQWdmVnh1WEc0Z0lDNXBkR1Z0TFdOdmJuUmhhVzVsY2lCN1hHNGdJQ0FnWkdsemNHeGhlVG9nWm14bGVEdGNiaUFnZlZ4dVhHNGdJQzVwZEdWdExXTnZiblJoYVc1bGNpQXVhWFJsYlNCN1hHNGdJSDFjYmx4dUlDQXVhWFJsYlMxamIyNTBZV2x1WlhJZ0xtbDBaVzB0YkdWbWRDQjdYRzRnSUNBZ1pteGxlQzFuY205M09pQXhPeUF2S2lCVFpYUWdkR2hsSUcxcFpHUnNaU0JsYkdWdFpXNTBJSFJ2SUdkeWIzY2dZVzVrSUhOMGNtVjBZMmdnS2k5Y2JpQWdJQ0J0WVhKbmFXNHRjbWxuYUhRNklEQXVOV1Z0TzF4dUlDQjlYRzRpWFgwPSAqLzwvc3R5bGU+XG5cbjxzY3JpcHQ+XG4gIGltcG9ydCB7IGNvcHlPYmplY3QsIGlzRGV2aWNlTW9iaWxlIH0gZnJvbSBcIi4uL2xpYi9oZWxwZXIuanNcIjtcbiAgaW1wb3J0IHsgdGFwIH0gZnJvbSBcIkBzdmVsdGVqcy9nZXN0dXJlc1wiO1xuICBpbXBvcnQgeyBhZnRlclVwZGF0ZSB9IGZyb20gXCJzdmVsdGVcIjtcbiAgaW1wb3J0IFJlY29yZCBmcm9tIFwiLi9saXN0X2VkaXRfZGF0YV92M19yZWNvcmQuc3ZlbHRlXCI7XG5cbiAgZXhwb3J0IGxldCBsaXN0RGF0YTtcblxuICBjb25zdCBwb3NzaWJsZUNvbW1hbmRzID0ge1xuICAgIG5vdGhpbmc6IFwiXCIsXG4gICAgbmV3SXRlbTogXCJXaGVuIGFuIGl0ZW0gaXMgYWRkZWRcIlxuICB9O1xuXG4gIGNvbnN0IGlzTW9iaWxlID0gaXNEZXZpY2VNb2JpbGUoKTtcbiAgY29uc3Qgb3JkZXJIZWxwZXJUZXh0ID0gIWlzTW9iaWxlID8gXCJkcmFnIGFuZCBkcm9wIHRvIHN3YXBcIiA6IFwidGFwIHRvIHN3YXBcIjtcblxuICBjb25zdCBuZXdSb3cgPSB7XG4gICAgd2hlbjogXCJcIixcbiAgICBvdmVyYWxsOiB7IHRpbWU6IFwiXCIsIGRpc3RhbmNlOiAwLCBwNTAwOiBcIlwiLCBzcG06IDAgfSxcbiAgICBzcGxpdHM6IFtdXG4gIH07XG5cbiAgY29uc3QgX3N3YXBJdGVtcyA9IHtcbiAgICBmcm9tOiAtMSxcbiAgICBmcm9tRWxlbWVudDogbnVsbCxcbiAgICB0bzogLTEsXG4gICAgdG9FbGVtZW50OiBudWxsXG4gIH07XG5cbiAgbGV0IGl0ZW1zQ29udGFpbmVyO1xuICBsZXQgbGFzdENtZCA9IHBvc3NpYmxlQ29tbWFuZHMubm90aGluZztcblxuICBsZXQgZW5hYmxlU29ydGFibGUgPSBmYWxzZTtcblxuICBsZXQgc3dhcEl0ZW1zID0gY29weU9iamVjdChfc3dhcEl0ZW1zKTtcblxuICBhZnRlclVwZGF0ZSgoKSA9PiB7XG4gICAgaWYgKGxhc3RDbWQgPT09IHBvc3NpYmxlQ29tbWFuZHMubmV3SXRlbSkge1xuICAgICAgLy8gVGhpcyBvbmx5IHdvcmtzIGZvciBWMSBlbGVtZW50c1xuICAgICAgLy9sZXQgbm9kZXMgPSBpdGVtc0NvbnRhaW5lci5xdWVyeVNlbGVjdG9yQWxsKFwiLml0ZW0tY29udGFpbmVyXCIpO1xuICAgICAgLy9ub2Rlc1tub2Rlcy5sZW5ndGggLSAxXS5xdWVyeVNlbGVjdG9yKFwiaW5wdXQ6Zmlyc3QtY2hpbGRcIikuZm9jdXMoKTtcbiAgICAgIGxhc3RDbWQgPSBwb3NzaWJsZUNvbW1hbmRzLm5vdGhpbmc7XG4gICAgfVxuICB9KTtcblxuICBmdW5jdGlvbiBhZGQoKSB7XG4gICAgbGlzdERhdGEucHVzaChjb3B5T2JqZWN0KG5ld1JvdykpO1xuICAgIGxpc3REYXRhID0gbGlzdERhdGE7XG4gICAgbGFzdENtZCA9IHBvc3NpYmxlQ29tbWFuZHMubmV3SXRlbTtcbiAgfVxuXG4gIGZ1bmN0aW9uIHJlbW92ZShldmVudCkge1xuICAgIGxpc3REYXRhLnNwbGljZShldmVudC5kZXRhaWwuaW5kZXgsIDEpO1xuICAgIGlmICghbGlzdERhdGEubGVuZ3RoKSB7XG4gICAgICBsaXN0RGF0YSA9IFtjb3B5T2JqZWN0KG5ld1JvdyldO1xuICAgIH1cbiAgICBsaXN0RGF0YSA9IGxpc3REYXRhO1xuICB9XG5cbiAgZnVuY3Rpb24gcmVtb3ZlQWxsKCkge1xuICAgIGxpc3REYXRhID0gW2NvcHlPYmplY3QobmV3Um93KV07XG4gIH1cblxuICBmdW5jdGlvbiB0b2dnbGVTb3J0YWJsZShldikge1xuICAgIGlmIChsaXN0RGF0YS5sZW5ndGggPD0gMSkge1xuICAgICAgYWxlcnQoXCJub3RoaW5nIHRvIHN3YXBcIik7XG4gICAgICByZXR1cm47XG4gICAgfVxuXG4gICAgZW5hYmxlU29ydGFibGUgPSBlbmFibGVTb3J0YWJsZSA/IGZhbHNlIDogdHJ1ZTtcbiAgICBpZiAoZW5hYmxlU29ydGFibGUpIHtcbiAgICAgIC8vIFJlc2V0IHN3YXBJdGVtc1xuICAgICAgc3dhcEl0ZW1zID0gY29weU9iamVjdChfc3dhcEl0ZW1zKTtcbiAgICB9XG4gIH1cblxuICBmdW5jdGlvbiBkcmFnc3RhcnQoZXYpIHtcbiAgICBzd2FwSXRlbXMgPSBjb3B5T2JqZWN0KF9zd2FwSXRlbXMpO1xuICAgIHN3YXBJdGVtcy5mcm9tID0gZXYudGFyZ2V0LmdldEF0dHJpYnV0ZShcImRhdGEtaW5kZXhcIik7XG4gIH1cblxuICBmdW5jdGlvbiBkcmFnb3Zlcihldikge1xuICAgIGV2LnByZXZlbnREZWZhdWx0KCk7XG4gIH1cblxuICBmdW5jdGlvbiBkcm9wKGV2KSB7XG4gICAgZXYucHJldmVudERlZmF1bHQoKTtcbiAgICBzd2FwSXRlbXMudG8gPSBldi50YXJnZXQuZ2V0QXR0cmlidXRlKFwiZGF0YS1pbmRleFwiKTtcblxuICAgIC8vIFdlIG1pZ2h0IGxhbmQgb24gdGhlIGNoaWxkcmVuLCBsb29rIHVwIGZvciB0aGUgZHJhZ2dhYmxlIGF0dHJpYnV0ZVxuICAgIGlmIChzd2FwSXRlbXMudG8gPT0gbnVsbCkge1xuICAgICAgc3dhcEl0ZW1zLnRvID0gZXYudGFyZ2V0XG4gICAgICAgIC5jbG9zZXN0KFwiW2RyYWdnYWJsZV1cIilcbiAgICAgICAgLmdldEF0dHJpYnV0ZShcImRhdGEtaW5kZXhcIik7XG4gICAgfVxuXG4gICAgbGV0IGEgPSBsaXN0RGF0YVtzd2FwSXRlbXMuZnJvbV07XG4gICAgbGV0IGIgPSBsaXN0RGF0YVtzd2FwSXRlbXMudG9dO1xuICAgIGxpc3REYXRhW3N3YXBJdGVtcy5mcm9tXSA9IGI7XG4gICAgbGlzdERhdGFbc3dhcEl0ZW1zLnRvXSA9IGE7XG4gIH1cblxuICBmdW5jdGlvbiB0YXBIYW5kbGVyKGV2KSB7XG4gICAgZXYucHJldmVudERlZmF1bHQoKTtcblxuICAgIGxldCBpbmRleCA9IGV2LnRhcmdldC5nZXRBdHRyaWJ1dGUoXCJkYXRhLWluZGV4XCIpO1xuXG4gICAgaWYgKGluZGV4ID09PSBudWxsKSB7XG4gICAgICBzd2FwSXRlbXMgPSBjb3B5T2JqZWN0KF9zd2FwSXRlbXMpO1xuICAgICAgcmV0dXJuO1xuICAgIH1cblxuICAgIGlmIChzd2FwSXRlbXMuZnJvbSA9PT0gLTEpIHtcbiAgICAgIHN3YXBJdGVtcy5mcm9tRWxlbWVudCA9IGV2LnRhcmdldDtcbiAgICAgIHN3YXBJdGVtcy5mcm9tRWxlbWVudC5zdHlsZVtcImJvcmRlci1sZWZ0XCJdID0gXCJzb2xpZCBncmVlblwiO1xuICAgICAgc3dhcEl0ZW1zLmZyb20gPSBpbmRleDtcbiAgICAgIHJldHVybjtcbiAgICB9XG5cbiAgICBpZiAoc3dhcEl0ZW1zLmZyb20gPT09IGluZGV4KSB7XG4gICAgICBzd2FwSXRlbXMuZnJvbUVsZW1lbnQuc3R5bGUuYm9yZGVyID0gXCJcIjtcbiAgICAgIHN3YXBJdGVtcyA9IGNvcHlPYmplY3QoX3N3YXBJdGVtcyk7XG4gICAgICByZXR1cm47XG4gICAgfVxuXG4gICAgc3dhcEl0ZW1zLnRvID0gaW5kZXg7XG4gICAgbGV0IGEgPSBsaXN0RGF0YVtzd2FwSXRlbXMuZnJvbV07XG4gICAgbGV0IGIgPSBsaXN0RGF0YVtzd2FwSXRlbXMudG9dO1xuICAgIGxpc3REYXRhW3N3YXBJdGVtcy5mcm9tXSA9IGI7XG4gICAgbGlzdERhdGFbc3dhcEl0ZW1zLnRvXSA9IGE7XG5cbiAgICBzd2FwSXRlbXMuZnJvbUVsZW1lbnQuc3R5bGUuYm9yZGVyID0gXCJcIjtcbiAgICBzd2FwSXRlbXMuZnJvbUVsZW1lbnQuc3R5bGVbXCJib3JkZXItcmFkaXVzXCJdID0gXCIwcHhcIjtcbiAgICBzd2FwSXRlbXMgPSBjb3B5T2JqZWN0KF9zd2FwSXRlbXMpO1xuICB9XG48L3NjcmlwdD5cblxuPGgxPkl0ZW1zPC9oMT5cblxuPGRpdiBiaW5kOnRoaXM9XCJ7aXRlbXNDb250YWluZXJ9XCI+XG4gIHsjaWYgIWVuYWJsZVNvcnRhYmxlfVxuICAgIHsjZWFjaCBsaXN0RGF0YSBhcyBsaXN0SXRlbSwgaW5kZXh9XG4gICAgICA8UmVjb3JkIHtpbmRleH0gYmluZDpyZWNvcmQ9XCJ7bGlzdEl0ZW19XCIgb246cmVtb3ZlUmVjb3JkPVwie3JlbW92ZX1cIiAvPlxuICAgIHsvZWFjaH1cblxuICAgIDxkaXYgY2xhc3M9XCJmbGV4IHB2MVwiPlxuICAgICAgPGJ1dHRvbiBjbGFzcz1cIm1yMSBwaDFcIiBvbjpjbGljaz1cInthZGR9XCI+TmV3PC9idXR0b24+XG4gICAgICA8YnV0dG9uIGNsYXNzPVwibWgxIHBoMVwiIG9uOmNsaWNrPVwie3JlbW92ZUFsbH1cIj5SZW1vdmUgYWxsPC9idXR0b24+XG4gICAgICA8YnV0dG9uIGNsYXNzPVwibWgxIHBoMVwiIG9uOmNsaWNrPVwie3RvZ2dsZVNvcnRhYmxlfVwiPkNoYW5nZSBPcmRlcjwvYnV0dG9uPlxuXG4gICAgPC9kaXY+XG4gIHsvaWZ9XG5cbiAgeyNpZiBlbmFibGVTb3J0YWJsZX1cbiAgICB7I2VhY2ggbGlzdERhdGEgYXMgbGlzdEl0ZW0sIHBvc31cbiAgICAgIDxkaXZcbiAgICAgICAgZHJhZ2dhYmxlPVwidHJ1ZVwiXG4gICAgICAgIGNsYXNzPVwiZHJvcHpvbmUgcHYyIGJiIGItLWJsYWNrLTA1XCJcbiAgICAgICAgZGF0YS1pbmRleD1cIntwb3N9XCJcbiAgICAgICAgb246ZHJhZ3N0YXJ0PVwie2RyYWdzdGFydH1cIlxuICAgICAgICBvbjpkcmFnb3Zlcj1cIntkcmFnb3Zlcn1cIlxuICAgICAgICBvbjpkcm9wPVwie2Ryb3B9XCJcbiAgICAgICAgdXNlOnRhcFxuICAgICAgICBvbjp0YXA9XCJ7dGFwSGFuZGxlcn1cIlxuICAgICAgPlxuICAgICAgICA8UmVjb3JkIGRpc2FibGVkPVwidHJ1ZVwiIGluZGV4PVwie3Bvc31cIiBiaW5kOnJlY29yZD1cIntsaXN0SXRlbX1cIiAvPlxuICAgICAgPC9kaXY+XG4gICAgey9lYWNofVxuXG4gICAgPGJ1dHRvbiBvbjpjbGljaz1cInt0b2dnbGVTb3J0YWJsZX1cIj5cbiAgICAgIEZpbmlzaGVkIG9yZGVyaW5nPyAoe29yZGVySGVscGVyVGV4dH0pXG4gICAgPC9idXR0b24+XG4gIHsvaWZ9XG48L2Rpdj5cbiJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFXd0IsQ0FBQyxBQUN2QixDQUFDIn0= */";
	append_dev(document.head, style);
}

function get_each_context$6(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[20] = list[i];
	child_ctx[21] = list;
	child_ctx[22] = i;
	return child_ctx;
}

function get_each_context_1$4(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[20] = list[i];
	child_ctx[23] = list;
	child_ctx[24] = i;
	return child_ctx;
}

// (162:2) {#if !enableSortable}
function create_if_block_1$8(ctx) {
	let t0;
	let div;
	let button0;
	let t2;
	let button1;
	let t4;
	let button2;
	let current;
	let dispose;
	let each_value_1 = /*listData*/ ctx[0];
	validate_each_argument(each_value_1);
	let each_blocks = [];

	for (let i = 0; i < each_value_1.length; i += 1) {
		each_blocks[i] = create_each_block_1$4(get_each_context_1$4(ctx, each_value_1, i));
	}

	const out = i => transition_out(each_blocks[i], 1, 1, () => {
		each_blocks[i] = null;
	});

	const block = {
		c: function create() {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			t0 = space();
			div = element("div");
			button0 = element("button");
			button0.textContent = "New";
			t2 = space();
			button1 = element("button");
			button1.textContent = "Remove all";
			t4 = space();
			button2 = element("button");
			button2.textContent = "Change Order";
			attr_dev(button0, "class", "mr1 ph1");
			add_location(button0, file$q, 167, 6, 4713);
			attr_dev(button1, "class", "mh1 ph1");
			add_location(button1, file$q, 168, 6, 4773);
			attr_dev(button2, "class", "mh1 ph1");
			add_location(button2, file$q, 169, 6, 4846);
			attr_dev(div, "class", "flex pv1");
			add_location(div, file$q, 166, 4, 4684);
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
			current = true;
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(button0, "click", /*add*/ ctx[4], false, false, false),
				listen_dev(button1, "click", /*removeAll*/ ctx[6], false, false, false),
				listen_dev(button2, "click", /*toggleSortable*/ ctx[7], false, false, false)
			];
		},
		p: function update(ctx, dirty) {
			if (dirty & /*listData, remove*/ 33) {
				each_value_1 = /*listData*/ ctx[0];
				validate_each_argument(each_value_1);
				let i;

				for (i = 0; i < each_value_1.length; i += 1) {
					const child_ctx = get_each_context_1$4(ctx, each_value_1, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
						transition_in(each_blocks[i], 1);
					} else {
						each_blocks[i] = create_each_block_1$4(child_ctx);
						each_blocks[i].c();
						transition_in(each_blocks[i], 1);
						each_blocks[i].m(t0.parentNode, t0);
					}
				}

				group_outros();

				for (i = each_value_1.length; i < each_blocks.length; i += 1) {
					out(i);
				}

				check_outros();
			}
		},
		i: function intro(local) {
			if (current) return;

			for (let i = 0; i < each_value_1.length; i += 1) {
				transition_in(each_blocks[i]);
			}

			current = true;
		},
		o: function outro(local) {
			each_blocks = each_blocks.filter(Boolean);

			for (let i = 0; i < each_blocks.length; i += 1) {
				transition_out(each_blocks[i]);
			}

			current = false;
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
		id: create_if_block_1$8.name,
		type: "if",
		source: "(162:2) {#if !enableSortable}",
		ctx
	});

	return block;
}

// (163:4) {#each listData as listItem, index}
function create_each_block_1$4(ctx) {
	let updating_record;
	let current;

	function record_record_binding(value) {
		/*record_record_binding*/ ctx[17].call(null, value, /*listItem*/ ctx[20], /*each_value_1*/ ctx[23], /*index*/ ctx[24]);
	}

	let record_props = { index: /*index*/ ctx[24] };

	if (/*listItem*/ ctx[20] !== void 0) {
		record_props.record = /*listItem*/ ctx[20];
	}

	const record = new List_edit_data_v3_record({ props: record_props, $$inline: true });
	binding_callbacks.push(() => bind(record, "record", record_record_binding));
	record.$on("removeRecord", /*remove*/ ctx[5]);

	const block = {
		c: function create() {
			create_component(record.$$.fragment);
		},
		m: function mount(target, anchor) {
			mount_component(record, target, anchor);
			current = true;
		},
		p: function update(new_ctx, dirty) {
			ctx = new_ctx;
			const record_changes = {};

			if (!updating_record && dirty & /*listData*/ 1) {
				updating_record = true;
				record_changes.record = /*listItem*/ ctx[20];
				add_flush_callback(() => updating_record = false);
			}

			record.$set(record_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(record.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(record.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(record, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block_1$4.name,
		type: "each",
		source: "(163:4) {#each listData as listItem, index}",
		ctx
	});

	return block;
}

// (175:2) {#if enableSortable}
function create_if_block$e(ctx) {
	let t0;
	let button;
	let current;
	let dispose;
	let each_value = /*listData*/ ctx[0];
	validate_each_argument(each_value);
	let each_blocks = [];

	for (let i = 0; i < each_value.length; i += 1) {
		each_blocks[i] = create_each_block$6(get_each_context$6(ctx, each_value, i));
	}

	const out = i => transition_out(each_blocks[i], 1, 1, () => {
		each_blocks[i] = null;
	});

	const block = {
		c: function create() {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			t0 = space();
			button = element("button");

			button.textContent = `
      Finished ordering? (${/*orderHelperText*/ ctx[3]})
    `;

			add_location(button, file$q, 190, 4, 5360);
		},
		m: function mount(target, anchor, remount) {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(target, anchor);
			}

			insert_dev(target, t0, anchor);
			insert_dev(target, button, anchor);
			current = true;
			if (remount) dispose();
			dispose = listen_dev(button, "click", /*toggleSortable*/ ctx[7], false, false, false);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*dragstart, dragover, drop, tapHandler, listData*/ 1793) {
				each_value = /*listData*/ ctx[0];
				validate_each_argument(each_value);
				let i;

				for (i = 0; i < each_value.length; i += 1) {
					const child_ctx = get_each_context$6(ctx, each_value, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
						transition_in(each_blocks[i], 1);
					} else {
						each_blocks[i] = create_each_block$6(child_ctx);
						each_blocks[i].c();
						transition_in(each_blocks[i], 1);
						each_blocks[i].m(t0.parentNode, t0);
					}
				}

				group_outros();

				for (i = each_value.length; i < each_blocks.length; i += 1) {
					out(i);
				}

				check_outros();
			}
		},
		i: function intro(local) {
			if (current) return;

			for (let i = 0; i < each_value.length; i += 1) {
				transition_in(each_blocks[i]);
			}

			current = true;
		},
		o: function outro(local) {
			each_blocks = each_blocks.filter(Boolean);

			for (let i = 0; i < each_blocks.length; i += 1) {
				transition_out(each_blocks[i]);
			}

			current = false;
		},
		d: function destroy(detaching) {
			destroy_each(each_blocks, detaching);
			if (detaching) detach_dev(t0);
			if (detaching) detach_dev(button);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$e.name,
		type: "if",
		source: "(175:2) {#if enableSortable}",
		ctx
	});

	return block;
}

// (176:4) {#each listData as listItem, pos}
function create_each_block$6(ctx) {
	let div;
	let updating_record;
	let div_data_index_value;
	let tap_action;
	let current;
	let dispose;

	function record_record_binding_1(value) {
		/*record_record_binding_1*/ ctx[18].call(null, value, /*listItem*/ ctx[20], /*each_value*/ ctx[21], /*pos*/ ctx[22]);
	}

	let record_props = { disabled: "true", index: /*pos*/ ctx[22] };

	if (/*listItem*/ ctx[20] !== void 0) {
		record_props.record = /*listItem*/ ctx[20];
	}

	const record = new List_edit_data_v3_record({ props: record_props, $$inline: true });
	binding_callbacks.push(() => bind(record, "record", record_record_binding_1));

	const block = {
		c: function create() {
			div = element("div");
			create_component(record.$$.fragment);
			attr_dev(div, "draggable", "true");
			attr_dev(div, "class", "dropzone pv2 bb b--black-05");
			attr_dev(div, "data-index", div_data_index_value = /*pos*/ ctx[22]);
			add_location(div, file$q, 176, 6, 5008);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div, anchor);
			mount_component(record, div, null);
			current = true;
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(div, "dragstart", /*dragstart*/ ctx[8], false, false, false),
				listen_dev(div, "dragover", dragover$2, false, false, false),
				listen_dev(div, "drop", /*drop*/ ctx[9], false, false, false),
				action_destroyer(tap_action = tap.call(null, div)),
				listen_dev(div, "tap", /*tapHandler*/ ctx[10], false, false, false)
			];
		},
		p: function update(new_ctx, dirty) {
			ctx = new_ctx;
			const record_changes = {};

			if (!updating_record && dirty & /*listData*/ 1) {
				updating_record = true;
				record_changes.record = /*listItem*/ ctx[20];
				add_flush_callback(() => updating_record = false);
			}

			record.$set(record_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(record.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(record.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div);
			destroy_component(record);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block$6.name,
		type: "each",
		source: "(176:4) {#each listData as listItem, pos}",
		ctx
	});

	return block;
}

function create_fragment$r(ctx) {
	let h1;
	let t1;
	let div;
	let t2;
	let current;
	let if_block0 = !/*enableSortable*/ ctx[2] && create_if_block_1$8(ctx);
	let if_block1 = /*enableSortable*/ ctx[2] && create_if_block$e(ctx);

	const block = {
		c: function create() {
			h1 = element("h1");
			h1.textContent = "Items";
			t1 = space();
			div = element("div");
			if (if_block0) if_block0.c();
			t2 = space();
			if (if_block1) if_block1.c();
			add_location(h1, file$q, 158, 0, 4475);
			add_location(div, file$q, 160, 0, 4491);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, h1, anchor);
			insert_dev(target, t1, anchor);
			insert_dev(target, div, anchor);
			if (if_block0) if_block0.m(div, null);
			append_dev(div, t2);
			if (if_block1) if_block1.m(div, null);
			/*div_binding*/ ctx[19](div);
			current = true;
		},
		p: function update(ctx, [dirty]) {
			if (!/*enableSortable*/ ctx[2]) {
				if (if_block0) {
					if_block0.p(ctx, dirty);
					transition_in(if_block0, 1);
				} else {
					if_block0 = create_if_block_1$8(ctx);
					if_block0.c();
					transition_in(if_block0, 1);
					if_block0.m(div, t2);
				}
			} else if (if_block0) {
				group_outros();

				transition_out(if_block0, 1, 1, () => {
					if_block0 = null;
				});

				check_outros();
			}

			if (/*enableSortable*/ ctx[2]) {
				if (if_block1) {
					if_block1.p(ctx, dirty);
					transition_in(if_block1, 1);
				} else {
					if_block1 = create_if_block$e(ctx);
					if_block1.c();
					transition_in(if_block1, 1);
					if_block1.m(div, null);
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
			if (detaching) detach_dev(h1);
			if (detaching) detach_dev(t1);
			if (detaching) detach_dev(div);
			if (if_block0) if_block0.d();
			if (if_block1) if_block1.d();
			/*div_binding*/ ctx[19](null);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$r.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function dragover$2(ev) {
	ev.preventDefault();
}

function instance$r($$self, $$props, $$invalidate) {
	let { listData } = $$props;

	const possibleCommands = {
		nothing: "",
		newItem: "When an item is added"
	};

	const isMobile = isDeviceMobile();
	const orderHelperText = !isMobile ? "drag and drop to swap" : "tap to swap";

	const newRow = {
		when: "",
		overall: { time: "", distance: 0, p500: "", spm: 0 },
		splits: []
	};

	const _swapItems = {
		from: -1,
		fromElement: null,
		to: -1,
		toElement: null
	};

	let itemsContainer;
	let lastCmd = possibleCommands.nothing;
	let enableSortable = false;
	let swapItems = copyObject(_swapItems);

	afterUpdate(() => {
		if (lastCmd === possibleCommands.newItem) {
			// This only works for V1 elements
			//let nodes = itemsContainer.querySelectorAll(".item-container");
			//nodes[nodes.length - 1].querySelector("input:first-child").focus();
			lastCmd = possibleCommands.nothing;
		}
	});

	function add() {
		listData.push(copyObject(newRow));
		$$invalidate(0, listData);
		lastCmd = possibleCommands.newItem;
	}

	function remove(event) {
		listData.splice(event.detail.index, 1);

		if (!listData.length) {
			$$invalidate(0, listData = [copyObject(newRow)]);
		}

		$$invalidate(0, listData);
	}

	function removeAll() {
		$$invalidate(0, listData = [copyObject(newRow)]);
	}

	function toggleSortable(ev) {
		if (listData.length <= 1) {
			alert("nothing to swap");
			return;
		}

		$$invalidate(2, enableSortable = enableSortable ? false : true);

		if (enableSortable) {
			// Reset swapItems
			swapItems = copyObject(_swapItems);
		}
	}

	function dragstart(ev) {
		swapItems = copyObject(_swapItems);
		swapItems.from = ev.target.getAttribute("data-index");
	}

	function drop(ev) {
		ev.preventDefault();
		swapItems.to = ev.target.getAttribute("data-index");

		// We might land on the children, look up for the draggable attribute
		if (swapItems.to == null) {
			swapItems.to = ev.target.closest("[draggable]").getAttribute("data-index");
		}

		let a = listData[swapItems.from];
		let b = listData[swapItems.to];
		$$invalidate(0, listData[swapItems.from] = b, listData);
		$$invalidate(0, listData[swapItems.to] = a, listData);
	}

	function tapHandler(ev) {
		ev.preventDefault();
		let index = ev.target.getAttribute("data-index");

		if (index === null) {
			swapItems = copyObject(_swapItems);
			return;
		}

		if (swapItems.from === -1) {
			swapItems.fromElement = ev.target;
			swapItems.fromElement.style["border-left"] = "solid green";
			swapItems.from = index;
			return;
		}

		if (swapItems.from === index) {
			swapItems.fromElement.style.border = "";
			swapItems = copyObject(_swapItems);
			return;
		}

		swapItems.to = index;
		let a = listData[swapItems.from];
		let b = listData[swapItems.to];
		$$invalidate(0, listData[swapItems.from] = b, listData);
		$$invalidate(0, listData[swapItems.to] = a, listData);
		swapItems.fromElement.style.border = "";
		swapItems.fromElement.style["border-radius"] = "0px";
		swapItems = copyObject(_swapItems);
	}

	const writable_props = ["listData"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_edit_data_v3> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_edit_data_v3", $$slots, []);

	function record_record_binding(value, listItem, each_value_1, index) {
		each_value_1[index] = value;
		$$invalidate(0, listData);
	}

	function record_record_binding_1(value, listItem, each_value, pos) {
		each_value[pos] = value;
		$$invalidate(0, listData);
	}

	function div_binding($$value) {
		binding_callbacks[$$value ? "unshift" : "push"](() => {
			$$invalidate(1, itemsContainer = $$value);
		});
	}

	$$self.$set = $$props => {
		if ("listData" in $$props) $$invalidate(0, listData = $$props.listData);
	};

	$$self.$capture_state = () => ({
		copyObject,
		isDeviceMobile,
		tap,
		afterUpdate,
		Record: List_edit_data_v3_record,
		listData,
		possibleCommands,
		isMobile,
		orderHelperText,
		newRow,
		_swapItems,
		itemsContainer,
		lastCmd,
		enableSortable,
		swapItems,
		add,
		remove,
		removeAll,
		toggleSortable,
		dragstart,
		dragover: dragover$2,
		drop,
		tapHandler
	});

	$$self.$inject_state = $$props => {
		if ("listData" in $$props) $$invalidate(0, listData = $$props.listData);
		if ("itemsContainer" in $$props) $$invalidate(1, itemsContainer = $$props.itemsContainer);
		if ("lastCmd" in $$props) lastCmd = $$props.lastCmd;
		if ("enableSortable" in $$props) $$invalidate(2, enableSortable = $$props.enableSortable);
		if ("swapItems" in $$props) swapItems = $$props.swapItems;
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [
		listData,
		itemsContainer,
		enableSortable,
		orderHelperText,
		add,
		remove,
		removeAll,
		toggleSortable,
		dragstart,
		drop,
		tapHandler,
		lastCmd,
		swapItems,
		possibleCommands,
		isMobile,
		newRow,
		_swapItems,
		record_record_binding,
		record_record_binding_1,
		div_binding
	];
}

class List_edit_data_v3 extends SvelteComponentDev {
	constructor(options) {
		super(options);
		if (!document.getElementById("svelte-1fdkwjz-style")) add_css$6();
		init(this, options, instance$r, create_fragment$r, safe_not_equal, { listData: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_edit_data_v3",
			options,
			id: create_fragment$r.name
		});

		const { ctx } = this.$$;
		const props = options.props || {};

		if (/*listData*/ ctx[0] === undefined && !("listData" in props)) {
			console.warn("<List_edit_data_v3> was created without expected prop 'listData'");
		}
	}

	get listData() {
		throw new Error("<List_edit_data_v3>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set listData(value) {
		throw new Error("<List_edit_data_v3>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list_edit_data_v4.svelte generated by Svelte v3.20.1 */

const { console: console_1$3 } = globals;
const file$r = "src/editor/components/list_edit_data_v4.svelte";

function add_css$7() {
	var style = element("style");
	style.id = "svelte-o2y3gy-style";
	style.textContent = "input.svelte-o2y3gy.svelte-o2y3gy:disabled{background:#ffcccc;color:#333}textarea.svelte-o2y3gy.svelte-o2y3gy:disabled{background:#ffcccc;color:#333}.item-container.svelte-o2y3gy.svelte-o2y3gy{display:flex}.item-container.svelte-o2y3gy .item.svelte-o2y3gy{}.item-container.svelte-o2y3gy .item-left.svelte-o2y3gy{flex-grow:1;margin-right:0.5em}\n/*# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibGlzdF9lZGl0X2RhdGFfdjQuc3ZlbHRlIiwic291cmNlcyI6WyJsaXN0X2VkaXRfZGF0YV92NC5zdmVsdGUiXSwic291cmNlc0NvbnRlbnQiOlsiPHN0eWxlPlxuICBpbnB1dDpkaXNhYmxlZCB7XG4gICAgYmFja2dyb3VuZDogI2ZmY2NjYztcbiAgICBjb2xvcjogIzMzMztcbiAgfVxuICB0ZXh0YXJlYTpkaXNhYmxlZCB7XG4gICAgYmFja2dyb3VuZDogI2ZmY2NjYztcbiAgICBjb2xvcjogIzMzMztcbiAgfVxuICAuY29udGFpbmVyIHtcbiAgICBkaXNwbGF5OiBmbGV4O1xuICAgIGp1c3RpZnktY29udGVudDogc3BhY2UtYmV0d2VlbjtcbiAgICBmbGV4LWRpcmVjdGlvbjogY29sdW1uO1xuICB9XG4gIC5pdGVtLWNvbnRhaW5lciB7XG4gICAgZGlzcGxheTogZmxleDtcbiAgfVxuICAuaXRlbS1jb250YWluZXIgLml0ZW0ge1xuICB9XG4gIC5pdGVtLWNvbnRhaW5lciAuaXRlbS1sZWZ0IHtcbiAgICBmbGV4LWdyb3c6IDE7IC8qIFNldCB0aGUgbWlkZGxlIGVsZW1lbnQgdG8gZ3JvdyBhbmQgc3RyZXRjaCAqL1xuICAgIG1hcmdpbi1yaWdodDogMC41ZW07XG4gIH1cblxuLyojIHNvdXJjZU1hcHBpbmdVUkw9ZGF0YTphcHBsaWNhdGlvbi9qc29uO2Jhc2U2NCxleUoyWlhKemFXOXVJam96TENKemIzVnlZMlZ6SWpwYkluTnlZeTlsWkdsMGIzSXZZMjl0Y0c5dVpXNTBjeTlzYVhOMFgyVmthWFJmWkdGMFlWOTJOQzV6ZG1Wc2RHVWlYU3dpYm1GdFpYTWlPbHRkTENKdFlYQndhVzVuY3lJNklqdEZRVU5GTzBsQlEwVXNiVUpCUVcxQ08wbEJRMjVDTEZkQlFWYzdSVUZEWWp0RlFVTkJPMGxCUTBVc2JVSkJRVzFDTzBsQlEyNUNMRmRCUVZjN1JVRkRZanRGUVVWQk8wbEJRMFVzWVVGQllUdEpRVU5pTERoQ1FVRTRRanRKUVVNNVFpeHpRa0ZCYzBJN1JVRkRlRUk3UlVGRlFUdEpRVU5GTEdGQlFXRTdSVUZEWmp0RlFVVkJPMFZCUTBFN1JVRkZRVHRKUVVORkxGbEJRVmtzUlVGQlJTd3JRMEZCSzBNN1NVRkROMFFzYlVKQlFXMUNPMFZCUTNKQ0lpd2labWxzWlNJNkluTnlZeTlsWkdsMGIzSXZZMjl0Y0c5dVpXNTBjeTlzYVhOMFgyVmthWFJmWkdGMFlWOTJOQzV6ZG1Wc2RHVWlMQ0p6YjNWeVkyVnpRMjl1ZEdWdWRDSTZXeUpjYmlBZ2FXNXdkWFE2WkdsellXSnNaV1FnZTF4dUlDQWdJR0poWTJ0bmNtOTFibVE2SUNObVptTmpZMk03WEc0Z0lDQWdZMjlzYjNJNklDTXpNek03WEc0Z0lIMWNiaUFnZEdWNGRHRnlaV0U2WkdsellXSnNaV1FnZTF4dUlDQWdJR0poWTJ0bmNtOTFibVE2SUNObVptTmpZMk03WEc0Z0lDQWdZMjlzYjNJNklDTXpNek03WEc0Z0lIMWNibHh1SUNBdVkyOXVkR0ZwYm1WeUlIdGNiaUFnSUNCa2FYTndiR0Y1T2lCbWJHVjRPMXh1SUNBZ0lHcDFjM1JwWm5rdFkyOXVkR1Z1ZERvZ2MzQmhZMlV0WW1WMGQyVmxianRjYmlBZ0lDQm1iR1Y0TFdScGNtVmpkR2x2YmpvZ1kyOXNkVzF1TzF4dUlDQjlYRzVjYmlBZ0xtbDBaVzB0WTI5dWRHRnBibVZ5SUh0Y2JpQWdJQ0JrYVhOd2JHRjVPaUJtYkdWNE8xeHVJQ0I5WEc1Y2JpQWdMbWwwWlcwdFkyOXVkR0ZwYm1WeUlDNXBkR1Z0SUh0Y2JpQWdmVnh1WEc0Z0lDNXBkR1Z0TFdOdmJuUmhhVzVsY2lBdWFYUmxiUzFzWldaMElIdGNiaUFnSUNCbWJHVjRMV2R5YjNjNklERTdJQzhxSUZObGRDQjBhR1VnYldsa1pHeGxJR1ZzWlcxbGJuUWdkRzhnWjNKdmR5QmhibVFnYzNSeVpYUmphQ0FxTDF4dUlDQWdJRzFoY21kcGJpMXlhV2RvZERvZ01DNDFaVzA3WEc0Z0lIMWNiaUpkZlE9PSAqLzwvc3R5bGU+XG5cbjxzY3JpcHQ+XG4gIGltcG9ydCB7IGNvcHlPYmplY3QsIGlzRGV2aWNlTW9iaWxlIH0gZnJvbSBcIi4uL2xpYi9oZWxwZXIuanNcIjtcbiAgaW1wb3J0IHsgdGFwIH0gZnJvbSBcIkBzdmVsdGVqcy9nZXN0dXJlc1wiO1xuICBpbXBvcnQgeyBhZnRlclVwZGF0ZSB9IGZyb20gXCJzdmVsdGVcIjtcblxuICBleHBvcnQgbGV0IGxpc3REYXRhO1xuXG4gIGNvbnNvbGUubG9nKFwibGlzdERhdGFcIiwgbGlzdERhdGEpO1xuXG4gIGNvbnN0IHBvc3NpYmxlQ29tbWFuZHMgPSB7XG4gICAgbm90aGluZzogXCJcIixcbiAgICBuZXdJdGVtOiBcIldoZW4gYW4gaXRlbSBpcyBhZGRlZFwiXG4gIH07XG5cbiAgY29uc3QgaXNNb2JpbGUgPSBpc0RldmljZU1vYmlsZSgpO1xuICBjb25zdCBvcmRlckhlbHBlclRleHQgPSAhaXNNb2JpbGUgPyBcImRyYWcgYW5kIGRyb3AgdG8gc3dhcFwiIDogXCJ0YXAgdG8gc3dhcFwiO1xuXG4gIGNvbnN0IG5ld1JvdyA9IHtcbiAgICBjb250ZW50OiBcIlwiLFxuICAgIHVybDogXCJcIlxuICB9O1xuICBjb25zdCBfc3dhcEl0ZW1zID0ge1xuICAgIGZyb206IC0xLFxuICAgIGZyb21FbGVtZW50OiBudWxsLFxuICAgIHRvOiAtMSxcbiAgICB0b0VsZW1lbnQ6IG51bGxcbiAgfTtcblxuICBsZXQgaXRlbXNDb250YWluZXI7XG4gIGxldCBsYXN0Q21kID0gcG9zc2libGVDb21tYW5kcy5ub3RoaW5nO1xuXG4gIGxldCBlbmFibGVTb3J0YWJsZSA9IGZhbHNlO1xuXG4gIGxldCBzd2FwSXRlbXMgPSBjb3B5T2JqZWN0KF9zd2FwSXRlbXMpO1xuXG4gIGFmdGVyVXBkYXRlKCgpID0+IHtcbiAgICBpZiAobGFzdENtZCA9PT0gcG9zc2libGVDb21tYW5kcy5uZXdJdGVtKSB7XG4gICAgICAvLyBUaGlzIG9ubHkgd29ya3MgZm9yIFYxIGVsZW1lbnRzXG4gICAgICBsZXQgbm9kZXMgPSBpdGVtc0NvbnRhaW5lci5xdWVyeVNlbGVjdG9yQWxsKFwiLml0ZW0tY29udGFpbmVyXCIpO1xuICAgICAgbm9kZXNbbm9kZXMubGVuZ3RoIC0gMV0ucXVlcnlTZWxlY3RvcihcInRleHRhcmVhOmZpcnN0LWNoaWxkXCIpLmZvY3VzKCk7XG4gICAgICBsYXN0Q21kID0gcG9zc2libGVDb21tYW5kcy5ub3RoaW5nO1xuICAgIH1cbiAgfSk7XG5cbiAgZnVuY3Rpb24gYWRkKCkge1xuICAgIGxpc3REYXRhID0gbGlzdERhdGEuY29uY2F0KGNvcHlPYmplY3QobmV3Um93KSk7XG4gICAgbGFzdENtZCA9IHBvc3NpYmxlQ29tbWFuZHMubmV3SXRlbTtcbiAgfVxuXG4gIGZ1bmN0aW9uIHJlbW92ZShsaXN0SXRlbSkge1xuICAgIGxpc3REYXRhID0gbGlzdERhdGEuZmlsdGVyKHQgPT4gdCAhPT0gbGlzdEl0ZW0pO1xuICAgIGlmICghbGlzdERhdGEubGVuZ3RoKSB7XG4gICAgICBsaXN0RGF0YSA9IFtjb3B5T2JqZWN0KG5ld1JvdyldO1xuICAgIH1cbiAgfVxuXG4gIGZ1bmN0aW9uIHJlbW92ZUFsbCgpIHtcbiAgICBsaXN0RGF0YSA9IFtjb3B5T2JqZWN0KG5ld1JvdyldO1xuICB9XG5cbiAgZnVuY3Rpb24gdG9nZ2xlU29ydGFibGUoZXYpIHtcbiAgICBpZiAobGlzdERhdGEubGVuZ3RoIDw9IDEpIHtcbiAgICAgIGFsZXJ0KFwibm90aGluZyB0byBzd2FwXCIpO1xuICAgICAgcmV0dXJuO1xuICAgIH1cblxuICAgIGVuYWJsZVNvcnRhYmxlID0gZW5hYmxlU29ydGFibGUgPyBmYWxzZSA6IHRydWU7XG4gICAgaWYgKGVuYWJsZVNvcnRhYmxlKSB7XG4gICAgICAvLyBSZXNldCBzd2FwSXRlbXNcbiAgICAgIHN3YXBJdGVtcyA9IGNvcHlPYmplY3QoX3N3YXBJdGVtcyk7XG4gICAgfVxuICB9XG5cbiAgZnVuY3Rpb24gZHJhZ3N0YXJ0KGV2KSB7XG4gICAgc3dhcEl0ZW1zID0gY29weU9iamVjdChfc3dhcEl0ZW1zKTtcbiAgICBzd2FwSXRlbXMuZnJvbSA9IGV2LnRhcmdldC5nZXRBdHRyaWJ1dGUoXCJkYXRhLWluZGV4XCIpO1xuICB9XG5cbiAgZnVuY3Rpb24gZHJhZ292ZXIoZXYpIHtcbiAgICBldi5wcmV2ZW50RGVmYXVsdCgpO1xuICB9XG5cbiAgZnVuY3Rpb24gZHJvcChldikge1xuICAgIGV2LnByZXZlbnREZWZhdWx0KCk7XG4gICAgc3dhcEl0ZW1zLnRvID0gZXYudGFyZ2V0LmdldEF0dHJpYnV0ZShcImRhdGEtaW5kZXhcIik7XG5cbiAgICAvLyBXZSBtaWdodCBsYW5kIG9uIHRoZSBjaGlsZHJlbiwgbG9vayB1cCBmb3IgdGhlIGRyYWdnYWJsZSBhdHRyaWJ1dGVcbiAgICBpZiAoc3dhcEl0ZW1zLnRvID09IG51bGwpIHtcbiAgICAgIHN3YXBJdGVtcy50byA9IGV2LnRhcmdldFxuICAgICAgICAuY2xvc2VzdChcIltkcmFnZ2FibGVdXCIpXG4gICAgICAgIC5nZXRBdHRyaWJ1dGUoXCJkYXRhLWluZGV4XCIpO1xuICAgIH1cblxuICAgIGxldCBhID0gbGlzdERhdGFbc3dhcEl0ZW1zLmZyb21dO1xuICAgIGxldCBiID0gbGlzdERhdGFbc3dhcEl0ZW1zLnRvXTtcbiAgICBsaXN0RGF0YVtzd2FwSXRlbXMuZnJvbV0gPSBiO1xuICAgIGxpc3REYXRhW3N3YXBJdGVtcy50b10gPSBhO1xuICB9XG5cbiAgZnVuY3Rpb24gdGFwSGFuZGxlcihldikge1xuICAgIGV2LnByZXZlbnREZWZhdWx0KCk7XG5cbiAgICBsZXQgaW5kZXggPSBldi50YXJnZXQuZ2V0QXR0cmlidXRlKFwiZGF0YS1pbmRleFwiKTtcblxuICAgIGlmIChpbmRleCA9PT0gbnVsbCkge1xuICAgICAgc3dhcEl0ZW1zID0gY29weU9iamVjdChfc3dhcEl0ZW1zKTtcbiAgICAgIHJldHVybjtcbiAgICB9XG5cbiAgICBpZiAoc3dhcEl0ZW1zLmZyb20gPT09IC0xKSB7XG4gICAgICBzd2FwSXRlbXMuZnJvbUVsZW1lbnQgPSBldi50YXJnZXQ7XG4gICAgICBzd2FwSXRlbXMuZnJvbUVsZW1lbnQuc3R5bGVbXCJib3JkZXItbGVmdFwiXSA9IFwic29saWQgZ3JlZW5cIjtcbiAgICAgIHN3YXBJdGVtcy5mcm9tID0gaW5kZXg7XG4gICAgICByZXR1cm47XG4gICAgfVxuXG4gICAgaWYgKHN3YXBJdGVtcy5mcm9tID09PSBpbmRleCkge1xuICAgICAgc3dhcEl0ZW1zLmZyb21FbGVtZW50LnN0eWxlLmJvcmRlciA9IFwiXCI7XG4gICAgICBzd2FwSXRlbXMgPSBjb3B5T2JqZWN0KF9zd2FwSXRlbXMpO1xuICAgICAgcmV0dXJuO1xuICAgIH1cblxuICAgIHN3YXBJdGVtcy50byA9IGluZGV4O1xuICAgIGxldCBhID0gbGlzdERhdGFbc3dhcEl0ZW1zLmZyb21dO1xuICAgIGxldCBiID0gbGlzdERhdGFbc3dhcEl0ZW1zLnRvXTtcbiAgICBsaXN0RGF0YVtzd2FwSXRlbXMuZnJvbV0gPSBiO1xuICAgIGxpc3REYXRhW3N3YXBJdGVtcy50b10gPSBhO1xuXG4gICAgc3dhcEl0ZW1zLmZyb21FbGVtZW50LnN0eWxlLmJvcmRlciA9IFwiXCI7XG4gICAgc3dhcEl0ZW1zLmZyb21FbGVtZW50LnN0eWxlW1wiYm9yZGVyLXJhZGl1c1wiXSA9IFwiMHB4XCI7XG4gICAgc3dhcEl0ZW1zID0gY29weU9iamVjdChfc3dhcEl0ZW1zKTtcbiAgfVxuPC9zY3JpcHQ+XG5cbjxoMT5JdGVtczwvaDE+XG5cbjxkaXYgYmluZDp0aGlzPVwie2l0ZW1zQ29udGFpbmVyfVwiPlxuICB7I2lmICFlbmFibGVTb3J0YWJsZX1cbiAgICB7I2VhY2ggbGlzdERhdGEgYXMgbGlzdEl0ZW19XG4gICAgICA8ZGl2IGNsYXNzPVwiaXRlbS1jb250YWluZXIgcHYyIGJiIGItLWJsYWNrLTA1XCI+XG4gICAgICAgIDxkaXYgY2xhc3M9XCJmbGV4IGZsZXgtY29sdW1uIGl0ZW0tbGVmdFwiPlxuICAgICAgICAgIDx0ZXh0YXJlYVxuICAgICAgICAgICAgcGxhY2Vob2xkZXI9XCJjb250ZW50XCJcbiAgICAgICAgICAgIGJpbmQ6dmFsdWU9XCJ7bGlzdEl0ZW0uY29udGVudH1cIlxuICAgICAgICAgICAgY2xhc3M9XCJpdGVtIGl0ZW0tbGVmdFwiXG4gICAgICAgICAgPjwvdGV4dGFyZWE+XG4gICAgICAgICAgPGlucHV0XG4gICAgICAgICAgICBwbGFjZWhvbGRlcj1cInVybFwiXG4gICAgICAgICAgICBiaW5kOnZhbHVlPVwie2xpc3RJdGVtLnVybH1cIlxuICAgICAgICAgICAgY2xhc3M9XCJpdGVtIGl0ZW0tbGVmdCBtdjJcIlxuICAgICAgICAgIC8+XG4gICAgICAgIDwvZGl2PlxuICAgICAgICA8ZGl2IGNsYXNzPVwiZmxleCBmbGV4LWNvbHVtblwiPlxuICAgICAgICAgIDxidXR0b24gb246Y2xpY2s9XCJ7KCkgPT4gcmVtb3ZlKGxpc3RJdGVtKX1cIiBjbGFzcz1cIml0ZW1cIj54PC9idXR0b24+XG4gICAgICAgIDwvZGl2PlxuICAgICAgPC9kaXY+XG4gICAgey9lYWNofVxuXG4gICAgPGRpdiBjbGFzcz1cImZsZXggcHYxXCI+XG4gICAgICA8YnV0dG9uIGNsYXNzPVwibXIxIHBoMVwiIG9uOmNsaWNrPVwie2FkZH1cIj5OZXc8L2J1dHRvbj5cblxuICAgICAgPGJ1dHRvbiBjbGFzcz1cIm1oMSBwaDFcIiBvbjpjbGljaz1cIntyZW1vdmVBbGx9XCI+UmVtb3ZlIGFsbDwvYnV0dG9uPlxuXG4gICAgICA8YnV0dG9uIGNsYXNzPVwibWgxIHBoMVwiIG9uOmNsaWNrPVwie3RvZ2dsZVNvcnRhYmxlfVwiPkNoYW5nZSBPcmRlcjwvYnV0dG9uPlxuICAgIDwvZGl2PlxuICB7L2lmfVxuXG4gIHsjaWYgZW5hYmxlU29ydGFibGV9XG4gICAgeyNlYWNoIGxpc3REYXRhIGFzIGxpc3RJdGVtLCBwb3N9XG4gICAgICA8ZGl2XG4gICAgICAgIGRyYWdnYWJsZT1cInRydWVcIlxuICAgICAgICBjbGFzcz1cImRyb3B6b25lIGl0ZW0tY29udGFpbmVyIHB2MiBiYiBiLS1ibGFjay0wNVwiXG4gICAgICAgIGRhdGEtaW5kZXg9XCJ7cG9zfVwiXG4gICAgICAgIG9uOmRyYWdzdGFydD1cIntkcmFnc3RhcnR9XCJcbiAgICAgICAgb246ZHJhZ292ZXI9XCJ7ZHJhZ292ZXJ9XCJcbiAgICAgICAgb246ZHJvcD1cIntkcm9wfVwiXG4gICAgICAgIHVzZTp0YXBcbiAgICAgICAgb246dGFwPVwie3RhcEhhbmRsZXJ9XCJcbiAgICAgID5cbiAgICAgICAgPGRpdiBjbGFzcz1cImZsZXggZmxleC1jb2x1bW4gaXRlbS1sZWZ0XCI+XG4gICAgICAgICAgPHRleHRhcmVhXG4gICAgICAgICAgICBwbGFjZWhvbGRlcj1cImNvbnRlbnRcIlxuICAgICAgICAgICAgYmluZDp2YWx1ZT1cIntsaXN0SXRlbS5jb250ZW50fVwiXG4gICAgICAgICAgICBjbGFzcz1cIml0ZW0gaXRlbS1sZWZ0XCJcbiAgICAgICAgICAgIGRpc2FibGVkXG4gICAgICAgICAgPjwvdGV4dGFyZWE+XG4gICAgICAgICAgPGlucHV0XG4gICAgICAgICAgICBwbGFjZWhvbGRlcj1cInVybFwiXG4gICAgICAgICAgICBiaW5kOnZhbHVlPVwie2xpc3RJdGVtLnVybH1cIlxuICAgICAgICAgICAgY2xhc3M9XCJpdGVtIGl0ZW0tbGVmdCBtdjJcIlxuICAgICAgICAgICAgZGlzYWJsZWRcbiAgICAgICAgICAvPlxuICAgICAgICA8L2Rpdj5cbiAgICAgIDwvZGl2PlxuICAgIHsvZWFjaH1cblxuICAgIDxidXR0b24gb246Y2xpY2s9XCJ7dG9nZ2xlU29ydGFibGV9XCI+XG4gICAgICBGaW5pc2hlZCBvcmRlcmluZz8gKHtvcmRlckhlbHBlclRleHR9KVxuICAgIDwvYnV0dG9uPlxuICB7L2lmfVxuPC9kaXY+XG4iXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQ0UsaUNBQUssU0FBUyxBQUFDLENBQUMsQUFDZCxVQUFVLENBQUUsT0FBTyxDQUNuQixLQUFLLENBQUUsSUFBSSxBQUNiLENBQUMsQUFDRCxvQ0FBUSxTQUFTLEFBQUMsQ0FBQyxBQUNqQixVQUFVLENBQUUsT0FBTyxDQUNuQixLQUFLLENBQUUsSUFBSSxBQUNiLENBQUMsQUFNRCxlQUFlLDRCQUFDLENBQUMsQUFDZixPQUFPLENBQUUsSUFBSSxBQUNmLENBQUMsQUFDRCw2QkFBZSxDQUFDLEtBQUssY0FBQyxDQUFDLEFBQ3ZCLENBQUMsQUFDRCw2QkFBZSxDQUFDLFVBQVUsY0FBQyxDQUFDLEFBQzFCLFNBQVMsQ0FBRSxDQUFDLENBQ1osWUFBWSxDQUFFLEtBQUssQUFDckIsQ0FBQyJ9 */";
	append_dev(document.head, style);
}

function get_each_context$7(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[23] = list[i];
	child_ctx[24] = list;
	child_ctx[25] = i;
	return child_ctx;
}

function get_each_context_1$5(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[23] = list[i];
	child_ctx[26] = list;
	child_ctx[27] = i;
	return child_ctx;
}

// (164:2) {#if !enableSortable}
function create_if_block_1$9(ctx) {
	let t0;
	let div;
	let button0;
	let t2;
	let button1;
	let t4;
	let button2;
	let dispose;
	let each_value_1 = /*listData*/ ctx[0];
	validate_each_argument(each_value_1);
	let each_blocks = [];

	for (let i = 0; i < each_value_1.length; i += 1) {
		each_blocks[i] = create_each_block_1$5(get_each_context_1$5(ctx, each_value_1, i));
	}

	const block = {
		c: function create() {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			t0 = space();
			div = element("div");
			button0 = element("button");
			button0.textContent = "New";
			t2 = space();
			button1 = element("button");
			button1.textContent = "Remove all";
			t4 = space();
			button2 = element("button");
			button2.textContent = "Change Order";
			attr_dev(button0, "class", "mr1 ph1");
			add_location(button0, file$r, 185, 6, 5488);
			attr_dev(button1, "class", "mh1 ph1");
			add_location(button1, file$r, 187, 6, 5549);
			attr_dev(button2, "class", "mh1 ph1");
			add_location(button2, file$r, 189, 6, 5623);
			attr_dev(div, "class", "flex pv1");
			add_location(div, file$r, 184, 4, 5459);
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
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(button0, "click", /*add*/ ctx[4], false, false, false),
				listen_dev(button1, "click", /*removeAll*/ ctx[6], false, false, false),
				listen_dev(button2, "click", /*toggleSortable*/ ctx[7], false, false, false)
			];
		},
		p: function update(ctx, dirty) {
			if (dirty & /*remove, listData*/ 33) {
				each_value_1 = /*listData*/ ctx[0];
				validate_each_argument(each_value_1);
				let i;

				for (i = 0; i < each_value_1.length; i += 1) {
					const child_ctx = get_each_context_1$5(ctx, each_value_1, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
					} else {
						each_blocks[i] = create_each_block_1$5(child_ctx);
						each_blocks[i].c();
						each_blocks[i].m(t0.parentNode, t0);
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
			if (detaching) detach_dev(t0);
			if (detaching) detach_dev(div);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_1$9.name,
		type: "if",
		source: "(164:2) {#if !enableSortable}",
		ctx
	});

	return block;
}

// (165:4) {#each listData as listItem}
function create_each_block_1$5(ctx) {
	let div2;
	let div0;
	let textarea;
	let t0;
	let input;
	let t1;
	let div1;
	let button;
	let dispose;

	function textarea_input_handler() {
		/*textarea_input_handler*/ ctx[17].call(textarea, /*listItem*/ ctx[23]);
	}

	function input_input_handler() {
		/*input_input_handler*/ ctx[18].call(input, /*listItem*/ ctx[23]);
	}

	function click_handler(...args) {
		return /*click_handler*/ ctx[19](/*listItem*/ ctx[23], ...args);
	}

	const block = {
		c: function create() {
			div2 = element("div");
			div0 = element("div");
			textarea = element("textarea");
			t0 = space();
			input = element("input");
			t1 = space();
			div1 = element("div");
			button = element("button");
			button.textContent = "x";
			attr_dev(textarea, "placeholder", "content");
			attr_dev(textarea, "class", "item item-left svelte-o2y3gy");
			add_location(textarea, file$r, 167, 10, 4997);
			attr_dev(input, "placeholder", "url");
			attr_dev(input, "class", "item item-left mv2 svelte-o2y3gy");
			add_location(input, file$r, 172, 10, 5153);
			attr_dev(div0, "class", "flex flex-column item-left svelte-o2y3gy");
			add_location(div0, file$r, 166, 8, 4946);
			attr_dev(button, "class", "item svelte-o2y3gy");
			add_location(button, file$r, 179, 10, 5346);
			attr_dev(div1, "class", "flex flex-column");
			add_location(div1, file$r, 178, 8, 5305);
			attr_dev(div2, "class", "item-container pv2 bb b--black-05 svelte-o2y3gy");
			add_location(div2, file$r, 165, 6, 4890);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div2, anchor);
			append_dev(div2, div0);
			append_dev(div0, textarea);
			set_input_value(textarea, /*listItem*/ ctx[23].content);
			append_dev(div0, t0);
			append_dev(div0, input);
			set_input_value(input, /*listItem*/ ctx[23].url);
			append_dev(div2, t1);
			append_dev(div2, div1);
			append_dev(div1, button);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(textarea, "input", textarea_input_handler),
				listen_dev(input, "input", input_input_handler),
				listen_dev(button, "click", click_handler, false, false, false)
			];
		},
		p: function update(new_ctx, dirty) {
			ctx = new_ctx;

			if (dirty & /*listData*/ 1) {
				set_input_value(textarea, /*listItem*/ ctx[23].content);
			}

			if (dirty & /*listData*/ 1 && input.value !== /*listItem*/ ctx[23].url) {
				set_input_value(input, /*listItem*/ ctx[23].url);
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div2);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block_1$5.name,
		type: "each",
		source: "(165:4) {#each listData as listItem}",
		ctx
	});

	return block;
}

// (194:2) {#if enableSortable}
function create_if_block$f(ctx) {
	let t0;
	let button;
	let dispose;
	let each_value = /*listData*/ ctx[0];
	validate_each_argument(each_value);
	let each_blocks = [];

	for (let i = 0; i < each_value.length; i += 1) {
		each_blocks[i] = create_each_block$7(get_each_context$7(ctx, each_value, i));
	}

	const block = {
		c: function create() {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			t0 = space();
			button = element("button");

			button.textContent = `
      Finished ordering? (${/*orderHelperText*/ ctx[3]})
    `;

			add_location(button, file$r, 222, 4, 6478);
		},
		m: function mount(target, anchor, remount) {
			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(target, anchor);
			}

			insert_dev(target, t0, anchor);
			insert_dev(target, button, anchor);
			if (remount) dispose();
			dispose = listen_dev(button, "click", /*toggleSortable*/ ctx[7], false, false, false);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*dragstart, dragover, drop, tapHandler, listData*/ 1793) {
				each_value = /*listData*/ ctx[0];
				validate_each_argument(each_value);
				let i;

				for (i = 0; i < each_value.length; i += 1) {
					const child_ctx = get_each_context$7(ctx, each_value, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
					} else {
						each_blocks[i] = create_each_block$7(child_ctx);
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
			if (detaching) detach_dev(button);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$f.name,
		type: "if",
		source: "(194:2) {#if enableSortable}",
		ctx
	});

	return block;
}

// (195:4) {#each listData as listItem, pos}
function create_each_block$7(ctx) {
	let div1;
	let div0;
	let textarea;
	let t;
	let input;
	let div1_data_index_value;
	let tap_action;
	let dispose;

	function textarea_input_handler_1() {
		/*textarea_input_handler_1*/ ctx[20].call(textarea, /*listItem*/ ctx[23]);
	}

	function input_input_handler_1() {
		/*input_input_handler_1*/ ctx[21].call(input, /*listItem*/ ctx[23]);
	}

	const block = {
		c: function create() {
			div1 = element("div");
			div0 = element("div");
			textarea = element("textarea");
			t = space();
			input = element("input");
			attr_dev(textarea, "placeholder", "content");
			attr_dev(textarea, "class", "item item-left svelte-o2y3gy");
			textarea.disabled = true;
			add_location(textarea, file$r, 206, 10, 6106);
			attr_dev(input, "placeholder", "url");
			attr_dev(input, "class", "item item-left mv2 svelte-o2y3gy");
			input.disabled = true;
			add_location(input, file$r, 212, 10, 6283);
			attr_dev(div0, "class", "flex flex-column item-left svelte-o2y3gy");
			add_location(div0, file$r, 205, 8, 6055);
			attr_dev(div1, "draggable", "true");
			attr_dev(div1, "class", "dropzone item-container pv2 bb b--black-05 svelte-o2y3gy");
			attr_dev(div1, "data-index", div1_data_index_value = /*pos*/ ctx[25]);
			add_location(div1, file$r, 195, 6, 5784);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div1, anchor);
			append_dev(div1, div0);
			append_dev(div0, textarea);
			set_input_value(textarea, /*listItem*/ ctx[23].content);
			append_dev(div0, t);
			append_dev(div0, input);
			set_input_value(input, /*listItem*/ ctx[23].url);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(textarea, "input", textarea_input_handler_1),
				listen_dev(input, "input", input_input_handler_1),
				listen_dev(div1, "dragstart", /*dragstart*/ ctx[8], false, false, false),
				listen_dev(div1, "dragover", dragover$3, false, false, false),
				listen_dev(div1, "drop", /*drop*/ ctx[9], false, false, false),
				action_destroyer(tap_action = tap.call(null, div1)),
				listen_dev(div1, "tap", /*tapHandler*/ ctx[10], false, false, false)
			];
		},
		p: function update(new_ctx, dirty) {
			ctx = new_ctx;

			if (dirty & /*listData*/ 1) {
				set_input_value(textarea, /*listItem*/ ctx[23].content);
			}

			if (dirty & /*listData*/ 1 && input.value !== /*listItem*/ ctx[23].url) {
				set_input_value(input, /*listItem*/ ctx[23].url);
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div1);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block$7.name,
		type: "each",
		source: "(195:4) {#each listData as listItem, pos}",
		ctx
	});

	return block;
}

function create_fragment$s(ctx) {
	let h1;
	let t1;
	let div;
	let t2;
	let if_block0 = !/*enableSortable*/ ctx[2] && create_if_block_1$9(ctx);
	let if_block1 = /*enableSortable*/ ctx[2] && create_if_block$f(ctx);

	const block = {
		c: function create() {
			h1 = element("h1");
			h1.textContent = "Items";
			t1 = space();
			div = element("div");
			if (if_block0) if_block0.c();
			t2 = space();
			if (if_block1) if_block1.c();
			add_location(h1, file$r, 160, 0, 4776);
			add_location(div, file$r, 162, 0, 4792);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, h1, anchor);
			insert_dev(target, t1, anchor);
			insert_dev(target, div, anchor);
			if (if_block0) if_block0.m(div, null);
			append_dev(div, t2);
			if (if_block1) if_block1.m(div, null);
			/*div_binding*/ ctx[22](div);
		},
		p: function update(ctx, [dirty]) {
			if (!/*enableSortable*/ ctx[2]) {
				if (if_block0) {
					if_block0.p(ctx, dirty);
				} else {
					if_block0 = create_if_block_1$9(ctx);
					if_block0.c();
					if_block0.m(div, t2);
				}
			} else if (if_block0) {
				if_block0.d(1);
				if_block0 = null;
			}

			if (/*enableSortable*/ ctx[2]) {
				if (if_block1) {
					if_block1.p(ctx, dirty);
				} else {
					if_block1 = create_if_block$f(ctx);
					if_block1.c();
					if_block1.m(div, null);
				}
			} else if (if_block1) {
				if_block1.d(1);
				if_block1 = null;
			}
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(h1);
			if (detaching) detach_dev(t1);
			if (detaching) detach_dev(div);
			if (if_block0) if_block0.d();
			if (if_block1) if_block1.d();
			/*div_binding*/ ctx[22](null);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$s.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function dragover$3(ev) {
	ev.preventDefault();
}

function instance$s($$self, $$props, $$invalidate) {
	let { listData } = $$props;
	console.log("listData", listData);

	const possibleCommands = {
		nothing: "",
		newItem: "When an item is added"
	};

	const isMobile = isDeviceMobile();
	const orderHelperText = !isMobile ? "drag and drop to swap" : "tap to swap";
	const newRow = { content: "", url: "" };

	const _swapItems = {
		from: -1,
		fromElement: null,
		to: -1,
		toElement: null
	};

	let itemsContainer;
	let lastCmd = possibleCommands.nothing;
	let enableSortable = false;
	let swapItems = copyObject(_swapItems);

	afterUpdate(() => {
		if (lastCmd === possibleCommands.newItem) {
			// This only works for V1 elements
			let nodes = itemsContainer.querySelectorAll(".item-container");

			nodes[nodes.length - 1].querySelector("textarea:first-child").focus();
			lastCmd = possibleCommands.nothing;
		}
	});

	function add() {
		$$invalidate(0, listData = listData.concat(copyObject(newRow)));
		lastCmd = possibleCommands.newItem;
	}

	function remove(listItem) {
		$$invalidate(0, listData = listData.filter(t => t !== listItem));

		if (!listData.length) {
			$$invalidate(0, listData = [copyObject(newRow)]);
		}
	}

	function removeAll() {
		$$invalidate(0, listData = [copyObject(newRow)]);
	}

	function toggleSortable(ev) {
		if (listData.length <= 1) {
			alert("nothing to swap");
			return;
		}

		$$invalidate(2, enableSortable = enableSortable ? false : true);

		if (enableSortable) {
			// Reset swapItems
			swapItems = copyObject(_swapItems);
		}
	}

	function dragstart(ev) {
		swapItems = copyObject(_swapItems);
		swapItems.from = ev.target.getAttribute("data-index");
	}

	function drop(ev) {
		ev.preventDefault();
		swapItems.to = ev.target.getAttribute("data-index");

		// We might land on the children, look up for the draggable attribute
		if (swapItems.to == null) {
			swapItems.to = ev.target.closest("[draggable]").getAttribute("data-index");
		}

		let a = listData[swapItems.from];
		let b = listData[swapItems.to];
		$$invalidate(0, listData[swapItems.from] = b, listData);
		$$invalidate(0, listData[swapItems.to] = a, listData);
	}

	function tapHandler(ev) {
		ev.preventDefault();
		let index = ev.target.getAttribute("data-index");

		if (index === null) {
			swapItems = copyObject(_swapItems);
			return;
		}

		if (swapItems.from === -1) {
			swapItems.fromElement = ev.target;
			swapItems.fromElement.style["border-left"] = "solid green";
			swapItems.from = index;
			return;
		}

		if (swapItems.from === index) {
			swapItems.fromElement.style.border = "";
			swapItems = copyObject(_swapItems);
			return;
		}

		swapItems.to = index;
		let a = listData[swapItems.from];
		let b = listData[swapItems.to];
		$$invalidate(0, listData[swapItems.from] = b, listData);
		$$invalidate(0, listData[swapItems.to] = a, listData);
		swapItems.fromElement.style.border = "";
		swapItems.fromElement.style["border-radius"] = "0px";
		swapItems = copyObject(_swapItems);
	}

	const writable_props = ["listData"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$3.warn(`<List_edit_data_v4> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_edit_data_v4", $$slots, []);

	function textarea_input_handler(listItem) {
		listItem.content = this.value;
		$$invalidate(0, listData);
	}

	function input_input_handler(listItem) {
		listItem.url = this.value;
		$$invalidate(0, listData);
	}

	const click_handler = listItem => remove(listItem);

	function textarea_input_handler_1(listItem) {
		listItem.content = this.value;
		$$invalidate(0, listData);
	}

	function input_input_handler_1(listItem) {
		listItem.url = this.value;
		$$invalidate(0, listData);
	}

	function div_binding($$value) {
		binding_callbacks[$$value ? "unshift" : "push"](() => {
			$$invalidate(1, itemsContainer = $$value);
		});
	}

	$$self.$set = $$props => {
		if ("listData" in $$props) $$invalidate(0, listData = $$props.listData);
	};

	$$self.$capture_state = () => ({
		copyObject,
		isDeviceMobile,
		tap,
		afterUpdate,
		listData,
		possibleCommands,
		isMobile,
		orderHelperText,
		newRow,
		_swapItems,
		itemsContainer,
		lastCmd,
		enableSortable,
		swapItems,
		add,
		remove,
		removeAll,
		toggleSortable,
		dragstart,
		dragover: dragover$3,
		drop,
		tapHandler
	});

	$$self.$inject_state = $$props => {
		if ("listData" in $$props) $$invalidate(0, listData = $$props.listData);
		if ("itemsContainer" in $$props) $$invalidate(1, itemsContainer = $$props.itemsContainer);
		if ("lastCmd" in $$props) lastCmd = $$props.lastCmd;
		if ("enableSortable" in $$props) $$invalidate(2, enableSortable = $$props.enableSortable);
		if ("swapItems" in $$props) swapItems = $$props.swapItems;
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [
		listData,
		itemsContainer,
		enableSortable,
		orderHelperText,
		add,
		remove,
		removeAll,
		toggleSortable,
		dragstart,
		drop,
		tapHandler,
		lastCmd,
		swapItems,
		possibleCommands,
		isMobile,
		newRow,
		_swapItems,
		textarea_input_handler,
		input_input_handler,
		click_handler,
		textarea_input_handler_1,
		input_input_handler_1,
		div_binding
	];
}

class List_edit_data_v4 extends SvelteComponentDev {
	constructor(options) {
		super(options);
		if (!document.getElementById("svelte-o2y3gy-style")) add_css$7();
		init(this, options, instance$s, create_fragment$s, safe_not_equal, { listData: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_edit_data_v4",
			options,
			id: create_fragment$s.name
		});

		const { ctx } = this.$$;
		const props = options.props || {};

		if (/*listData*/ ctx[0] === undefined && !("listData" in props)) {
			console_1$3.warn("<List_edit_data_v4> was created without expected prop 'listData'");
		}
	}

	get listData() {
		throw new Error("<List_edit_data_v4>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set listData(value) {
		throw new Error("<List_edit_data_v4>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list_edit_data_todo.svelte generated by Svelte v3.20.1 */

const { console: console_1$4 } = globals;
const file$s = "src/editor/components/list_edit_data_todo.svelte";

function create_fragment$t(ctx) {
	let h1;

	const block = {
		c: function create() {
			h1 = element("h1");
			h1.textContent = "TODO";
			add_location(h1, file$s, 5, 0, 68);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, h1, anchor);
		},
		p: noop,
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(h1);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$t.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$t($$self, $$props, $$invalidate) {
	let { listData } = $$props;
	console.log(listData);
	const writable_props = ["listData"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$4.warn(`<List_edit_data_todo> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_edit_data_todo", $$slots, []);

	$$self.$set = $$props => {
		if ("listData" in $$props) $$invalidate(0, listData = $$props.listData);
	};

	$$self.$capture_state = () => ({ listData });

	$$self.$inject_state = $$props => {
		if ("listData" in $$props) $$invalidate(0, listData = $$props.listData);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [listData];
}

class List_edit_data_todo extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$t, create_fragment$t, safe_not_equal, { listData: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_edit_data_todo",
			options,
			id: create_fragment$t.name
		});

		const { ctx } = this.$$;
		const props = options.props || {};

		if (/*listData*/ ctx[0] === undefined && !("listData" in props)) {
			console_1$4.warn("<List_edit_data_todo> was created without expected prop 'listData'");
		}
	}

	get listData() {
		throw new Error("<List_edit_data_todo>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set listData(value) {
		throw new Error("<List_edit_data_todo>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list_edit_labels.svelte generated by Svelte v3.20.1 */
const file$t = "src/editor/components/list_edit_labels.svelte";

function add_css$8() {
	var style = element("style");
	style.id = "svelte-1963evb-style";
	style.textContent = "input.svelte-1963evb:disabled{background:#fff;color:#333}.container.svelte-1963evb{display:flex}span.svelte-1963evb{text-decoration:underline}span+span.svelte-1963evb{margin-left:0.5em}\n/*# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibGlzdF9lZGl0X2xhYmVscy5zdmVsdGUiLCJzb3VyY2VzIjpbImxpc3RfZWRpdF9sYWJlbHMuc3ZlbHRlIl0sInNvdXJjZXNDb250ZW50IjpbIjxzdHlsZT5cbiAgaW5wdXQ6ZGlzYWJsZWQge1xuICAgIGJhY2tncm91bmQ6ICNmZmY7XG4gICAgY29sb3I6ICMzMzM7XG4gIH1cblxuICAuY29udGFpbmVyIHtcbiAgICBkaXNwbGF5OiBmbGV4O1xuICB9XG5cbiAgc3BhbiB7XG4gICAgdGV4dC1kZWNvcmF0aW9uOiB1bmRlcmxpbmU7XG4gIH1cblxuICBzcGFuICsgc3BhbiB7XG4gICAgbWFyZ2luLWxlZnQ6IDAuNWVtO1xuICB9XG5cbi8qIyBzb3VyY2VNYXBwaW5nVVJMPWRhdGE6YXBwbGljYXRpb24vanNvbjtiYXNlNjQsZXlKMlpYSnphVzl1SWpvekxDSnpiM1Z5WTJWeklqcGJJbk55WXk5bFpHbDBiM0l2WTI5dGNHOXVaVzUwY3k5c2FYTjBYMlZrYVhSZmJHRmlaV3h6TG5OMlpXeDBaU0pkTENKdVlXMWxjeUk2VzEwc0ltMWhjSEJwYm1keklqb2lPMFZCUzBVN1NVRkRSU3huUWtGQlowSTdTVUZEYUVJc1YwRkJWenRGUVVOaU96dEZRVVZCTzBsQlEwVXNZVUZCWVR0RlFVTm1PenRGUVVWQk8wbEJRMFVzTUVKQlFUQkNPMFZCUXpWQ096dEZRVVZCTzBsQlEwVXNhMEpCUVd0Q08wVkJRM0JDSWl3aVptbHNaU0k2SW5OeVl5OWxaR2wwYjNJdlkyOXRjRzl1Wlc1MGN5OXNhWE4wWDJWa2FYUmZiR0ZpWld4ekxuTjJaV3gwWlNJc0luTnZkWEpqWlhORGIyNTBaVzUwSWpwYklseHVJQ0F1WkhKaFoyZGhZbXhsUTI5dWRHRnBibVZ5SUh0Y2JpQWdJQ0J2ZFhSc2FXNWxPaUJ1YjI1bE8xeHVJQ0I5WEc1Y2JpQWdhVzV3ZFhRNlpHbHpZV0pzWldRZ2UxeHVJQ0FnSUdKaFkydG5jbTkxYm1RNklDTm1abVk3WEc0Z0lDQWdZMjlzYjNJNklDTXpNek03WEc0Z0lIMWNibHh1SUNBdVkyOXVkR0ZwYm1WeUlIdGNiaUFnSUNCa2FYTndiR0Y1T2lCbWJHVjRPMXh1SUNCOVhHNWNiaUFnYzNCaGJpQjdYRzRnSUNBZ2RHVjRkQzFrWldOdmNtRjBhVzl1T2lCMWJtUmxjbXhwYm1VN1hHNGdJSDFjYmx4dUlDQnpjR0Z1SUNzZ2MzQmhiaUI3WEc0Z0lDQWdiV0Z5WjJsdUxXeGxablE2SURBdU5XVnRPMXh1SUNCOVhHNGlYWDA9ICovPC9zdHlsZT5cblxuPHNjcmlwdD5cbiAgaW1wb3J0IHsgY29weU9iamVjdCB9IGZyb20gXCIuLi9saWIvaGVscGVyLmpzXCI7XG4gIGltcG9ydCB7IG9uTW91bnQgfSBmcm9tIFwic3ZlbHRlXCI7XG5cbiAgbGV0IGxhYmVsRWxlbWVudDtcblxuICBleHBvcnQgbGV0IGxhYmVscyA9IFtdO1xuXG4gIGZ1bmN0aW9uIGFkZChldmVudCkge1xuICAgIGlmIChsYWJlbEVsZW1lbnQudmFsdWUgPT09IFwiXCIpIHtcbiAgICAgIHJldHVybjtcbiAgICB9XG4gICAgbGFiZWxzID0gbGFiZWxzXG4gICAgICAuZmlsdGVyKGYgPT4gZiAhPT0gbGFiZWxFbGVtZW50LnZhbHVlKVxuICAgICAgLmNvbmNhdChbbGFiZWxFbGVtZW50LnZhbHVlXSk7XG4gICAgbGFiZWxFbGVtZW50LnZhbHVlID0gXCJcIjtcbiAgICBsYWJlbEVsZW1lbnQuZm9jdXMoKTtcbiAgfVxuXG4gIGZ1bmN0aW9uIGVkaXQoaW5wdXQpIHtcbiAgICBsYWJlbEVsZW1lbnQudmFsdWUgPSBpbnB1dDtcbiAgICBsYWJlbEVsZW1lbnQuZm9jdXMoKTtcbiAgfVxuXG4gIGZ1bmN0aW9uIHJlbW92ZSgpIHtcbiAgICBsYWJlbHMgPSBsYWJlbHMuZmlsdGVyKHQgPT4gdCAhPT0gbGFiZWxFbGVtZW50LnZhbHVlKTtcbiAgICBsYWJlbEVsZW1lbnQudmFsdWUgPSBcIlwiO1xuICAgIGxhYmVsRWxlbWVudC5mb2N1cygpO1xuICB9XG5cbiAgbGV0IGVuYWJsZVNvcnRhYmxlID0gZmFsc2U7XG4gIGZ1bmN0aW9uIHRvZ2dsZVNvcnRhYmxlKCkge1xuICAgIGVuYWJsZVNvcnRhYmxlID0gZW5hYmxlU29ydGFibGUgPyBmYWxzZSA6IHRydWU7XG4gIH1cbjwvc2NyaXB0PlxuXG57I2lmICFlbmFibGVTb3J0YWJsZX1cbiAgPGRpdj5cbiAgICA8aW5wdXQgYmluZDp0aGlzPVwie2xhYmVsRWxlbWVudH1cIiBwbGFjZWhvbGRlcj1cIkxhYmVsXCIgLz5cblxuICAgIDxidXR0b24gb246Y2xpY2s9XCJ7YWRkfVwiPkFkZDwvYnV0dG9uPlxuICAgIHsjaWYgbGFiZWxFbGVtZW50ICYmIGxhYmVsRWxlbWVudC52YWx1ZSAhPT0gJyd9XG4gICAgICA8YnV0dG9uIG9uOmNsaWNrPVwie3JlbW92ZX1cIj54PC9idXR0b24+XG4gICAgey9pZn1cblxuICA8L2Rpdj5cbiAgPGRpdiBjbGFzcz1cImNvbnRhaW5lclwiPlxuICAgIHsjZWFjaCBsYWJlbHMgYXMgbGFiZWx9XG4gICAgICA8c3BhbiBjbGFzcz1cIml0ZW1cIiBvbjpjbGljaz1cInsoKSA9PiBlZGl0KGxhYmVsKX1cIj57bGFiZWx9PC9zcGFuPlxuICAgIHsvZWFjaH1cbiAgPC9kaXY+XG57L2lmfVxuIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUNFLG9CQUFLLFNBQVMsQUFBQyxDQUFDLEFBQ2QsVUFBVSxDQUFFLElBQUksQ0FDaEIsS0FBSyxDQUFFLElBQUksQUFDYixDQUFDLEFBRUQsVUFBVSxlQUFDLENBQUMsQUFDVixPQUFPLENBQUUsSUFBSSxBQUNmLENBQUMsQUFFRCxJQUFJLGVBQUMsQ0FBQyxBQUNKLGVBQWUsQ0FBRSxTQUFTLEFBQzVCLENBQUMsQUFFRCxJQUFJLENBQUcsSUFBSSxlQUFDLENBQUMsQUFDWCxXQUFXLENBQUUsS0FBSyxBQUNwQixDQUFDIn0= */";
	append_dev(document.head, style);
}

function get_each_context$8(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[9] = list[i];
	return child_ctx;
}

// (57:0) {#if !enableSortable}
function create_if_block$g(ctx) {
	let div0;
	let input;
	let t0;
	let button;
	let t2;
	let t3;
	let div1;
	let dispose;
	let if_block = /*labelElement*/ ctx[1] && /*labelElement*/ ctx[1].value !== "" && create_if_block_1$a(ctx);
	let each_value = /*labels*/ ctx[0];
	validate_each_argument(each_value);
	let each_blocks = [];

	for (let i = 0; i < each_value.length; i += 1) {
		each_blocks[i] = create_each_block$8(get_each_context$8(ctx, each_value, i));
	}

	const block = {
		c: function create() {
			div0 = element("div");
			input = element("input");
			t0 = space();
			button = element("button");
			button.textContent = "Add";
			t2 = space();
			if (if_block) if_block.c();
			t3 = space();
			div1 = element("div");

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].c();
			}

			attr_dev(input, "placeholder", "Label");
			attr_dev(input, "class", "svelte-1963evb");
			add_location(input, file$t, 58, 4, 1760);
			add_location(button, file$t, 60, 4, 1822);
			add_location(div0, file$t, 57, 2, 1750);
			attr_dev(div1, "class", "container svelte-1963evb");
			add_location(div1, file$t, 66, 2, 1979);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, div0, anchor);
			append_dev(div0, input);
			/*input_binding*/ ctx[7](input);
			append_dev(div0, t0);
			append_dev(div0, button);
			append_dev(div0, t2);
			if (if_block) if_block.m(div0, null);
			insert_dev(target, t3, anchor);
			insert_dev(target, div1, anchor);

			for (let i = 0; i < each_blocks.length; i += 1) {
				each_blocks[i].m(div1, null);
			}

			if (remount) dispose();
			dispose = listen_dev(button, "click", /*add*/ ctx[3], false, false, false);
		},
		p: function update(ctx, dirty) {
			if (/*labelElement*/ ctx[1] && /*labelElement*/ ctx[1].value !== "") {
				if (if_block) {
					if_block.p(ctx, dirty);
				} else {
					if_block = create_if_block_1$a(ctx);
					if_block.c();
					if_block.m(div0, null);
				}
			} else if (if_block) {
				if_block.d(1);
				if_block = null;
			}

			if (dirty & /*edit, labels*/ 17) {
				each_value = /*labels*/ ctx[0];
				validate_each_argument(each_value);
				let i;

				for (i = 0; i < each_value.length; i += 1) {
					const child_ctx = get_each_context$8(ctx, each_value, i);

					if (each_blocks[i]) {
						each_blocks[i].p(child_ctx, dirty);
					} else {
						each_blocks[i] = create_each_block$8(child_ctx);
						each_blocks[i].c();
						each_blocks[i].m(div1, null);
					}
				}

				for (; i < each_blocks.length; i += 1) {
					each_blocks[i].d(1);
				}

				each_blocks.length = each_value.length;
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(div0);
			/*input_binding*/ ctx[7](null);
			if (if_block) if_block.d();
			if (detaching) detach_dev(t3);
			if (detaching) detach_dev(div1);
			destroy_each(each_blocks, detaching);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$g.name,
		type: "if",
		source: "(57:0) {#if !enableSortable}",
		ctx
	});

	return block;
}

// (62:4) {#if labelElement && labelElement.value !== ''}
function create_if_block_1$a(ctx) {
	let button;
	let dispose;

	const block = {
		c: function create() {
			button = element("button");
			button.textContent = "x";
			add_location(button, file$t, 62, 6, 1918);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, button, anchor);
			if (remount) dispose();
			dispose = listen_dev(button, "click", /*remove*/ ctx[5], false, false, false);
		},
		p: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(button);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_1$a.name,
		type: "if",
		source: "(62:4) {#if labelElement && labelElement.value !== ''}",
		ctx
	});

	return block;
}

// (68:4) {#each labels as label}
function create_each_block$8(ctx) {
	let span;
	let t_value = /*label*/ ctx[9] + "";
	let t;
	let dispose;

	function click_handler(...args) {
		return /*click_handler*/ ctx[8](/*label*/ ctx[9], ...args);
	}

	const block = {
		c: function create() {
			span = element("span");
			t = text(t_value);
			attr_dev(span, "class", "item svelte-1963evb");
			add_location(span, file$t, 68, 6, 2037);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, span, anchor);
			append_dev(span, t);
			if (remount) dispose();
			dispose = listen_dev(span, "click", click_handler, false, false, false);
		},
		p: function update(new_ctx, dirty) {
			ctx = new_ctx;
			if (dirty & /*labels*/ 1 && t_value !== (t_value = /*label*/ ctx[9] + "")) set_data_dev(t, t_value);
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(span);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_each_block$8.name,
		type: "each",
		source: "(68:4) {#each labels as label}",
		ctx
	});

	return block;
}

function create_fragment$u(ctx) {
	let if_block_anchor;
	let if_block = !/*enableSortable*/ ctx[2] && create_if_block$g(ctx);

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
			if (!/*enableSortable*/ ctx[2]) {
				if (if_block) {
					if_block.p(ctx, dirty);
				} else {
					if_block = create_if_block$g(ctx);
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
		id: create_fragment$u.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$u($$self, $$props, $$invalidate) {
	let labelElement;
	let { labels = [] } = $$props;

	function add(event) {
		if (labelElement.value === "") {
			return;
		}

		$$invalidate(0, labels = labels.filter(f => f !== labelElement.value).concat([labelElement.value]));
		$$invalidate(1, labelElement.value = "", labelElement);
		labelElement.focus();
	}

	function edit(input) {
		$$invalidate(1, labelElement.value = input, labelElement);
		labelElement.focus();
	}

	function remove() {
		$$invalidate(0, labels = labels.filter(t => t !== labelElement.value));
		$$invalidate(1, labelElement.value = "", labelElement);
		labelElement.focus();
	}

	let enableSortable = false;

	function toggleSortable() {
		$$invalidate(2, enableSortable = enableSortable ? false : true);
	}

	const writable_props = ["labels"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_edit_labels> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_edit_labels", $$slots, []);

	function input_binding($$value) {
		binding_callbacks[$$value ? "unshift" : "push"](() => {
			$$invalidate(1, labelElement = $$value);
		});
	}

	const click_handler = label => edit(label);

	$$self.$set = $$props => {
		if ("labels" in $$props) $$invalidate(0, labels = $$props.labels);
	};

	$$self.$capture_state = () => ({
		copyObject,
		onMount,
		labelElement,
		labels,
		add,
		edit,
		remove,
		enableSortable,
		toggleSortable
	});

	$$self.$inject_state = $$props => {
		if ("labelElement" in $$props) $$invalidate(1, labelElement = $$props.labelElement);
		if ("labels" in $$props) $$invalidate(0, labels = $$props.labels);
		if ("enableSortable" in $$props) $$invalidate(2, enableSortable = $$props.enableSortable);
	};

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	return [
		labels,
		labelElement,
		enableSortable,
		add,
		edit,
		remove,
		toggleSortable,
		input_binding,
		click_handler
	];
}

class List_edit_labels extends SvelteComponentDev {
	constructor(options) {
		super(options);
		if (!document.getElementById("svelte-1963evb-style")) add_css$8();
		init(this, options, instance$u, create_fragment$u, safe_not_equal, { labels: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_edit_labels",
			options,
			id: create_fragment$u.name
		});
	}

	get labels() {
		throw new Error("<List_edit_labels>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set labels(value) {
		throw new Error("<List_edit_labels>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/components/list_edit.svelte generated by Svelte v3.20.1 */

const { Object: Object_1$1, console: console_1$5 } = globals;
const file$u = "src/editor/components/list_edit.svelte";

// (81:0) <Box>
function create_default_slot_9(ctx) {
	let updating_title;
	let current;

	function listedittitle_title_binding(value) {
		/*listedittitle_title_binding*/ ctx[7].call(null, value);
	}

	let listedittitle_props = {};

	if (/*aList*/ ctx[0].info.title !== void 0) {
		listedittitle_props.title = /*aList*/ ctx[0].info.title;
	}

	const listedittitle = new List_edit_title({
			props: listedittitle_props,
			$$inline: true
		});

	binding_callbacks.push(() => bind(listedittitle, "title", listedittitle_title_binding));

	const block = {
		c: function create() {
			create_component(listedittitle.$$.fragment);
		},
		m: function mount(target, anchor) {
			mount_component(listedittitle, target, anchor);
			current = true;
		},
		p: function update(ctx, dirty) {
			const listedittitle_changes = {};

			if (!updating_title && dirty & /*aList*/ 1) {
				updating_title = true;
				listedittitle_changes.title = /*aList*/ ctx[0].info.title;
				add_flush_callback(() => updating_title = false);
			}

			listedittitle.$set(listedittitle_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(listedittitle.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(listedittitle.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(listedittitle, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_default_slot_9.name,
		type: "slot",
		source: "(81:0) <Box>",
		ctx
	});

	return block;
}

// (85:0) <Box>
function create_default_slot_8(ctx) {
	let updating_labels;
	let current;

	function listeditlabels_labels_binding(value) {
		/*listeditlabels_labels_binding*/ ctx[8].call(null, value);
	}

	let listeditlabels_props = {};

	if (/*aList*/ ctx[0].info.labels !== void 0) {
		listeditlabels_props.labels = /*aList*/ ctx[0].info.labels;
	}

	const listeditlabels = new List_edit_labels({
			props: listeditlabels_props,
			$$inline: true
		});

	binding_callbacks.push(() => bind(listeditlabels, "labels", listeditlabels_labels_binding));

	const block = {
		c: function create() {
			create_component(listeditlabels.$$.fragment);
		},
		m: function mount(target, anchor) {
			mount_component(listeditlabels, target, anchor);
			current = true;
		},
		p: function update(ctx, dirty) {
			const listeditlabels_changes = {};

			if (!updating_labels && dirty & /*aList*/ 1) {
				updating_labels = true;
				listeditlabels_changes.labels = /*aList*/ ctx[0].info.labels;
				add_flush_callback(() => updating_labels = false);
			}

			listeditlabels.$set(listeditlabels_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(listeditlabels.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(listeditlabels.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(listeditlabels, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_default_slot_8.name,
		type: "slot",
		source: "(85:0) <Box>",
		ctx
	});

	return block;
}

// (89:0) <Box>
function create_default_slot_7(ctx) {
	let updating_listData;
	let switch_instance_anchor;
	let current;

	function switch_instance_listData_binding(value) {
		/*switch_instance_listData_binding*/ ctx[9].call(null, value);
	}

	var switch_value = /*renderItem*/ ctx[2];

	function switch_props(ctx) {
		let switch_instance_props = {};

		if (/*aList*/ ctx[0].data !== void 0) {
			switch_instance_props.listData = /*aList*/ ctx[0].data;
		}

		return {
			props: switch_instance_props,
			$$inline: true
		};
	}

	if (switch_value) {
		var switch_instance = new switch_value(switch_props(ctx));
		binding_callbacks.push(() => bind(switch_instance, "listData", switch_instance_listData_binding));
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

			if (!updating_listData && dirty & /*aList*/ 1) {
				updating_listData = true;
				switch_instance_changes.listData = /*aList*/ ctx[0].data;
				add_flush_callback(() => updating_listData = false);
			}

			if (switch_value !== (switch_value = /*renderItem*/ ctx[2])) {
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
					binding_callbacks.push(() => bind(switch_instance, "listData", switch_instance_listData_binding));
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
		id: create_default_slot_7.name,
		type: "slot",
		source: "(89:0) <Box>",
		ctx
	});

	return block;
}

// (92:0) <Box>
function create_default_slot_6(ctx) {
	let button0;
	let t1;
	let button1;
	let dispose;

	const block = {
		c: function create() {
			button0 = element("button");
			button0.textContent = "Save";
			t1 = space();
			button1 = element("button");
			button1.textContent = "Cancel";
			add_location(button0, file$u, 92, 2, 2422);
			add_location(button1, file$u, 93, 2, 2464);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, button0, anchor);
			insert_dev(target, t1, anchor);
			insert_dev(target, button1, anchor);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(button0, "click", /*save*/ ctx[5], false, false, false),
				listen_dev(button1, "click", /*cancel*/ ctx[3], false, false, false)
			];
		},
		p: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(button0);
			if (detaching) detach_dev(t1);
			if (detaching) detach_dev(button1);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_default_slot_6.name,
		type: "slot",
		source: "(92:0) <Box>",
		ctx
	});

	return block;
}

// (99:2) <Box>
function create_default_slot_5(ctx) {
	let h2;
	let t1;
	let label0;
	let input0;
	let t2;
	let t3;
	let label1;
	let input1;
	let t4;
	let t5;
	let label2;
	let input2;
	let t6;
	let dispose;

	const block = {
		c: function create() {
			h2 = element("h2");
			h2.textContent = "Share";
			t1 = space();
			label0 = element("label");
			input0 = element("input");
			t2 = text("\n      Private");
			t3 = space();
			label1 = element("label");
			input1 = element("input");
			t4 = text("\n      Public");
			t5 = space();
			label2 = element("label");
			input2 = element("input");
			t6 = text("\n      Friends");
			add_location(h2, file$u, 99, 4, 2554);
			attr_dev(input0, "type", "radio");
			input0.__value = "private";
			input0.value = input0.__value;
			/*$$binding_groups*/ ctx[11][1].push(input0);
			add_location(input0, file$u, 101, 6, 2587);
			add_location(label0, file$u, 100, 4, 2573);
			attr_dev(input1, "type", "radio");
			input1.__value = "public";
			input1.value = input1.__value;
			/*$$binding_groups*/ ctx[11][1].push(input1);
			add_location(input1, file$u, 109, 6, 2739);
			add_location(label1, file$u, 108, 4, 2725);
			attr_dev(input2, "type", "radio");
			input2.__value = "friends";
			input2.value = input2.__value;
			/*$$binding_groups*/ ctx[11][1].push(input2);
			add_location(input2, file$u, 117, 6, 2889);
			add_location(label2, file$u, 116, 4, 2875);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, h2, anchor);
			insert_dev(target, t1, anchor);
			insert_dev(target, label0, anchor);
			append_dev(label0, input0);
			input0.checked = input0.__value === /*aList*/ ctx[0].info.shared_with;
			append_dev(label0, t2);
			insert_dev(target, t3, anchor);
			insert_dev(target, label1, anchor);
			append_dev(label1, input1);
			input1.checked = input1.__value === /*aList*/ ctx[0].info.shared_with;
			append_dev(label1, t4);
			insert_dev(target, t5, anchor);
			insert_dev(target, label2, anchor);
			append_dev(label2, input2);
			input2.checked = input2.__value === /*aList*/ ctx[0].info.shared_with;
			append_dev(label2, t6);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(input0, "change", /*input0_change_handler*/ ctx[10]),
				listen_dev(input1, "change", /*input1_change_handler*/ ctx[12]),
				listen_dev(input2, "change", /*input2_change_handler*/ ctx[13])
			];
		},
		p: function update(ctx, dirty) {
			if (dirty & /*aList*/ 1) {
				input0.checked = input0.__value === /*aList*/ ctx[0].info.shared_with;
			}

			if (dirty & /*aList*/ 1) {
				input1.checked = input1.__value === /*aList*/ ctx[0].info.shared_with;
			}

			if (dirty & /*aList*/ 1) {
				input2.checked = input2.__value === /*aList*/ ctx[0].info.shared_with;
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(h2);
			if (detaching) detach_dev(t1);
			if (detaching) detach_dev(label0);
			/*$$binding_groups*/ ctx[11][1].splice(/*$$binding_groups*/ ctx[11][1].indexOf(input0), 1);
			if (detaching) detach_dev(t3);
			if (detaching) detach_dev(label1);
			/*$$binding_groups*/ ctx[11][1].splice(/*$$binding_groups*/ ctx[11][1].indexOf(input1), 1);
			if (detaching) detach_dev(t5);
			if (detaching) detach_dev(label2);
			/*$$binding_groups*/ ctx[11][1].splice(/*$$binding_groups*/ ctx[11][1].indexOf(input2), 1);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_default_slot_5.name,
		type: "slot",
		source: "(99:2) <Box>",
		ctx
	});

	return block;
}

// (127:2) {#if canInteract}
function create_if_block$h(ctx) {
	let current;

	const box = new Box({
			props: {
				$$slots: { default: [create_default_slot_3] },
				$$scope: { ctx }
			},
			$$inline: true
		});

	const block = {
		c: function create() {
			create_component(box.$$.fragment);
		},
		m: function mount(target, anchor) {
			mount_component(box, target, anchor);
			current = true;
		},
		p: function update(ctx, dirty) {
			const box_changes = {};

			if (dirty & /*$$scope, aList*/ 65537) {
				box_changes.$$scope = { dirty, ctx };
			}

			box.$set(box_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(box.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(box.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(box, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$h.name,
		type: "if",
		source: "(127:2) {#if canInteract}",
		ctx
	});

	return block;
}

// (130:6) <Box>
function create_default_slot_4(ctx) {
	let h3;
	let t1;
	let label0;
	let input0;
	let t2;
	let t3;
	let label1;
	let input1;
	let t4;
	let dispose;

	const block = {
		c: function create() {
			h3 = element("h3");
			h3.textContent = "Slideshow";
			t1 = space();
			label0 = element("label");
			input0 = element("input");
			t2 = text("\n          Disable");
			t3 = space();
			label1 = element("label");
			input1 = element("input");
			t4 = text("\n          Enable");
			add_location(h3, file$u, 130, 8, 3107);
			attr_dev(input0, "type", "radio");
			input0.__value = "0";
			input0.value = input0.__value;
			/*$$binding_groups*/ ctx[11][0].push(input0);
			add_location(input0, file$u, 132, 10, 3152);
			add_location(label0, file$u, 131, 8, 3134);
			attr_dev(input1, "type", "radio");
			input1.__value = "1";
			input1.value = input1.__value;
			/*$$binding_groups*/ ctx[11][0].push(input1);
			add_location(input1, file$u, 141, 10, 3338);
			add_location(label1, file$u, 140, 8, 3320);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, h3, anchor);
			insert_dev(target, t1, anchor);
			insert_dev(target, label0, anchor);
			append_dev(label0, input0);
			input0.checked = input0.__value === /*aList*/ ctx[0].info.interact.slideshow;
			append_dev(label0, t2);
			insert_dev(target, t3, anchor);
			insert_dev(target, label1, anchor);
			append_dev(label1, input1);
			input1.checked = input1.__value === /*aList*/ ctx[0].info.interact.slideshow;
			append_dev(label1, t4);
			if (remount) run_all(dispose);

			dispose = [
				listen_dev(input0, "change", /*input0_change_handler_1*/ ctx[14]),
				listen_dev(input1, "change", /*input1_change_handler_1*/ ctx[15])
			];
		},
		p: function update(ctx, dirty) {
			if (dirty & /*aList*/ 1) {
				input0.checked = input0.__value === /*aList*/ ctx[0].info.interact.slideshow;
			}

			if (dirty & /*aList*/ 1) {
				input1.checked = input1.__value === /*aList*/ ctx[0].info.interact.slideshow;
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(h3);
			if (detaching) detach_dev(t1);
			if (detaching) detach_dev(label0);
			/*$$binding_groups*/ ctx[11][0].splice(/*$$binding_groups*/ ctx[11][0].indexOf(input0), 1);
			if (detaching) detach_dev(t3);
			if (detaching) detach_dev(label1);
			/*$$binding_groups*/ ctx[11][0].splice(/*$$binding_groups*/ ctx[11][0].indexOf(input1), 1);
			run_all(dispose);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_default_slot_4.name,
		type: "slot",
		source: "(130:6) <Box>",
		ctx
	});

	return block;
}

// (128:4) <Box>
function create_default_slot_3(ctx) {
	let h2;
	let t1;
	let current;

	const box = new Box({
			props: {
				$$slots: { default: [create_default_slot_4] },
				$$scope: { ctx }
			},
			$$inline: true
		});

	const block = {
		c: function create() {
			h2 = element("h2");
			h2.textContent = "Interact";
			t1 = space();
			create_component(box.$$.fragment);
			add_location(h2, file$u, 128, 6, 3069);
		},
		m: function mount(target, anchor) {
			insert_dev(target, h2, anchor);
			insert_dev(target, t1, anchor);
			mount_component(box, target, anchor);
			current = true;
		},
		p: function update(ctx, dirty) {
			const box_changes = {};

			if (dirty & /*$$scope, aList*/ 65537) {
				box_changes.$$scope = { dirty, ctx };
			}

			box.$set(box_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(box.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(box.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(h2);
			if (detaching) detach_dev(t1);
			destroy_component(box, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_default_slot_3.name,
		type: "slot",
		source: "(128:4) <Box>",
		ctx
	});

	return block;
}

// (155:4) <Box>
function create_default_slot_2(ctx) {
	let button;
	let dispose;

	const block = {
		c: function create() {
			button = element("button");
			button.textContent = "Delete this list forever";
			add_location(button, file$u, 155, 6, 3573);
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, button, anchor);
			if (remount) dispose();
			dispose = listen_dev(button, "click", /*remove*/ ctx[4], false, false, false);
		},
		p: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(button);
			dispose();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_default_slot_2.name,
		type: "slot",
		source: "(155:4) <Box>",
		ctx
	});

	return block;
}

// (153:2) <Box>
function create_default_slot_1(ctx) {
	let h1;
	let t1;
	let current;

	const box = new Box({
			props: {
				$$slots: { default: [create_default_slot_2] },
				$$scope: { ctx }
			},
			$$inline: true
		});

	const block = {
		c: function create() {
			h1 = element("h1");
			h1.textContent = "Danger";
			t1 = space();
			create_component(box.$$.fragment);
			add_location(h1, file$u, 153, 4, 3541);
		},
		m: function mount(target, anchor) {
			insert_dev(target, h1, anchor);
			insert_dev(target, t1, anchor);
			mount_component(box, target, anchor);
			current = true;
		},
		p: function update(ctx, dirty) {
			const box_changes = {};

			if (dirty & /*$$scope*/ 65536) {
				box_changes.$$scope = { dirty, ctx };
			}

			box.$set(box_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(box.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(box.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(h1);
			if (detaching) detach_dev(t1);
			destroy_component(box, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_default_slot_1.name,
		type: "slot",
		source: "(153:2) <Box>",
		ctx
	});

	return block;
}

// (97:0) <Box>
function create_default_slot(ctx) {
	let h1;
	let t1;
	let t2;
	let t3;
	let current;

	const box0 = new Box({
			props: {
				$$slots: { default: [create_default_slot_5] },
				$$scope: { ctx }
			},
			$$inline: true
		});

	let if_block = /*canInteract*/ ctx[1] && create_if_block$h(ctx);

	const box1 = new Box({
			props: {
				$$slots: { default: [create_default_slot_1] },
				$$scope: { ctx }
			},
			$$inline: true
		});

	const block = {
		c: function create() {
			h1 = element("h1");
			h1.textContent = "Advanced";
			t1 = space();
			create_component(box0.$$.fragment);
			t2 = space();
			if (if_block) if_block.c();
			t3 = space();
			create_component(box1.$$.fragment);
			add_location(h1, file$u, 97, 2, 2524);
		},
		m: function mount(target, anchor) {
			insert_dev(target, h1, anchor);
			insert_dev(target, t1, anchor);
			mount_component(box0, target, anchor);
			insert_dev(target, t2, anchor);
			if (if_block) if_block.m(target, anchor);
			insert_dev(target, t3, anchor);
			mount_component(box1, target, anchor);
			current = true;
		},
		p: function update(ctx, dirty) {
			const box0_changes = {};

			if (dirty & /*$$scope, aList*/ 65537) {
				box0_changes.$$scope = { dirty, ctx };
			}

			box0.$set(box0_changes);

			if (/*canInteract*/ ctx[1]) {
				if (if_block) {
					if_block.p(ctx, dirty);
					transition_in(if_block, 1);
				} else {
					if_block = create_if_block$h(ctx);
					if_block.c();
					transition_in(if_block, 1);
					if_block.m(t3.parentNode, t3);
				}
			} else if (if_block) {
				group_outros();

				transition_out(if_block, 1, 1, () => {
					if_block = null;
				});

				check_outros();
			}

			const box1_changes = {};

			if (dirty & /*$$scope*/ 65536) {
				box1_changes.$$scope = { dirty, ctx };
			}

			box1.$set(box1_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(box0.$$.fragment, local);
			transition_in(if_block);
			transition_in(box1.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(box0.$$.fragment, local);
			transition_out(if_block);
			transition_out(box1.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(h1);
			if (detaching) detach_dev(t1);
			destroy_component(box0, detaching);
			if (detaching) detach_dev(t2);
			if (if_block) if_block.d(detaching);
			if (detaching) detach_dev(t3);
			destroy_component(box1, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_default_slot.name,
		type: "slot",
		source: "(97:0) <Box>",
		ctx
	});

	return block;
}

function create_fragment$v(ctx) {
	let t0;
	let t1;
	let t2;
	let t3;
	let current;

	const box0 = new Box({
			props: {
				$$slots: { default: [create_default_slot_9] },
				$$scope: { ctx }
			},
			$$inline: true
		});

	const box1 = new Box({
			props: {
				$$slots: { default: [create_default_slot_8] },
				$$scope: { ctx }
			},
			$$inline: true
		});

	const box2 = new Box({
			props: {
				$$slots: { default: [create_default_slot_7] },
				$$scope: { ctx }
			},
			$$inline: true
		});

	const box3 = new Box({
			props: {
				$$slots: { default: [create_default_slot_6] },
				$$scope: { ctx }
			},
			$$inline: true
		});

	const box4 = new Box({
			props: {
				$$slots: { default: [create_default_slot] },
				$$scope: { ctx }
			},
			$$inline: true
		});

	const block = {
		c: function create() {
			create_component(box0.$$.fragment);
			t0 = space();
			create_component(box1.$$.fragment);
			t1 = space();
			create_component(box2.$$.fragment);
			t2 = space();
			create_component(box3.$$.fragment);
			t3 = space();
			create_component(box4.$$.fragment);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			mount_component(box0, target, anchor);
			insert_dev(target, t0, anchor);
			mount_component(box1, target, anchor);
			insert_dev(target, t1, anchor);
			mount_component(box2, target, anchor);
			insert_dev(target, t2, anchor);
			mount_component(box3, target, anchor);
			insert_dev(target, t3, anchor);
			mount_component(box4, target, anchor);
			current = true;
		},
		p: function update(ctx, [dirty]) {
			const box0_changes = {};

			if (dirty & /*$$scope, aList*/ 65537) {
				box0_changes.$$scope = { dirty, ctx };
			}

			box0.$set(box0_changes);
			const box1_changes = {};

			if (dirty & /*$$scope, aList*/ 65537) {
				box1_changes.$$scope = { dirty, ctx };
			}

			box1.$set(box1_changes);
			const box2_changes = {};

			if (dirty & /*$$scope, aList*/ 65537) {
				box2_changes.$$scope = { dirty, ctx };
			}

			box2.$set(box2_changes);
			const box3_changes = {};

			if (dirty & /*$$scope*/ 65536) {
				box3_changes.$$scope = { dirty, ctx };
			}

			box3.$set(box3_changes);
			const box4_changes = {};

			if (dirty & /*$$scope, aList, canInteract*/ 65539) {
				box4_changes.$$scope = { dirty, ctx };
			}

			box4.$set(box4_changes);
		},
		i: function intro(local) {
			if (current) return;
			transition_in(box0.$$.fragment, local);
			transition_in(box1.$$.fragment, local);
			transition_in(box2.$$.fragment, local);
			transition_in(box3.$$.fragment, local);
			transition_in(box4.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(box0.$$.fragment, local);
			transition_out(box1.$$.fragment, local);
			transition_out(box2.$$.fragment, local);
			transition_out(box3.$$.fragment, local);
			transition_out(box4.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(box0, detaching);
			if (detaching) detach_dev(t0);
			destroy_component(box1, detaching);
			if (detaching) detach_dev(t1);
			destroy_component(box2, detaching);
			if (detaching) detach_dev(t2);
			destroy_component(box3, detaching);
			if (detaching) detach_dev(t3);
			destroy_component(box4, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$v.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$v($$self, $$props, $$invalidate) {
	let { aList } = $$props;

	let listTypes = {
		v1: List_edit_data_v1,
		v2: List_edit_data_v2,
		v3: List_edit_data_v3,
		v4: List_edit_data_v4
	};

	let renderItem = Object.keys(listTypes).filter(key => aList.info.type === key).reduce(
		(notFound, key) => {
			return listTypes[key];
		},
		List_edit_data_todo
	);

	if (!aList.info.hasOwnProperty("interact") || !aList.info.interact.hasOwnProperty("slideshow")) {
		aList.info.interact = { slideshow: "0" };
	}

	function cancel() {
		listsEdits.remove(aList.uuid);
		paths.list.view(aList.uuid);
	}

	async function remove() {
		const response = await deleteList(aList.uuid);

		if (response.status !== 200) {
			alert("failed try again");
			console.log("status from server was", response.status);
			return;
		}

		myLists.remove(aList.uuid);
		listsEdits.remove(aList.uuid);

		// TODO how to remove /lists/view/:uuid as well
		replace("/list/deleted");
	}

	async function save() {
		const response = await putList(aList);

		if (response.status !== 200) {
			alert("failed try again");
			console.log("status from server was", response.status);
			return;
		}

		try {
			listsEdits.remove(aList.uuid);
			myLists.update(aList);
			paths.list.view(aList.uuid);
		} catch(e) {
			alert("failed to clean up your edits");
		}
	}

	const writable_props = ["aList"];

	Object_1$1.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$5.warn(`<List_edit> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_edit", $$slots, []);
	const $$binding_groups = [[], []];

	function listedittitle_title_binding(value) {
		aList.info.title = value;
		$$invalidate(0, aList);
	}

	function listeditlabels_labels_binding(value) {
		aList.info.labels = value;
		$$invalidate(0, aList);
	}

	function switch_instance_listData_binding(value) {
		aList.data = value;
		$$invalidate(0, aList);
	}

	function input0_change_handler() {
		aList.info.shared_with = this.__value;
		$$invalidate(0, aList);
	}

	function input1_change_handler() {
		aList.info.shared_with = this.__value;
		$$invalidate(0, aList);
	}

	function input2_change_handler() {
		aList.info.shared_with = this.__value;
		$$invalidate(0, aList);
	}

	function input0_change_handler_1() {
		aList.info.interact.slideshow = this.__value;
		$$invalidate(0, aList);
	}

	function input1_change_handler_1() {
		aList.info.interact.slideshow = this.__value;
		$$invalidate(0, aList);
	}

	$$self.$set = $$props => {
		if ("aList" in $$props) $$invalidate(0, aList = $$props.aList);
	};

	$$self.$capture_state = () => ({
		replace,
		cache,
		putList,
		deleteList,
		goto: paths,
		myLists,
		listsEdits,
		Box,
		ListEditTitle: List_edit_title,
		ListEditDataV1: List_edit_data_v1,
		ListEditDataV2: List_edit_data_v2,
		ListEditDataV3: List_edit_data_v3,
		ListEditDataV4: List_edit_data_v4,
		ListEditDataTodo: List_edit_data_todo,
		ListEditLabels: List_edit_labels,
		aList,
		listTypes,
		renderItem,
		cancel,
		remove,
		save,
		canInteract
	});

	$$self.$inject_state = $$props => {
		if ("aList" in $$props) $$invalidate(0, aList = $$props.aList);
		if ("listTypes" in $$props) listTypes = $$props.listTypes;
		if ("renderItem" in $$props) $$invalidate(2, renderItem = $$props.renderItem);
		if ("canInteract" in $$props) $$invalidate(1, canInteract = $$props.canInteract);
	};

	let canInteract;

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	$$self.$$.update = () => {
		if ($$self.$$.dirty & /*aList*/ 1) {
			 $$invalidate(1, canInteract = aList && aList.info.type === "v1");
		}

		if ($$self.$$.dirty & /*aList*/ 1) {
			 listsEdits.update(aList);
		}
	};

	return [
		aList,
		canInteract,
		renderItem,
		cancel,
		remove,
		save,
		listTypes,
		listedittitle_title_binding,
		listeditlabels_labels_binding,
		switch_instance_listData_binding,
		input0_change_handler,
		$$binding_groups,
		input1_change_handler,
		input2_change_handler,
		input0_change_handler_1,
		input1_change_handler_1
	];
}

class List_edit extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$v, create_fragment$v, safe_not_equal, { aList: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_edit",
			options,
			id: create_fragment$v.name
		});

		const { ctx } = this.$$;
		const props = options.props || {};

		if (/*aList*/ ctx[0] === undefined && !("aList" in props)) {
			console_1$5.warn("<List_edit> was created without expected prop 'aList'");
		}
	}

	get aList() {
		throw new Error("<List_edit>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set aList(value) {
		throw new Error("<List_edit>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/routes/list_edit.svelte generated by Svelte v3.20.1 */
const file$v = "src/editor/routes/list_edit.svelte";

// (15:0) {:else}
function create_else_block$6(ctx) {
	let p;
	let t0;
	let span;
	let t1_value = /*params*/ ctx[0].uuid + "";
	let t1;

	const block = {
		c: function create() {
			p = element("p");
			t0 = text("list uuid is\n    ");
			span = element("span");
			t1 = text(t1_value);
			add_location(span, file$v, 17, 4, 382);
			add_location(p, file$v, 15, 2, 357);
		},
		m: function mount(target, anchor) {
			insert_dev(target, p, anchor);
			append_dev(p, t0);
			append_dev(p, span);
			append_dev(span, t1);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*params*/ 1 && t1_value !== (t1_value = /*params*/ ctx[0].uuid + "")) set_data_dev(t1, t1_value);
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(p);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_else_block$6.name,
		type: "else",
		source: "(15:0) {:else}",
		ctx
	});

	return block;
}

// (13:0) {#if exists}
function create_if_block$i(ctx) {
	let current;

	const listedit = new List_edit({
			props: { aList: /*aList*/ ctx[2] },
			$$inline: true
		});

	const block = {
		c: function create() {
			create_component(listedit.$$.fragment);
		},
		m: function mount(target, anchor) {
			mount_component(listedit, target, anchor);
			current = true;
		},
		p: noop,
		i: function intro(local) {
			if (current) return;
			transition_in(listedit.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(listedit.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(listedit, detaching);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$i.name,
		type: "if",
		source: "(13:0) {#if exists}",
		ctx
	});

	return block;
}

function create_fragment$w(ctx) {
	let current_block_type_index;
	let if_block;
	let if_block_anchor;
	let current;
	const if_block_creators = [create_if_block$i, create_else_block$6];
	const if_blocks = [];

	function select_block_type(ctx, dirty) {
		if (/*exists*/ ctx[1]) return 0;
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
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
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
		id: create_fragment$w.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$w($$self, $$props, $$invalidate) {
	let { params = {} } = $$props;
	const aList = listsEdits.find(params.uuid);
	const writable_props = ["params"];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_edit> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_edit", $$slots, []);

	$$self.$set = $$props => {
		if ("params" in $$props) $$invalidate(0, params = $$props.params);
	};

	$$self.$capture_state = () => ({
		cache,
		listsEdits,
		Box,
		ListEdit: List_edit,
		params,
		aList,
		exists
	});

	$$self.$inject_state = $$props => {
		if ("params" in $$props) $$invalidate(0, params = $$props.params);
		if ("exists" in $$props) $$invalidate(1, exists = $$props.exists);
	};

	let exists;

	if ($$props && "$$inject" in $$props) {
		$$self.$inject_state($$props.$$inject);
	}

	 $$invalidate(1, exists = !!aList);
	return [params, exists, aList];
}

class List_edit$1 extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$w, create_fragment$w, safe_not_equal, { params: 0 });

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_edit",
			options,
			id: create_fragment$w.name
		});
	}

	get params() {
		throw new Error("<List_edit>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}

	set params(value) {
		throw new Error("<List_edit>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
	}
}

/* src/editor/routes/list_deleted.svelte generated by Svelte v3.20.1 */

const file$w = "src/editor/routes/list_deleted.svelte";

function create_fragment$x(ctx) {
	let p;

	const block = {
		c: function create() {
			p = element("p");
			p.textContent = "The list, which we wont speak mention, has been deleted forever.";
			add_location(p, file$w, 0, 0, 0);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, p, anchor);
		},
		p: noop,
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(p);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$x.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$x($$self, $$props) {
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<List_deleted> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("List_deleted", $$slots, []);
	return [];
}

class List_deleted extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$x, create_fragment$x, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "List_deleted",
			options,
			id: create_fragment$x.name
		});
	}
}

const {subscribe: subscribe$3, set: set$2, update: update$3} = writable({
  "gitHash":"na",
  "gitDate":"na",
  "version":"na",
  "url":"https://github.com/freshteapot/learnalist-api"}
);

const loading$1 = writable(false);
const error$1 = writable('');

const VersionStore = () => ({
    subscribe: subscribe$3,
    set: set$2,
    loading: loading$1,
    error: error$1,
    async get(query) {
      console.log("Not here");
      try {
          error$1.set('');
          loading$1.set(true);
          const response = await getVersion();
          loading$1.set(false);
          set$2(response);
          return response;
      } catch(e) {
          loading$1.set(false);
          set$2({
            "gitHash":"na",
            "gitDate":"na",
            "version":"na",
            "url":"https://github.com/freshteapot/learnalist-api"});
          error$1.set(`Error has been occurred. Details: ${e.message}`);
      }
    }
});

var version = VersionStore();

/* src/editor/routes/settings_server_information.svelte generated by Svelte v3.20.1 */
const file$x = "src/editor/routes/settings_server_information.svelte";

// (13:4) {:else}
function create_else_block$7(ctx) {
	let p0;
	let t0;
	let t1_value = /*$version*/ ctx[2].version + "";
	let t1;
	let t2;
	let p1;
	let t3;
	let t4_value = /*$version*/ ctx[2].gitHash + "";
	let t4;
	let t5;
	let p2;
	let t6;
	let t7_value = /*$version*/ ctx[2].gitDate + "";
	let t7;
	let t8;
	let p3;
	let t9;
	let a;
	let t10;
	let a_href_value;

	const block = {
		c: function create() {
			p0 = element("p");
			t0 = text("Version is ");
			t1 = text(t1_value);
			t2 = space();
			p1 = element("p");
			t3 = text("Git hash is ");
			t4 = text(t4_value);
			t5 = space();
			p2 = element("p");
			t6 = text("Git date is ");
			t7 = text(t7_value);
			t8 = space();
			p3 = element("p");
			t9 = text("On\n        ");
			a = element("a");
			t10 = text("Github");
			add_location(p0, file$x, 13, 6, 277);
			add_location(p1, file$x, 14, 6, 320);
			add_location(p2, file$x, 15, 6, 364);
			attr_dev(a, "href", a_href_value = /*$version*/ ctx[2].url);
			attr_dev(a, "target", "_blank");
			add_location(a, file$x, 18, 8, 431);
			add_location(p3, file$x, 16, 6, 408);
		},
		m: function mount(target, anchor) {
			insert_dev(target, p0, anchor);
			append_dev(p0, t0);
			append_dev(p0, t1);
			insert_dev(target, t2, anchor);
			insert_dev(target, p1, anchor);
			append_dev(p1, t3);
			append_dev(p1, t4);
			insert_dev(target, t5, anchor);
			insert_dev(target, p2, anchor);
			append_dev(p2, t6);
			append_dev(p2, t7);
			insert_dev(target, t8, anchor);
			insert_dev(target, p3, anchor);
			append_dev(p3, t9);
			append_dev(p3, a);
			append_dev(a, t10);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*$version*/ 4 && t1_value !== (t1_value = /*$version*/ ctx[2].version + "")) set_data_dev(t1, t1_value);
			if (dirty & /*$version*/ 4 && t4_value !== (t4_value = /*$version*/ ctx[2].gitHash + "")) set_data_dev(t4, t4_value);
			if (dirty & /*$version*/ 4 && t7_value !== (t7_value = /*$version*/ ctx[2].gitDate + "")) set_data_dev(t7, t7_value);

			if (dirty & /*$version*/ 4 && a_href_value !== (a_href_value = /*$version*/ ctx[2].url)) {
				attr_dev(a, "href", a_href_value);
			}
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(p0);
			if (detaching) detach_dev(t2);
			if (detaching) detach_dev(p1);
			if (detaching) detach_dev(t5);
			if (detaching) detach_dev(p2);
			if (detaching) detach_dev(t8);
			if (detaching) detach_dev(p3);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_else_block$7.name,
		type: "else",
		source: "(13:4) {:else}",
		ctx
	});

	return block;
}

// (11:23) 
function create_if_block_1$b(ctx) {
	let t;

	const block = {
		c: function create() {
			t = text("Loading...");
		},
		m: function mount(target, anchor) {
			insert_dev(target, t, anchor);
		},
		p: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(t);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block_1$b.name,
		type: "if",
		source: "(11:23) ",
		ctx
	});

	return block;
}

// (9:4) {#if $error}
function create_if_block$j(ctx) {
	let t0;
	let t1;

	const block = {
		c: function create() {
			t0 = text("error is ");
			t1 = text(/*$error*/ ctx[0]);
		},
		m: function mount(target, anchor) {
			insert_dev(target, t0, anchor);
			insert_dev(target, t1, anchor);
		},
		p: function update(ctx, dirty) {
			if (dirty & /*$error*/ 1) set_data_dev(t1, /*$error*/ ctx[0]);
		},
		d: function destroy(detaching) {
			if (detaching) detach_dev(t0);
			if (detaching) detach_dev(t1);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_if_block$j.name,
		type: "if",
		source: "(9:4) {#if $error}",
		ctx
	});

	return block;
}

function create_fragment$y(ctx) {
	let div1;
	let div0;

	function select_block_type(ctx, dirty) {
		if (/*$error*/ ctx[0]) return create_if_block$j;
		if (/*$loading*/ ctx[1]) return create_if_block_1$b;
		return create_else_block$7;
	}

	let current_block_type = select_block_type(ctx);
	let if_block = current_block_type(ctx);

	const block = {
		c: function create() {
			div1 = element("div");
			div0 = element("div");
			if_block.c();
			attr_dev(div0, "class", "pl0 measure center");
			add_location(div0, file$x, 7, 2, 144);
			attr_dev(div1, "class", "pa3 pa5-ns");
			add_location(div1, file$x, 6, 0, 117);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			insert_dev(target, div1, anchor);
			append_dev(div1, div0);
			if_block.m(div0, null);
		},
		p: function update(ctx, [dirty]) {
			if (current_block_type === (current_block_type = select_block_type(ctx)) && if_block) {
				if_block.p(ctx, dirty);
			} else {
				if_block.d(1);
				if_block = current_block_type(ctx);

				if (if_block) {
					if_block.c();
					if_block.m(div0, null);
				}
			}
		},
		i: noop,
		o: noop,
		d: function destroy(detaching) {
			if (detaching) detach_dev(div1);
			if_block.d();
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$y.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function instance$y($$self, $$props, $$invalidate) {
	let $error;
	let $loading;
	let $version;
	validate_store(version, "version");
	component_subscribe($$self, version, $$value => $$invalidate(2, $version = $$value));
	const { loading, error } = version;
	validate_store(loading, "loading");
	component_subscribe($$self, loading, value => $$invalidate(1, $loading = value));
	validate_store(error, "error");
	component_subscribe($$self, error, value => $$invalidate(0, $error = value));
	version.get();
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Settings_server_information> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Settings_server_information", $$slots, []);

	$$self.$capture_state = () => ({
		version,
		loading,
		error,
		$error,
		$loading,
		$version
	});

	return [$error, $loading, $version, loading, error];
}

class Settings_server_information extends SvelteComponentDev {
	constructor(options) {
		super(options);
		init(this, options, instance$y, create_fragment$y, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "Settings_server_information",
			options,
			id: create_fragment$y.name
		});
	}
}

// Outside of svelte, auto subscribing doesnt work.
let lh;
const unsubscribe = loginHelper.subscribe(value => {
  lh = value;
});

function logout() {
  loginHelper.logout();
  return true;
}

function checkIfLoggedIn(detail) {
  if (!lh.loggedIn) {
    loginHelper.redirectURLAfterLogin(detail.location);
    return false;
  }
  return true;
}

let routes = {
  '/': Home,
  '/login': Login_1,
  '/logout': wrap(
    Logout,
    logout),
  '/create': wrap(
    Create,
    checkIfLoggedIn),
  '/create/list': wrap(
    Create_list,
    checkIfLoggedIn),
  '/create/label': wrap(
    Create_label,
    checkIfLoggedIn),
  '/list/edit/:uuid': wrap(
    List_edit$1,
    checkIfLoggedIn),
  '/list/view/:uuid': wrap(
    List_view$1,
    checkIfLoggedIn),
  '/list/deleted': wrap(
    List_deleted,
    checkIfLoggedIn),
  '/lists/by/me': wrap(
    List_find,
    checkIfLoggedIn),
  '/settings/server_information': Settings_server_information,
  // Catch-all, must be last
  '*': Not_found,
};

/* src/editor/App.svelte generated by Svelte v3.20.1 */

const file$y = "src/editor/App.svelte";

function add_css$9() {
	var style = element("style");
	style.id = "svelte-64n3a6-style";
	style.textContent = "h1{font-size:2em;margin:.67em 0}a{background-color:transparent}b{font-weight:bolder}small{font-size:80%}button,input,select,textarea{font-family:inherit;font-size:100%;line-height:1.15;margin:0}button,input{overflow:visible}button,select{text-transform:none}button{-webkit-appearance:button}button::-moz-focus-inner{border-style:none;padding:0}button:-moz-focusring{outline:1px dotted ButtonText}fieldset{padding:.35em .75em .625em}textarea{overflow:auto}a,article,blockquote,div,dt,fieldset,form,h1,h2,h3,header,li,main,p,section,textarea,ul{box-sizing:border-box}.ba{border-style:solid;border-width:1px}.bt{border-top-style:solid;border-top-width:1px}.br{border-right-style:solid;border-right-width:1px}.bb{border-bottom-style:solid;border-bottom-width:1px}.bl{border-left-style:solid;border-left-width:1px}.bn{border-style:none;border-width:0}.b--black{border-color:#000}.b--moon-gray{border-color:#ccc}.b--black-30{border-color:rgba(0,0,0,.3)}.b--black-20{border-color:rgba(0,0,0,.2)}.b--black-10{border-color:rgba(0,0,0,.1)}.b--black-05{border-color:rgba(0,0,0,.05)}.b--red{border-color:#ff4136}.b--yellow{border-color:gold}.b--washed-yellow{border-color:#fffceb}.b--transparent{border-color:transparent}.br1{border-radius:.125rem}.br2{border-radius:.25rem}.br3{border-radius:.5rem}.b--dotted{border-style:dotted}.bw1{border-width:.125rem}.bw2{border-width:.25rem}.bw3{border-width:.5rem}.bt-0{border-top-width:0}.br-0{border-right-width:0}.bl-0{border-left-width:0}.di{display:inline}.db{display:block}.dib{display:inline-block}.dt{display:table}.dtc{display:table-cell}.flex{display:flex}.flex-column{flex-direction:column}.items-end{align-items:flex-end}.items-center{align-items:center}.justify-center{justify-content:center}.fl{float:left}.fl,.fr{_display:inline}.fr{float:right}.athelas{font-family:athelas,georgia,serif}.fs-normal{font-style:normal}.b{font-weight:700}.fw3{font-weight:300}.fw4{font-weight:400}.fw5{font-weight:500}.fw6{font-weight:600}.fw9{font-weight:900}.input-reset{-webkit-appearance:none;-moz-appearance:none}.input-reset::-moz-focus-inner{border:0;padding:0}.h1{height:1rem}.h2{height:2rem}.h3{height:4rem}.tracked{letter-spacing:.1em}.lh-title{line-height:1.25}.lh-copy{line-height:1.5}.link{text-decoration:none}.link,.link:active,.link:focus,.link:hover,.link:link,.link:visited{transition:color .15s ease-in}.link:focus{outline:1px dotted currentColor}.list{list-style-type:none}.mw-100{max-width:100%}.w1{width:1rem}.w-25{width:25%}.w-75{width:75%}.w-100{width:100%}.black-90{color:rgba(0,0,0,.9)}.black-80{color:rgba(0,0,0,.8)}.black-70{color:rgba(0,0,0,.7)}.black-60{color:rgba(0,0,0,.6)}.black-40{color:rgba(0,0,0,.4)}.black-20{color:rgba(0,0,0,.2)}.black{color:#000}.dark-gray{color:#333}.white{color:#fff}.navy{color:#001b44}.bg-white{background-color:#fff}.bg-transparent{background-color:transparent}.bg-light-red{background-color:#ff725c}.bg-washed-yellow{background-color:#fffceb}.bg-washed-red{background-color:#ffdfdf}.hover-red:focus,.hover-red:hover{color:#ff4136}.hover-blue:focus,.hover-blue:hover{color:#357edd}.pa0{padding:0}.pa1{padding:.25rem}.pa2{padding:.5rem}.pa3{padding:1rem}.pa4{padding:2rem}.pl0{padding-left:0}.pl4{padding-left:2rem}.pr2{padding-right:.5rem}.pb2{padding-bottom:.5rem}.pt2{padding-top:.5rem}.pt5{padding-top:4rem}.pv0{padding-top:0;padding-bottom:0}.pv1{padding-top:.25rem;padding-bottom:.25rem}.pv2{padding-top:.5rem;padding-bottom:.5rem}.pv3{padding-top:1rem;padding-bottom:1rem}.pv5{padding-top:4rem;padding-bottom:4rem}.ph0{padding-left:0;padding-right:0}.ph1{padding-left:.25rem;padding-right:.25rem}.ph3{padding-left:1rem;padding-right:1rem}.ph4{padding-left:2rem;padding-right:2rem}.ml0{margin-left:0}.ml3{margin-left:1rem}.mr1{margin-right:.25rem}.mr2{margin-right:.5rem}.mb0{margin-bottom:0}.mb1{margin-bottom:.25rem}.mb2{margin-bottom:.5rem}.mb3{margin-bottom:1rem}.mb5{margin-bottom:4rem}.mt0{margin-top:0}.mt2{margin-top:.5rem}.mt3{margin-top:1rem}.mt4{margin-top:2rem}.mv0{margin-top:0;margin-bottom:0}.mv2{margin-top:.5rem;margin-bottom:.5rem}.mv3{margin-top:1rem;margin-bottom:1rem}.mh0{margin-left:0;margin-right:0}.mh1{margin-left:.25rem;margin-right:.25rem}.underline{text-decoration:underline}.tc{text-align:center}.ttu{text-transform:uppercase}.f2{font-size:2.25rem}.f3{font-size:1.5rem}.f4{font-size:1.25rem}.f5{font-size:1rem}.f6{font-size:.875rem}.measure{max-width:30em}.center{margin-left:auto}.center{margin-right:auto}.nowrap{white-space:nowrap}.v-mid{vertical-align:middle}.dim{opacity:1}.dim,.dim:focus,.dim:hover{transition:opacity .15s ease-in}.dim:focus,.dim:hover{opacity:.5}.dim:active{opacity:.8;transition:opacity .15s ease-out}@media screen and (min-width:30em){.pa2-ns{padding:.5rem}.pa5-ns{padding:4rem}.ph1-ns{padding-left:.25rem;padding-right:.25rem}.ph5-ns{padding-left:4rem;padding-right:4rem}.mr6-ns{margin-right:8rem}.mt3-ns{margin-top:1rem}.f2-ns{font-size:2.25rem}.f4-ns{font-size:1.25rem}.f5-ns{font-size:1rem}}@media screen and (min-width:30em) and (max-width:60em){.mr3-m{margin-right:1rem}.f4-m{font-size:1.25rem}}@media screen and (min-width:60em){.ph4-l{padding-left:2rem;padding-right:2rem}.mr2-l{margin-right:.5rem}.mr4-l{margin-right:2rem}.mr5-l{margin-right:4rem}.f3-l{font-size:1.5rem}}\n/*# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiQXBwLnN2ZWx0ZSIsInNvdXJjZXMiOlsiQXBwLnN2ZWx0ZSJdLCJzb3VyY2VzQ29udGVudCI6WyI8c2NyaXB0PlxuICBpbXBvcnQgY2FjaGUgZnJvbSBcIi4vbGliL2NhY2hlLmpzXCI7XG4gIGltcG9ydCBSb3V0ZXIgZnJvbSBcInN2ZWx0ZS1zcGEtcm91dGVyXCI7XG4gIGltcG9ydCBUb3BNZW51IGZyb20gXCIuL2NvbXBvbmVudHMvbWVudV90b3Auc3ZlbHRlXCI7XG4gIGltcG9ydCBGb290ZXIgZnJvbSBcIi4vY29tcG9uZW50cy9mb290ZXIuc3ZlbHRlXCI7XG4gIGltcG9ydCBNZW51IGZyb20gXCIuL2NvbXBvbmVudHMvbWVudS5zdmVsdGVcIjtcbiAgaW1wb3J0IEJveCBmcm9tIFwiLi9jb21wb25lbnRzL0JveC5zdmVsdGVcIjtcblxuICAvLyBJbXBvcnQgdGhlIFwibGlua1wiIGFjdGlvbiBhbmQgdGhlIG1ldGhvZHMgdG8gY29udHJvbCBoaXN0b3J5IHByb2dyYW1tYXRpY2FsbHkgZnJvbSB0aGUgc2FtZSBtb2R1bGUsIGFzIHdlbGwgYXMgdGhlIGxvY2F0aW9uIHN0b3JlXG4gIGltcG9ydCB7XG4gICAgbGluayxcbiAgICBwdXNoLFxuICAgIHBvcCxcbiAgICByZXBsYWNlLFxuICAgIGxvY2F0aW9uLFxuICAgIHF1ZXJ5c3RyaW5nXG4gIH0gZnJvbSBcInN2ZWx0ZS1zcGEtcm91dGVyXCI7XG4gIC8vIEltcG9ydCB0aGUgXCJhY3RpdmVcIiBhY3Rpb25cbiAgLy8gTm9ybWFsbHksIHRoaXMgd291bGQgYmUgaW1wb3J0OiBgaW1wb3J0IGFjdGl2ZSBmcm9tICdzdmVsdGUtc3BhLXJvdXRlci9hY3RpdmUnYFxuICBpbXBvcnQgYWN0aXZlIGZyb20gXCJzdmVsdGUtc3BhLXJvdXRlci9hY3RpdmVcIjtcblxuICAvLyBJbXBvcnQgdGhlIGxpc3Qgb2Ygcm91dGVzXG4gIGltcG9ydCByb3V0ZXMgZnJvbSBcIi4vcm91dGVzLmpzXCI7XG5cbiAgLy8gQ29udGFpbnMgbG9nZ2luZyBpbmZvcm1hdGlvbiB1c2VkIGJ5IHRlc3RzXG4gIC8vbGV0IGxvZ2JveCA9IFwiXCI7XG5cbiAgLy8gSGFuZGxlcyB0aGUgXCJjb25kaXRpb25zRmFpbGVkXCIgZXZlbnQgZGlzcGF0Y2hlZCBieSB0aGUgcm91dGVyIHdoZW4gYSBjb21wb25lbnQgY2FuJ3QgYmUgbG9hZGVkIGJlY2F1c2Ugb25lIG9mIGl0cyBwcmUtY29uZGl0aW9uIGZhaWxlZFxuICBmdW5jdGlvbiBjb25kaXRpb25zRmFpbGVkKGV2ZW50KSB7XG4gICAgLy8gZXNsaW50LWRpc2FibGUtbmV4dC1saW5lIG5vLWNvbnNvbGVcbiAgICAvLyBjb25zb2xlLmVycm9yKCdDYXVnaHQgZXZlbnQgY29uZGl0aW9uc0ZhaWxlZCcsIGV2ZW50LmRldGFpbClcbiAgICAvLyBsb2dib3ggKz0gJ2NvbmRpdGlvbnNGYWlsZWQgLSAnICsgSlNPTi5zdHJpbmdpZnkoZXZlbnQuZGV0YWlsKSArICdcXG4nXG5cbiAgICAvLyBSZXBsYWNlIHRoZSByb3V0ZVxuICAgIHJlcGxhY2UoXCIvbG9naW5cIik7XG4gIH1cblxuICAvLyBIYW5kbGVzIHRoZSBcInJvdXRlTG9hZGVkXCIgZXZlbnQgZGlzcGF0Y2hlZCBieSB0aGUgcm91dGVyIGFmdGVyIGEgcm91dGUgaGFzIGJlZW4gc3VjY2Vzc2Z1bGx5IGxvYWRlZFxuICBmdW5jdGlvbiByb3V0ZUxvYWRlZChldmVudCkge1xuICAgIC8vIGVzbGludC1kaXNhYmxlLW5leHQtbGluZSBuby1jb25zb2xlXG4gICAgLy8gY29uc29sZS5pbmZvKCdDYXVnaHQgZXZlbnQgcm91dGVMb2FkZWQnLCBldmVudC5kZXRhaWwpXG4gICAgLy8gbG9nYm94ICs9ICdyb3V0ZUxvYWRlZCAtICcgKyBKU09OLnN0cmluZ2lmeShldmVudC5kZXRhaWwpICsgJ1xcbidcbiAgICBjYWNoZS5zYXZlKGNhY2hlLmtleXNbXCJsYXN0LnNjcmVlblwiXSwgXCIjXCIgKyBldmVudC5kZXRhaWwubG9jYXRpb24pO1xuICB9XG48L3NjcmlwdD5cblxuPHN0eWxlIGdsb2JhbD4vKiEgVEFDSFlPTlMgdjQuMTEuMSB8IGh0dHA6Ly90YWNoeW9ucy5pbyAqL1xuLyohIG5vcm1hbGl6ZS5jc3MgdjguMC4wIHwgTUlUIExpY2Vuc2UgfCBnaXRodWIuY29tL25lY29sYXMvbm9ybWFsaXplLmNzcyAqL1xuOmdsb2JhbChoMSl7Zm9udC1zaXplOjJlbTttYXJnaW46LjY3ZW0gMH1cbjpnbG9iYWwoYSl7YmFja2dyb3VuZC1jb2xvcjp0cmFuc3BhcmVudH1cbjpnbG9iYWwoYil7Zm9udC13ZWlnaHQ6Ym9sZGVyfVxuOmdsb2JhbChzbWFsbCl7Zm9udC1zaXplOjgwJX1cbjpnbG9iYWwoYnV0dG9uKSw6Z2xvYmFsKGlucHV0KSw6Z2xvYmFsKHNlbGVjdCksOmdsb2JhbCh0ZXh0YXJlYSl7Zm9udC1mYW1pbHk6aW5oZXJpdDtmb250LXNpemU6MTAwJTtsaW5lLWhlaWdodDoxLjE1O21hcmdpbjowfVxuOmdsb2JhbChidXR0b24pLDpnbG9iYWwoaW5wdXQpe292ZXJmbG93OnZpc2libGV9XG46Z2xvYmFsKGJ1dHRvbiksOmdsb2JhbChzZWxlY3Qpe3RleHQtdHJhbnNmb3JtOm5vbmV9XG46Z2xvYmFsKGJ1dHRvbil7LXdlYmtpdC1hcHBlYXJhbmNlOmJ1dHRvbn1cbjpnbG9iYWwoYnV0dG9uOjotbW96LWZvY3VzLWlubmVyKXtib3JkZXItc3R5bGU6bm9uZTtwYWRkaW5nOjB9XG46Z2xvYmFsKGJ1dHRvbjotbW96LWZvY3VzcmluZyl7b3V0bGluZToxcHggZG90dGVkIEJ1dHRvblRleHR9XG46Z2xvYmFsKGZpZWxkc2V0KXtwYWRkaW5nOi4zNWVtIC43NWVtIC42MjVlbX1cbjpnbG9iYWwodGV4dGFyZWEpe292ZXJmbG93OmF1dG99XG46Z2xvYmFsKGEpLDpnbG9iYWwoYXJ0aWNsZSksOmdsb2JhbChibG9ja3F1b3RlKSw6Z2xvYmFsKGRpdiksOmdsb2JhbChkdCksOmdsb2JhbChmaWVsZHNldCksOmdsb2JhbChmb3JtKSw6Z2xvYmFsKGgxKSw6Z2xvYmFsKGgyKSw6Z2xvYmFsKGgzKSw6Z2xvYmFsKGhlYWRlciksOmdsb2JhbChsaSksOmdsb2JhbChtYWluKSw6Z2xvYmFsKHApLDpnbG9iYWwoc2VjdGlvbiksOmdsb2JhbCh0ZXh0YXJlYSksOmdsb2JhbCh1bCl7Ym94LXNpemluZzpib3JkZXItYm94fVxuOmdsb2JhbCguYmEpe2JvcmRlci1zdHlsZTpzb2xpZDtib3JkZXItd2lkdGg6MXB4fVxuOmdsb2JhbCguYnQpe2JvcmRlci10b3Atc3R5bGU6c29saWQ7Ym9yZGVyLXRvcC13aWR0aDoxcHh9XG46Z2xvYmFsKC5icil7Ym9yZGVyLXJpZ2h0LXN0eWxlOnNvbGlkO2JvcmRlci1yaWdodC13aWR0aDoxcHh9XG46Z2xvYmFsKC5iYil7Ym9yZGVyLWJvdHRvbS1zdHlsZTpzb2xpZDtib3JkZXItYm90dG9tLXdpZHRoOjFweH1cbjpnbG9iYWwoLmJsKXtib3JkZXItbGVmdC1zdHlsZTpzb2xpZDtib3JkZXItbGVmdC13aWR0aDoxcHh9XG46Z2xvYmFsKC5ibil7Ym9yZGVyLXN0eWxlOm5vbmU7Ym9yZGVyLXdpZHRoOjB9XG46Z2xvYmFsKC5iLS1ibGFjayl7Ym9yZGVyLWNvbG9yOiMwMDB9XG46Z2xvYmFsKC5iLS1tb29uLWdyYXkpe2JvcmRlci1jb2xvcjojY2NjfVxuOmdsb2JhbCguYi0tYmxhY2stMzApe2JvcmRlci1jb2xvcjpyZ2JhKDAsMCwwLC4zKX1cbjpnbG9iYWwoLmItLWJsYWNrLTIwKXtib3JkZXItY29sb3I6cmdiYSgwLDAsMCwuMil9XG46Z2xvYmFsKC5iLS1ibGFjay0xMCl7Ym9yZGVyLWNvbG9yOnJnYmEoMCwwLDAsLjEpfVxuOmdsb2JhbCguYi0tYmxhY2stMDUpe2JvcmRlci1jb2xvcjpyZ2JhKDAsMCwwLC4wNSl9XG46Z2xvYmFsKC5iLS1yZWQpe2JvcmRlci1jb2xvcjojZmY0MTM2fVxuOmdsb2JhbCguYi0teWVsbG93KXtib3JkZXItY29sb3I6Z29sZH1cbjpnbG9iYWwoLmItLXdhc2hlZC15ZWxsb3cpe2JvcmRlci1jb2xvcjojZmZmY2VifVxuOmdsb2JhbCguYi0tdHJhbnNwYXJlbnQpe2JvcmRlci1jb2xvcjp0cmFuc3BhcmVudH1cbjpnbG9iYWwoLmJyMSl7Ym9yZGVyLXJhZGl1czouMTI1cmVtfVxuOmdsb2JhbCguYnIyKXtib3JkZXItcmFkaXVzOi4yNXJlbX1cbjpnbG9iYWwoLmJyMyl7Ym9yZGVyLXJhZGl1czouNXJlbX1cbjpnbG9iYWwoLmItLWRvdHRlZCl7Ym9yZGVyLXN0eWxlOmRvdHRlZH1cbjpnbG9iYWwoLmJ3MSl7Ym9yZGVyLXdpZHRoOi4xMjVyZW19XG46Z2xvYmFsKC5idzIpe2JvcmRlci13aWR0aDouMjVyZW19XG46Z2xvYmFsKC5idzMpe2JvcmRlci13aWR0aDouNXJlbX1cbjpnbG9iYWwoLmJ0LTApe2JvcmRlci10b3Atd2lkdGg6MH1cbjpnbG9iYWwoLmJyLTApe2JvcmRlci1yaWdodC13aWR0aDowfVxuOmdsb2JhbCguYmwtMCl7Ym9yZGVyLWxlZnQtd2lkdGg6MH1cbjpnbG9iYWwoLmRpKXtkaXNwbGF5OmlubGluZX1cbjpnbG9iYWwoLmRiKXtkaXNwbGF5OmJsb2NrfVxuOmdsb2JhbCguZGliKXtkaXNwbGF5OmlubGluZS1ibG9ja31cbjpnbG9iYWwoLmR0KXtkaXNwbGF5OnRhYmxlfVxuOmdsb2JhbCguZHRjKXtkaXNwbGF5OnRhYmxlLWNlbGx9XG46Z2xvYmFsKC5mbGV4KXtkaXNwbGF5OmZsZXh9XG46Z2xvYmFsKC5mbGV4LWNvbHVtbil7ZmxleC1kaXJlY3Rpb246Y29sdW1ufVxuOmdsb2JhbCguaXRlbXMtZW5kKXthbGlnbi1pdGVtczpmbGV4LWVuZH1cbjpnbG9iYWwoLml0ZW1zLWNlbnRlcil7YWxpZ24taXRlbXM6Y2VudGVyfVxuOmdsb2JhbCguanVzdGlmeS1jZW50ZXIpe2p1c3RpZnktY29udGVudDpjZW50ZXJ9XG46Z2xvYmFsKC5mbCl7ZmxvYXQ6bGVmdH1cbjpnbG9iYWwoLmZsKSw6Z2xvYmFsKC5mcil7X2Rpc3BsYXk6aW5saW5lfVxuOmdsb2JhbCguZnIpe2Zsb2F0OnJpZ2h0fVxuOmdsb2JhbCguYXRoZWxhcyl7Zm9udC1mYW1pbHk6YXRoZWxhcyxnZW9yZ2lhLHNlcmlmfVxuOmdsb2JhbCguZnMtbm9ybWFsKXtmb250LXN0eWxlOm5vcm1hbH1cbjpnbG9iYWwoLmIpe2ZvbnQtd2VpZ2h0OjcwMH1cbjpnbG9iYWwoLmZ3Myl7Zm9udC13ZWlnaHQ6MzAwfVxuOmdsb2JhbCguZnc0KXtmb250LXdlaWdodDo0MDB9XG46Z2xvYmFsKC5mdzUpe2ZvbnQtd2VpZ2h0OjUwMH1cbjpnbG9iYWwoLmZ3Nil7Zm9udC13ZWlnaHQ6NjAwfVxuOmdsb2JhbCguZnc5KXtmb250LXdlaWdodDo5MDB9XG46Z2xvYmFsKC5pbnB1dC1yZXNldCl7LXdlYmtpdC1hcHBlYXJhbmNlOm5vbmU7LW1vei1hcHBlYXJhbmNlOm5vbmV9XG46Z2xvYmFsKC5pbnB1dC1yZXNldDo6LW1vei1mb2N1cy1pbm5lcil7Ym9yZGVyOjA7cGFkZGluZzowfVxuOmdsb2JhbCguaDEpe2hlaWdodDoxcmVtfVxuOmdsb2JhbCguaDIpe2hlaWdodDoycmVtfVxuOmdsb2JhbCguaDMpe2hlaWdodDo0cmVtfVxuOmdsb2JhbCgudHJhY2tlZCl7bGV0dGVyLXNwYWNpbmc6LjFlbX1cbjpnbG9iYWwoLmxoLXRpdGxlKXtsaW5lLWhlaWdodDoxLjI1fVxuOmdsb2JhbCgubGgtY29weSl7bGluZS1oZWlnaHQ6MS41fVxuOmdsb2JhbCgubGluayl7dGV4dC1kZWNvcmF0aW9uOm5vbmV9XG46Z2xvYmFsKC5saW5rKSw6Z2xvYmFsKC5saW5rOmFjdGl2ZSksOmdsb2JhbCgubGluazpmb2N1cyksOmdsb2JhbCgubGluazpob3ZlciksOmdsb2JhbCgubGluazpsaW5rKSw6Z2xvYmFsKC5saW5rOnZpc2l0ZWQpe3RyYW5zaXRpb246Y29sb3IgLjE1cyBlYXNlLWlufVxuOmdsb2JhbCgubGluazpmb2N1cyl7b3V0bGluZToxcHggZG90dGVkIGN1cnJlbnRDb2xvcn1cbjpnbG9iYWwoLmxpc3Qpe2xpc3Qtc3R5bGUtdHlwZTpub25lfVxuOmdsb2JhbCgubXctMTAwKXttYXgtd2lkdGg6MTAwJX1cbjpnbG9iYWwoLncxKXt3aWR0aDoxcmVtfVxuOmdsb2JhbCgudy0yNSl7d2lkdGg6MjUlfVxuOmdsb2JhbCgudy03NSl7d2lkdGg6NzUlfVxuOmdsb2JhbCgudy0xMDApe3dpZHRoOjEwMCV9XG46Z2xvYmFsKC5ibGFjay05MCl7Y29sb3I6cmdiYSgwLDAsMCwuOSl9XG46Z2xvYmFsKC5ibGFjay04MCl7Y29sb3I6cmdiYSgwLDAsMCwuOCl9XG46Z2xvYmFsKC5ibGFjay03MCl7Y29sb3I6cmdiYSgwLDAsMCwuNyl9XG46Z2xvYmFsKC5ibGFjay02MCl7Y29sb3I6cmdiYSgwLDAsMCwuNil9XG46Z2xvYmFsKC5ibGFjay00MCl7Y29sb3I6cmdiYSgwLDAsMCwuNCl9XG46Z2xvYmFsKC5ibGFjay0yMCl7Y29sb3I6cmdiYSgwLDAsMCwuMil9XG46Z2xvYmFsKC5ibGFjayl7Y29sb3I6IzAwMH1cbjpnbG9iYWwoLmRhcmstZ3JheSl7Y29sb3I6IzMzM31cbjpnbG9iYWwoLndoaXRlKXtjb2xvcjojZmZmfVxuOmdsb2JhbCgubmF2eSl7Y29sb3I6IzAwMWI0NH1cbjpnbG9iYWwoLmJnLXdoaXRlKXtiYWNrZ3JvdW5kLWNvbG9yOiNmZmZ9XG46Z2xvYmFsKC5iZy10cmFuc3BhcmVudCl7YmFja2dyb3VuZC1jb2xvcjp0cmFuc3BhcmVudH1cbjpnbG9iYWwoLmJnLWxpZ2h0LXJlZCl7YmFja2dyb3VuZC1jb2xvcjojZmY3MjVjfVxuOmdsb2JhbCguYmctd2FzaGVkLXllbGxvdyl7YmFja2dyb3VuZC1jb2xvcjojZmZmY2VifVxuOmdsb2JhbCguYmctd2FzaGVkLXJlZCl7YmFja2dyb3VuZC1jb2xvcjojZmZkZmRmfVxuOmdsb2JhbCguaG92ZXItcmVkOmZvY3VzKSw6Z2xvYmFsKC5ob3Zlci1yZWQ6aG92ZXIpe2NvbG9yOiNmZjQxMzZ9XG46Z2xvYmFsKC5ob3Zlci1ibHVlOmZvY3VzKSw6Z2xvYmFsKC5ob3Zlci1ibHVlOmhvdmVyKXtjb2xvcjojMzU3ZWRkfVxuOmdsb2JhbCgucGEwKXtwYWRkaW5nOjB9XG46Z2xvYmFsKC5wYTEpe3BhZGRpbmc6LjI1cmVtfVxuOmdsb2JhbCgucGEyKXtwYWRkaW5nOi41cmVtfVxuOmdsb2JhbCgucGEzKXtwYWRkaW5nOjFyZW19XG46Z2xvYmFsKC5wYTQpe3BhZGRpbmc6MnJlbX1cbjpnbG9iYWwoLnBsMCl7cGFkZGluZy1sZWZ0OjB9XG46Z2xvYmFsKC5wbDQpe3BhZGRpbmctbGVmdDoycmVtfVxuOmdsb2JhbCgucHIyKXtwYWRkaW5nLXJpZ2h0Oi41cmVtfVxuOmdsb2JhbCgucGIyKXtwYWRkaW5nLWJvdHRvbTouNXJlbX1cbjpnbG9iYWwoLnB0Mil7cGFkZGluZy10b3A6LjVyZW19XG46Z2xvYmFsKC5wdDUpe3BhZGRpbmctdG9wOjRyZW19XG46Z2xvYmFsKC5wdjApe3BhZGRpbmctdG9wOjA7cGFkZGluZy1ib3R0b206MH1cbjpnbG9iYWwoLnB2MSl7cGFkZGluZy10b3A6LjI1cmVtO3BhZGRpbmctYm90dG9tOi4yNXJlbX1cbjpnbG9iYWwoLnB2Mil7cGFkZGluZy10b3A6LjVyZW07cGFkZGluZy1ib3R0b206LjVyZW19XG46Z2xvYmFsKC5wdjMpe3BhZGRpbmctdG9wOjFyZW07cGFkZGluZy1ib3R0b206MXJlbX1cbjpnbG9iYWwoLnB2NSl7cGFkZGluZy10b3A6NHJlbTtwYWRkaW5nLWJvdHRvbTo0cmVtfVxuOmdsb2JhbCgucGgwKXtwYWRkaW5nLWxlZnQ6MDtwYWRkaW5nLXJpZ2h0OjB9XG46Z2xvYmFsKC5waDEpe3BhZGRpbmctbGVmdDouMjVyZW07cGFkZGluZy1yaWdodDouMjVyZW19XG46Z2xvYmFsKC5waDMpe3BhZGRpbmctbGVmdDoxcmVtO3BhZGRpbmctcmlnaHQ6MXJlbX1cbjpnbG9iYWwoLnBoNCl7cGFkZGluZy1sZWZ0OjJyZW07cGFkZGluZy1yaWdodDoycmVtfVxuOmdsb2JhbCgubWwwKXttYXJnaW4tbGVmdDowfVxuOmdsb2JhbCgubWwzKXttYXJnaW4tbGVmdDoxcmVtfVxuOmdsb2JhbCgubXIxKXttYXJnaW4tcmlnaHQ6LjI1cmVtfVxuOmdsb2JhbCgubXIyKXttYXJnaW4tcmlnaHQ6LjVyZW19XG46Z2xvYmFsKC5tYjApe21hcmdpbi1ib3R0b206MH1cbjpnbG9iYWwoLm1iMSl7bWFyZ2luLWJvdHRvbTouMjVyZW19XG46Z2xvYmFsKC5tYjIpe21hcmdpbi1ib3R0b206LjVyZW19XG46Z2xvYmFsKC5tYjMpe21hcmdpbi1ib3R0b206MXJlbX1cbjpnbG9iYWwoLm1iNSl7bWFyZ2luLWJvdHRvbTo0cmVtfVxuOmdsb2JhbCgubXQwKXttYXJnaW4tdG9wOjB9XG46Z2xvYmFsKC5tdDIpe21hcmdpbi10b3A6LjVyZW19XG46Z2xvYmFsKC5tdDMpe21hcmdpbi10b3A6MXJlbX1cbjpnbG9iYWwoLm10NCl7bWFyZ2luLXRvcDoycmVtfVxuOmdsb2JhbCgubXYwKXttYXJnaW4tdG9wOjA7bWFyZ2luLWJvdHRvbTowfVxuOmdsb2JhbCgubXYyKXttYXJnaW4tdG9wOi41cmVtO21hcmdpbi1ib3R0b206LjVyZW19XG46Z2xvYmFsKC5tdjMpe21hcmdpbi10b3A6MXJlbTttYXJnaW4tYm90dG9tOjFyZW19XG46Z2xvYmFsKC5taDApe21hcmdpbi1sZWZ0OjA7bWFyZ2luLXJpZ2h0OjB9XG46Z2xvYmFsKC5taDEpe21hcmdpbi1sZWZ0Oi4yNXJlbTttYXJnaW4tcmlnaHQ6LjI1cmVtfVxuOmdsb2JhbCgudW5kZXJsaW5lKXt0ZXh0LWRlY29yYXRpb246dW5kZXJsaW5lfVxuOmdsb2JhbCgudGMpe3RleHQtYWxpZ246Y2VudGVyfVxuOmdsb2JhbCgudHR1KXt0ZXh0LXRyYW5zZm9ybTp1cHBlcmNhc2V9XG46Z2xvYmFsKC5mMil7Zm9udC1zaXplOjIuMjVyZW19XG46Z2xvYmFsKC5mMyl7Zm9udC1zaXplOjEuNXJlbX1cbjpnbG9iYWwoLmY0KXtmb250LXNpemU6MS4yNXJlbX1cbjpnbG9iYWwoLmY1KXtmb250LXNpemU6MXJlbX1cbjpnbG9iYWwoLmY2KXtmb250LXNpemU6Ljg3NXJlbX1cbjpnbG9iYWwoLm1lYXN1cmUpe21heC13aWR0aDozMGVtfVxuOmdsb2JhbCguY2VudGVyKXttYXJnaW4tbGVmdDphdXRvfVxuOmdsb2JhbCguY2VudGVyKXttYXJnaW4tcmlnaHQ6YXV0b31cbjpnbG9iYWwoLm5vd3JhcCl7d2hpdGUtc3BhY2U6bm93cmFwfVxuOmdsb2JhbCgudi1taWQpe3ZlcnRpY2FsLWFsaWduOm1pZGRsZX1cbjpnbG9iYWwoLmRpbSl7b3BhY2l0eToxfVxuOmdsb2JhbCguZGltKSw6Z2xvYmFsKC5kaW06Zm9jdXMpLDpnbG9iYWwoLmRpbTpob3Zlcil7dHJhbnNpdGlvbjpvcGFjaXR5IC4xNXMgZWFzZS1pbn1cbjpnbG9iYWwoLmRpbTpmb2N1cyksOmdsb2JhbCguZGltOmhvdmVyKXtvcGFjaXR5Oi41fVxuOmdsb2JhbCguZGltOmFjdGl2ZSl7b3BhY2l0eTouODt0cmFuc2l0aW9uOm9wYWNpdHkgLjE1cyBlYXNlLW91dH1cbkBtZWRpYSBzY3JlZW4gYW5kIChtaW4td2lkdGg6MzBlbSl7Omdsb2JhbCgucGEyLW5zKXtwYWRkaW5nOi41cmVtfTpnbG9iYWwoLnBhNS1ucyl7cGFkZGluZzo0cmVtfTpnbG9iYWwoLnBoMS1ucyl7cGFkZGluZy1sZWZ0Oi4yNXJlbTtwYWRkaW5nLXJpZ2h0Oi4yNXJlbX06Z2xvYmFsKC5waDUtbnMpe3BhZGRpbmctbGVmdDo0cmVtO3BhZGRpbmctcmlnaHQ6NHJlbX06Z2xvYmFsKC5tcjYtbnMpe21hcmdpbi1yaWdodDo4cmVtfTpnbG9iYWwoLm10My1ucyl7bWFyZ2luLXRvcDoxcmVtfTpnbG9iYWwoLmYyLW5zKXtmb250LXNpemU6Mi4yNXJlbX06Z2xvYmFsKC5mNC1ucyl7Zm9udC1zaXplOjEuMjVyZW19Omdsb2JhbCguZjUtbnMpe2ZvbnQtc2l6ZToxcmVtfX1cbkBtZWRpYSBzY3JlZW4gYW5kIChtaW4td2lkdGg6MzBlbSkgYW5kIChtYXgtd2lkdGg6NjBlbSl7Omdsb2JhbCgubXIzLW0pe21hcmdpbi1yaWdodDoxcmVtfTpnbG9iYWwoLmY0LW0pe2ZvbnQtc2l6ZToxLjI1cmVtfX1cbkBtZWRpYSBzY3JlZW4gYW5kIChtaW4td2lkdGg6NjBlbSl7Omdsb2JhbCgucGg0LWwpe3BhZGRpbmctbGVmdDoycmVtO3BhZGRpbmctcmlnaHQ6MnJlbX06Z2xvYmFsKC5tcjItbCl7bWFyZ2luLXJpZ2h0Oi41cmVtfTpnbG9iYWwoLm1yNC1sKXttYXJnaW4tcmlnaHQ6MnJlbX06Z2xvYmFsKC5tcjUtbCl7bWFyZ2luLXJpZ2h0OjRyZW19Omdsb2JhbCguZjMtbCl7Zm9udC1zaXplOjEuNXJlbX19XG4vKiMgc291cmNlTWFwcGluZ1VSTD1kYXRhOmFwcGxpY2F0aW9uL2pzb247YmFzZTY0LGV5SjJaWEp6YVc5dUlqb3pMQ0p6YjNWeVkyVnpJanBiSW5OeVl5OWxaR2wwYjNJdmJtOWtaVjl0YjJSMWJHVnpMM1JoWTJoNWIyNXpMMk56Y3k5MFlXTm9lVzl1Y3k1dGFXNHVZM056SWwwc0ltNWhiV1Z6SWpwYlhTd2liV0Z3Y0dsdVozTWlPaUpCUVVGQkxESkRRVUV5UXp0QlFVTXpReXd5UlVGQk1rVTdRVUZCYlVVc1dVRkJSeXhoUVVGaExFTkJRVU1zWTBGQll6dEJRVUYxUnl4WFFVRkZMRFJDUVVFMFFqdEJRVUZ2U1N4WFFVRlRMR3RDUVVGclFqdEJRVUUyUkN4bFFVRk5MR0ZCUVdFN1FVRkJjVWtzYVVWQlFYTkRMRzFDUVVGdFFpeERRVUZETEdOQlFXTXNRMEZCUXl4blFrRkJaMElzUTBGQlF5eFJRVUZSTzBGQlFVTXNLMEpCUVdFc1owSkJRV2RDTzBGQlFVTXNaME5CUVdNc2JVSkJRVzFDTzBGQlFVTXNaMEpCUVdkRUxIbENRVUY1UWp0QlFVRkRMR3REUVVGM1NDeHBRa0ZCYVVJc1EwRkJReXhUUVVGVE8wRkJRVU1zSzBKQlFUUkhMRFpDUVVFMlFqdEJRVUZETEd0Q1FVRlRMREJDUVVFd1FqdEJRVUYxU1N4clFrRkJVeXhoUVVGaE8wRkJRU3RoTEdsUVFVRTJVeXh4UWtGQmNVSTdRVUZCTkdwRExHRkJRVWtzYTBKQlFXdENMRU5CUVVNc1owSkJRV2RDTzBGQlFVTXNZVUZCU1N4elFrRkJjMElzUTBGQlF5eHZRa0ZCYjBJN1FVRkJReXhoUVVGSkxIZENRVUYzUWl4RFFVRkRMSE5DUVVGelFqdEJRVUZETEdGQlFVa3NlVUpCUVhsQ0xFTkJRVU1zZFVKQlFYVkNPMEZCUVVNc1lVRkJTU3gxUWtGQmRVSXNRMEZCUXl4eFFrRkJjVUk3UVVGQlF5eGhRVUZKTEdsQ1FVRnBRaXhEUVVGRExHTkJRV003UVVGQlF5eHRRa0ZCVlN4cFFrRkJhVUk3UVVGQk5Fd3NkVUpCUVdNc2FVSkJRV2xDTzBGQlFUWXpRaXh6UWtGQllTd3lRa0ZCTWtJN1FVRkJReXh6UWtGQllTd3lRa0ZCTWtJN1FVRkJReXh6UWtGQllTd3lRa0ZCTWtJN1FVRkJReXh6UWtGQllTdzBRa0ZCTkVJN1FVRkJOa2dzYVVKQlFWRXNiMEpCUVc5Q08wRkJRV3RITEc5Q1FVRlhMR2xDUVVGcFFqdEJRVUZwYkVJc01rSkJRV3RDTEc5Q1FVRnZRanRCUVVGeFF5eDVRa0ZCWjBJc2QwSkJRWGRDTzBGQlFYVkVMR05CUVVzc2NVSkJRWEZDTzBGQlFVTXNZMEZCU3l4dlFrRkJiMEk3UVVGQlF5eGpRVUZMTEcxQ1FVRnRRanRCUVVFMFZTeHZRa0ZCVnl4dFFrRkJiVUk3UVVGQk5FY3NZMEZCU3l4dlFrRkJiMEk3UVVGQlF5eGpRVUZMTEcxQ1FVRnRRanRCUVVGRExHTkJRVXNzYTBKQlFXdENPMEZCUVN0RExHVkJRVTBzYTBKQlFXdENPMEZCUVVNc1pVRkJUU3h2UWtGQmIwSTdRVUZCTmtJc1pVRkJUU3h0UWtGQmJVSTdRVUZCTkROQ0xHRkJRVWtzWTBGQll6dEJRVUZETEdGQlFVa3NZVUZCWVR0QlFVRkRMR05CUVVzc2IwSkJRVzlDTzBGQlFUSkNMR0ZCUVVrc1lVRkJZVHRCUVVGRExHTkJRVXNzYTBKQlFXdENPMEZCUVhOTUxHVkJRVTBzV1VGQldUdEJRVUY1Unl4elFrRkJZU3h4UWtGQmNVSTdRVUZCYjFFc2IwSkJRVmNzYjBKQlFXOUNPMEZCUVVNc2RVSkJRV01zYTBKQlFXdENPMEZCUVRSVUxIbENRVUZuUWl4elFrRkJjMEk3UVVGQk1HMUNMR0ZCUVVrc1ZVRkJWVHRCUVVGRExESkNRVUZCTEdOQlFYVkNPMEZCUVVNc1lVRkJTU3hYUVVGWE8wRkJRWEZsTEd0Q1FVRlRMR2xEUVVGcFF6dEJRVUZyVUN4dlFrRkJWeXhwUWtGQmFVSTdRVUZCZVVJc1dVRkJSeXhsUVVGbE8wRkJRVEpETEdOQlFVc3NaVUZCWlR0QlFVRkRMR05CUVVzc1pVRkJaVHRCUVVGRExHTkJRVXNzWlVGQlpUdEJRVUZETEdOQlFVc3NaVUZCWlR0QlFVRXlReXhqUVVGTExHVkJRV1U3UVVGQlF5eHpRa0ZCWVN4MVFrRkJkVUlzUTBGQlF5eHZRa0ZCYjBJN1FVRkJReXgzUTBGQkswUXNVVUZCVVN4RFFVRkRMRk5CUVZNN1FVRkJReXhoUVVGSkxGZEJRVmM3UVVGQlF5eGhRVUZKTEZkQlFWYzdRVUZCUXl4aFFVRkpMRmRCUVZjN1FVRkJORklzYTBKQlFWTXNiVUpCUVcxQ08wRkJRV2xITEcxQ1FVRlZMR2RDUVVGblFqdEJRVUZETEd0Q1FVRlRMR1ZCUVdVN1FVRkJReXhsUVVGTkxHOUNRVUZ2UWp0QlFVRkRMREJJUVVGdlJTdzJRa0ZCTmtJN1FVRkJReXh4UWtGQldTd3JRa0ZCSzBJN1FVRkJReXhsUVVGTkxHOUNRVUZ2UWp0QlFVRkRMR2xDUVVGUkxHTkJRV003UVVGQmEwNHNZVUZCU1N4VlFVRlZPMEZCUVRoR0xHVkJRVTBzVTBGQlV6dEJRVUZwU0N4bFFVRk5MRk5CUVZNN1FVRkJhVU1zWjBKQlFVOHNWVUZCVlR0QlFVRXdlRU1zYlVKQlFWVXNiMEpCUVc5Q08wRkJRVU1zYlVKQlFWVXNiMEpCUVc5Q08wRkJRVU1zYlVKQlFWVXNiMEpCUVc5Q08wRkJRVU1zYlVKQlFWVXNiMEpCUVc5Q08wRkJRV2RETEcxQ1FVRlZMRzlDUVVGdlFqdEJRVUZuUXl4dFFrRkJWU3h2UWtGQmIwSTdRVUZCTWxnc1owSkJRVThzVlVGQlZUdEJRVUYzUWl4dlFrRkJWeXhWUVVGVk8wRkJRVEJLTEdkQ1FVRlBMRlZCUVZVN1FVRkJlVmdzWlVGQlRTeGhRVUZoTzBGQlFXYzVReXh0UWtGQlZTeHhRa0ZCY1VJN1FVRkJReXg1UWtGQlowSXNORUpCUVRSQ08wRkJRWGRGTEhWQ1FVRmpMSGRDUVVGM1FqdEJRVUUwZGtJc01rSkJRV3RDTEhkQ1FVRjNRanRCUVVGRExIZENRVUZsTEhkQ1FVRjNRanRCUVVGemQwa3NiMFJCUVd0RExHRkJRV0U3UVVGQmR6VkNMSE5FUVVGdlF5eGhRVUZoTzBGQlFXdDZSU3hqUVVGTExGTkJRVk03UVVGQlF5eGpRVUZMTEdOQlFXTTdRVUZCUXl4alFVRkxMR0ZCUVdFN1FVRkJReXhqUVVGTExGbEJRVms3UVVGQlF5eGpRVUZMTEZsQlFWazdRVUZCZDBRc1kwRkJTeXhqUVVGak8wRkJRWGxGTEdOQlFVc3NhVUpCUVdsQ08wRkJRWE5JTEdOQlFVc3NiVUpCUVcxQ08wRkJRVEpMTEdOQlFVc3NiMEpCUVc5Q08wRkJRVEJMTEdOQlFVc3NhVUpCUVdsQ08wRkJRVFpETEdOQlFVc3NaMEpCUVdkQ08wRkJRVGhETEdOQlFVc3NZVUZCWVN4RFFVRkRMR2RDUVVGblFqdEJRVUZETEdOQlFVc3NhMEpCUVd0Q0xFTkJRVU1zY1VKQlFYRkNPMEZCUVVNc1kwRkJTeXhwUWtGQmFVSXNRMEZCUXl4dlFrRkJiMEk3UVVGQlF5eGpRVUZMTEdkQ1FVRm5RaXhEUVVGRExHMUNRVUZ0UWp0QlFVRXlReXhqUVVGTExHZENRVUZuUWl4RFFVRkRMRzFDUVVGdFFqdEJRVUYxUml4alFVRkxMR05CUVdNc1EwRkJReXhsUVVGbE8wRkJRVU1zWTBGQlN5eHRRa0ZCYlVJc1EwRkJReXh2UWtGQmIwSTdRVUZCTmtNc1kwRkJTeXhwUWtGQmFVSXNRMEZCUXl4clFrRkJhMEk3UVVGQlF5eGpRVUZMTEdsQ1FVRnBRaXhEUVVGRExHdENRVUZyUWp0QlFVRXdVU3hqUVVGTExHRkJRV0U3UVVGQlowUXNZMEZCU3l4blFrRkJaMEk3UVVGQk9FY3NZMEZCU3l4dFFrRkJiVUk3UVVGQlF5eGpRVUZMTEd0Q1FVRnJRanRCUVVGeFNDeGpRVUZMTEdWQlFXVTdRVUZCUXl4alFVRkxMRzlDUVVGdlFqdEJRVUZETEdOQlFVc3NiVUpCUVcxQ08wRkJRVU1zWTBGQlN5eHJRa0ZCYTBJN1FVRkJlVUlzWTBGQlN5eHJRa0ZCYTBJN1FVRkJhMFFzWTBGQlN5eFpRVUZaTzBGQlFYZENMR05CUVVzc1owSkJRV2RDTzBGQlFVTXNZMEZCU3l4bFFVRmxPMEZCUVVNc1kwRkJTeXhsUVVGbE8wRkJRV2xGTEdOQlFVc3NXVUZCV1N4RFFVRkRMR1ZCUVdVN1FVRkJOa01zWTBGQlN5eG5Ra0ZCWjBJc1EwRkJReXh0UWtGQmJVSTdRVUZCUXl4alFVRkxMR1ZCUVdVc1EwRkJReXhyUWtGQmEwSTdRVUZCYlVzc1kwRkJTeXhoUVVGaExFTkJRVU1zWTBGQll6dEJRVUZETEdOQlFVc3NhMEpCUVd0Q0xFTkJRVU1zYlVKQlFXMUNPMEZCUVRJNVF5eHZRa0ZCVnl4NVFrRkJlVUk3UVVGQk5rVXNZVUZCU1N4cFFrRkJhVUk3UVVGQmNVWXNZMEZCU3l4M1FrRkJkMEk3UVVGQlowZ3NZVUZCU1N4cFFrRkJhVUk3UVVGQlF5eGhRVUZKTEdkQ1FVRm5RanRCUVVGRExHRkJRVWtzYVVKQlFXbENPMEZCUVVNc1lVRkJTU3hqUVVGak8wRkJRVU1zWVVGQlNTeHBRa0ZCYVVJN1FVRkJjMElzYTBKQlFWTXNZMEZCWXp0QlFVRm5VU3hwUWtGQlVTeG5Ra0ZCWjBJN1FVRkJReXhwUWtGQmFVSXNhVUpCUVdsQ08wRkJRVEpMTEdsQ1FVRlJMR3RDUVVGclFqdEJRVUZ6UkN4blFrRkJUeXh4UWtGQmNVSTdRVUZCZDBRc1kwRkJTeXhUUVVGVE8wRkJRVU1zYzBSQlFUSkNMQ3RDUVVFclFqdEJRVUZETEhkRFFVRnpRaXhWUVVGVk8wRkJRVU1zY1VKQlFWa3NWVUZCVlN4RFFVRkRMR2REUVVGblF6dEJRVUZwT1Vjc2JVTkJRVFIzVUN4cFFrRkJVU3hoUVVGaExFTkJRVEpETEdsQ1FVRlJMRmxCUVZrc1EwRkJhWGhETEdsQ1FVRlJMRzFDUVVGdFFpeERRVUZETEc5Q1FVRnZRaXhEUVVFd1NTeHBRa0ZCVVN4cFFrRkJhVUlzUTBGQlF5eHJRa0ZCYTBJc1EwRkJiVzFDTEdsQ1FVRlJMR2xDUVVGcFFpeERRVUUyVkN4cFFrRkJVU3hsUVVGbExFTkJRV2R0UlN4blFrRkJUeXhwUWtGQmFVSXNRMEZCZVVJc1owSkJRVThzYVVKQlFXbENMRU5CUVVNc1owSkJRVThzWTBGQll5eERRVUZ0ZFVJN1FVRkJReXgzUkVGQmFXdFVMR2RDUVVGUExHbENRVUZwUWl4RFFVRXdPVVVzWlVGQlRTeHBRa0ZCYVVJc1EwRkJiM1ZDTzBGQlFVTXNiVU5CUVhFNVVpeG5Ra0ZCVHl4cFFrRkJhVUlzUTBGQlF5eHJRa0ZCYTBJc1EwRkJhMmhDTEdkQ1FVRlBMR3RDUVVGclFpeERRVUV3UWl4blFrRkJUeXhwUWtGQmFVSXNRMEZCUXl4blFrRkJUeXhwUWtGQmFVSXNRMEZCYVRWRkxHVkJRVTBzWjBKQlFXZENMRU5CUVRSMlFpSXNJbVpwYkdVaU9pSnpjbU12WldScGRHOXlMMEZ3Y0M1emRtVnNkR1VpTENKemIzVnlZMlZ6UTI5dWRHVnVkQ0k2V3lJdktpRWdWRUZEU0ZsUFRsTWdkalF1TVRFdU1TQjhJR2gwZEhBNkx5OTBZV05vZVc5dWN5NXBieUFxTDF4dUx5b2hJRzV2Y20xaGJHbDZaUzVqYzNNZ2RqZ3VNQzR3SUh3Z1RVbFVJRXhwWTJWdWMyVWdmQ0JuYVhSb2RXSXVZMjl0TDI1bFkyOXNZWE12Ym05eWJXRnNhWHBsTG1OemN5QXFMMmgwYld4N2JHbHVaUzFvWldsbmFIUTZNUzR4TlRzdGQyVmlhMmwwTFhSbGVIUXRjMmw2WlMxaFpHcDFjM1E2TVRBd0pYMWliMlI1ZTIxaGNtZHBiam93ZldneGUyWnZiblF0YzJsNlpUb3laVzA3YldGeVoybHVPaTQyTjJWdElEQjlhSEo3WW05NExYTnBlbWx1WnpwamIyNTBaVzUwTFdKdmVEdG9aV2xuYUhRNk1EdHZkbVZ5Wm14dmR6cDJhWE5wWW14bGZYQnlaWHRtYjI1MExXWmhiV2xzZVRwdGIyNXZjM0JoWTJVc2JXOXViM053WVdObE8yWnZiblF0YzJsNlpUb3haVzE5WVh0aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T25SeVlXNXpjR0Z5Wlc1MGZXRmlZbkpiZEdsMGJHVmRlMkp2Y21SbGNpMWliM1IwYjIwNmJtOXVaVHQwWlhoMExXUmxZMjl5WVhScGIyNDZkVzVrWlhKc2FXNWxPeTEzWldKcmFYUXRkR1Y0ZEMxa1pXTnZjbUYwYVc5dU9uVnVaR1Z5YkdsdVpTQmtiM1IwWldRN2RHVjRkQzFrWldOdmNtRjBhVzl1T25WdVpHVnliR2x1WlNCa2IzUjBaV1I5WWl4emRISnZibWQ3Wm05dWRDMTNaV2xuYUhRNlltOXNaR1Z5ZldOdlpHVXNhMkprTEhOaGJYQjdabTl1ZEMxbVlXMXBiSGs2Ylc5dWIzTndZV05sTEcxdmJtOXpjR0ZqWlR0bWIyNTBMWE5wZW1VNk1XVnRmWE50WVd4c2UyWnZiblF0YzJsNlpUbzRNQ1Y5YzNWaUxITjFjSHRtYjI1MExYTnBlbVU2TnpVbE8yeHBibVV0YUdWcFoyaDBPakE3Y0c5emFYUnBiMjQ2Y21Wc1lYUnBkbVU3ZG1WeWRHbGpZV3d0WVd4cFoyNDZZbUZ6Wld4cGJtVjljM1ZpZTJKdmRIUnZiVG90TGpJMVpXMTljM1Z3ZTNSdmNEb3RMalZsYlgxcGJXZDdZbTl5WkdWeUxYTjBlV3hsT201dmJtVjlZblYwZEc5dUxHbHVjSFYwTEc5d2RHZHliM1Z3TEhObGJHVmpkQ3gwWlhoMFlYSmxZWHRtYjI1MExXWmhiV2xzZVRwcGJtaGxjbWwwTzJadmJuUXRjMmw2WlRveE1EQWxPMnhwYm1VdGFHVnBaMmgwT2pFdU1UVTdiV0Z5WjJsdU9qQjlZblYwZEc5dUxHbHVjSFYwZTI5MlpYSm1iRzkzT25acGMybGliR1Y5WW5WMGRHOXVMSE5sYkdWamRIdDBaWGgwTFhSeVlXNXpabTl5YlRwdWIyNWxmVnQwZVhCbFBXSjFkSFJ2Ymwwc1czUjVjR1U5Y21WelpYUmRMRnQwZVhCbFBYTjFZbTFwZEYwc1luVjBkRzl1ZXkxM1pXSnJhWFF0WVhCd1pXRnlZVzVqWlRwaWRYUjBiMjU5VzNSNWNHVTlZblYwZEc5dVhUbzZMVzF2ZWkxbWIyTjFjeTFwYm01bGNpeGJkSGx3WlQxeVpYTmxkRjA2T2kxdGIzb3RabTlqZFhNdGFXNXVaWElzVzNSNWNHVTljM1ZpYldsMFhUbzZMVzF2ZWkxbWIyTjFjeTFwYm01bGNpeGlkWFIwYjI0Nk9pMXRiM290Wm05amRYTXRhVzV1WlhKN1ltOXlaR1Z5TFhOMGVXeGxPbTV2Ym1VN2NHRmtaR2x1Wnpvd2ZWdDBlWEJsUFdKMWRIUnZibDA2TFcxdmVpMW1iMk4xYzNKcGJtY3NXM1I1Y0dVOWNtVnpaWFJkT2kxdGIzb3RabTlqZFhOeWFXNW5MRnQwZVhCbFBYTjFZbTFwZEYwNkxXMXZlaTFtYjJOMWMzSnBibWNzWW5WMGRHOXVPaTF0YjNvdFptOWpkWE55YVc1bmUyOTFkR3hwYm1VNk1YQjRJR1J2ZEhSbFpDQkNkWFIwYjI1VVpYaDBmV1pwWld4a2MyVjBlM0JoWkdScGJtYzZMak0xWlcwZ0xqYzFaVzBnTGpZeU5XVnRmV3hsWjJWdVpIdGliM2d0YzJsNmFXNW5PbUp2Y21SbGNpMWliM2c3WTI5c2IzSTZhVzVvWlhKcGREdGthWE53YkdGNU9uUmhZbXhsTzIxaGVDMTNhV1IwYURveE1EQWxPM0JoWkdScGJtYzZNRHQzYUdsMFpTMXpjR0ZqWlRwdWIzSnRZV3g5Y0hKdlozSmxjM043ZG1WeWRHbGpZV3d0WVd4cFoyNDZZbUZ6Wld4cGJtVjlkR1Y0ZEdGeVpXRjdiM1psY21ac2IzYzZZWFYwYjMxYmRIbHdaVDFqYUdWamEySnZlRjBzVzNSNWNHVTljbUZrYVc5ZGUySnZlQzF6YVhwcGJtYzZZbTl5WkdWeUxXSnZlRHR3WVdSa2FXNW5PakI5VzNSNWNHVTliblZ0WW1WeVhUbzZMWGRsWW10cGRDMXBibTVsY2kxemNHbHVMV0oxZEhSdmJpeGJkSGx3WlQxdWRXMWlaWEpkT2pvdGQyVmlhMmwwTFc5MWRHVnlMWE53YVc0dFluVjBkRzl1ZTJobGFXZG9kRHBoZFhSdmZWdDBlWEJsUFhObFlYSmphRjE3TFhkbFltdHBkQzFoY0hCbFlYSmhibU5sT25SbGVIUm1hV1ZzWkR0dmRYUnNhVzVsTFc5bVpuTmxkRG90TW5CNGZWdDBlWEJsUFhObFlYSmphRjA2T2kxM1pXSnJhWFF0YzJWaGNtTm9MV1JsWTI5eVlYUnBiMjU3TFhkbFltdHBkQzFoY0hCbFlYSmhibU5sT201dmJtVjlPam90ZDJWaWEybDBMV1pwYkdVdGRYQnNiMkZrTFdKMWRIUnZibnN0ZDJWaWEybDBMV0Z3Y0dWaGNtRnVZMlU2WW5WMGRHOXVPMlp2Ym5RNmFXNW9aWEpwZEgxa1pYUmhhV3h6ZTJScGMzQnNZWGs2WW14dlkydDljM1Z0YldGeWVYdGthWE53YkdGNU9teHBjM1F0YVhSbGJYMWJhR2xrWkdWdVhTeDBaVzF3YkdGMFpYdGthWE53YkdGNU9tNXZibVY5TG1KdmNtUmxjaTFpYjNnc1lTeGhjblJwWTJ4bExHRnphV1JsTEdKc2IyTnJjWFZ2ZEdVc1ltOWtlU3hqYjJSbExHUmtMR1JwZGl4a2JDeGtkQ3htYVdWc1pITmxkQ3htYVdkallYQjBhVzl1TEdacFozVnlaU3htYjI5MFpYSXNabTl5YlN4b01TeG9NaXhvTXl4b05DeG9OU3hvTml4b1pXRmtaWElzYUhSdGJDeHBibkIxZEZ0MGVYQmxQV1Z0WVdsc1hTeHBibkIxZEZ0MGVYQmxQVzUxYldKbGNsMHNhVzV3ZFhSYmRIbHdaVDF3WVhOemQyOXlaRjBzYVc1d2RYUmJkSGx3WlQxMFpXeGRMR2x1Y0hWMFczUjVjR1U5ZEdWNGRGMHNhVzV3ZFhSYmRIbHdaVDExY214ZExHeGxaMlZ1WkN4c2FTeHRZV2x1TEc1aGRpeHZiQ3h3TEhCeVpTeHpaV04wYVc5dUxIUmhZbXhsTEhSa0xIUmxlSFJoY21WaExIUm9MSFJ5TEhWc2UySnZlQzF6YVhwcGJtYzZZbTl5WkdWeUxXSnZlSDB1WVhOd1pXTjBMWEpoZEdsdmUyaGxhV2RvZERvd08zQnZjMmwwYVc5dU9uSmxiR0YwYVhabGZTNWhjM0JsWTNRdGNtRjBhVzh0TFRFMmVEbDdjR0ZrWkdsdVp5MWliM1IwYjIwNk5UWXVNalVsZlM1aGMzQmxZM1F0Y21GMGFXOHRMVGw0TVRaN2NHRmtaR2x1WnkxaWIzUjBiMjA2TVRjM0xqYzNKWDB1WVhOd1pXTjBMWEpoZEdsdkxTMDBlRE43Y0dGa1pHbHVaeTFpYjNSMGIyMDZOelVsZlM1aGMzQmxZM1F0Y21GMGFXOHRMVE40Tkh0d1lXUmthVzVuTFdKdmRIUnZiVG94TXpNdU16TWxmUzVoYzNCbFkzUXRjbUYwYVc4dExUWjROSHR3WVdSa2FXNW5MV0p2ZEhSdmJUbzJOaTQySlgwdVlYTndaV04wTFhKaGRHbHZMUzAwZURaN2NHRmtaR2x1WnkxaWIzUjBiMjA2TVRVd0pYMHVZWE53WldOMExYSmhkR2x2TFMwNGVEVjdjR0ZrWkdsdVp5MWliM1IwYjIwNk5qSXVOU1Y5TG1GemNHVmpkQzF5WVhScGJ5MHROWGc0ZTNCaFpHUnBibWN0WW05MGRHOXRPakUyTUNWOUxtRnpjR1ZqZEMxeVlYUnBieTB0TjNnMWUzQmhaR1JwYm1jdFltOTBkRzl0T2pjeExqUXlKWDB1WVhOd1pXTjBMWEpoZEdsdkxTMDFlRGQ3Y0dGa1pHbHVaeTFpYjNSMGIyMDZNVFF3SlgwdVlYTndaV04wTFhKaGRHbHZMUzB4ZURGN2NHRmtaR2x1WnkxaWIzUjBiMjA2TVRBd0pYMHVZWE53WldOMExYSmhkR2x2TFMxdlltcGxZM1I3Y0c5emFYUnBiMjQ2WVdKemIyeDFkR1U3ZEc5d09qQTdjbWxuYUhRNk1EdGliM1IwYjIwNk1EdHNaV1owT2pBN2QybGtkR2c2TVRBd0pUdG9aV2xuYUhRNk1UQXdKVHQ2TFdsdVpHVjRPakV3TUgxcGJXZDdiV0Y0TFhkcFpIUm9PakV3TUNWOUxtTnZkbVZ5ZTJKaFkydG5jbTkxYm1RdGMybDZaVHBqYjNabGNpRnBiWEJ2Y25SaGJuUjlMbU52Ym5SaGFXNTdZbUZqYTJkeWIzVnVaQzF6YVhwbE9tTnZiblJoYVc0aGFXMXdiM0owWVc1MGZTNWlaeTFqWlc1MFpYSjdZbUZqYTJkeWIzVnVaQzF3YjNOcGRHbHZiam8xTUNWOUxtSm5MV05sYm5SbGNpd3VZbWN0ZEc5d2UySmhZMnRuY205MWJtUXRjbVZ3WldGME9tNXZMWEpsY0dWaGRIMHVZbWN0ZEc5d2UySmhZMnRuY205MWJtUXRjRzl6YVhScGIyNDZkRzl3ZlM1aVp5MXlhV2RvZEh0aVlXTnJaM0p2ZFc1a0xYQnZjMmwwYVc5dU9qRXdNQ1Y5TG1KbkxXSnZkSFJ2YlN3dVltY3RjbWxuYUhSN1ltRmphMmR5YjNWdVpDMXlaWEJsWVhRNmJtOHRjbVZ3WldGMGZTNWlaeTFpYjNSMGIyMTdZbUZqYTJkeWIzVnVaQzF3YjNOcGRHbHZianBpYjNSMGIyMTlMbUpuTFd4bFpuUjdZbUZqYTJkeWIzVnVaQzF5WlhCbFlYUTZibTh0Y21Wd1pXRjBPMkpoWTJ0bmNtOTFibVF0Y0c5emFYUnBiMjQ2TUgwdWIzVjBiR2x1Wlh0dmRYUnNhVzVsT2pGd2VDQnpiMnhwWkgwdWIzVjBiR2x1WlMxMGNtRnVjM0JoY21WdWRIdHZkWFJzYVc1bE9qRndlQ0J6YjJ4cFpDQjBjbUZ1YzNCaGNtVnVkSDB1YjNWMGJHbHVaUzB3ZTI5MWRHeHBibVU2TUgwdVltRjdZbTl5WkdWeUxYTjBlV3hsT25OdmJHbGtPMkp2Y21SbGNpMTNhV1IwYURveGNIaDlMbUowZTJKdmNtUmxjaTEwYjNBdGMzUjViR1U2YzI5c2FXUTdZbTl5WkdWeUxYUnZjQzEzYVdSMGFEb3hjSGg5TG1KeWUySnZjbVJsY2kxeWFXZG9kQzF6ZEhsc1pUcHpiMnhwWkR0aWIzSmtaWEl0Y21sbmFIUXRkMmxrZEdnNk1YQjRmUzVpWW50aWIzSmtaWEl0WW05MGRHOXRMWE4wZVd4bE9uTnZiR2xrTzJKdmNtUmxjaTFpYjNSMGIyMHRkMmxrZEdnNk1YQjRmUzVpYkh0aWIzSmtaWEl0YkdWbWRDMXpkSGxzWlRwemIyeHBaRHRpYjNKa1pYSXRiR1ZtZEMxM2FXUjBhRG94Y0hoOUxtSnVlMkp2Y21SbGNpMXpkSGxzWlRwdWIyNWxPMkp2Y21SbGNpMTNhV1IwYURvd2ZTNWlMUzFpYkdGamEzdGliM0prWlhJdFkyOXNiM0k2SXpBd01IMHVZaTB0Ym1WaGNpMWliR0ZqYTN0aWIzSmtaWEl0WTI5c2IzSTZJekV4TVgwdVlpMHRaR0Z5YXkxbmNtRjVlMkp2Y21SbGNpMWpiMnh2Y2pvak16TXpmUzVpTFMxdGFXUXRaM0poZVh0aWIzSmtaWEl0WTI5c2IzSTZJelUxTlgwdVlpMHRaM0poZVh0aWIzSmtaWEl0WTI5c2IzSTZJemMzTjMwdVlpMHRjMmxzZG1WeWUySnZjbVJsY2kxamIyeHZjam9qT1RrNWZTNWlMUzFzYVdkb2RDMXphV3gyWlhKN1ltOXlaR1Z5TFdOdmJHOXlPaU5oWVdGOUxtSXRMVzF2YjI0dFozSmhlWHRpYjNKa1pYSXRZMjlzYjNJNkkyTmpZMzB1WWkwdGJHbG5hSFF0WjNKaGVYdGliM0prWlhJdFkyOXNiM0k2STJWbFpYMHVZaTB0Ym1WaGNpMTNhR2wwWlh0aWIzSmtaWEl0WTI5c2IzSTZJMlkwWmpSbU5IMHVZaTB0ZDJocGRHVjdZbTl5WkdWeUxXTnZiRzl5T2lObVptWjlMbUl0TFhkb2FYUmxMVGt3ZTJKdmNtUmxjaTFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU9TbDlMbUl0TFhkb2FYUmxMVGd3ZTJKdmNtUmxjaTFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU9DbDlMbUl0TFhkb2FYUmxMVGN3ZTJKdmNtUmxjaTFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU55bDlMbUl0TFhkb2FYUmxMVFl3ZTJKdmNtUmxjaTFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU5pbDlMbUl0TFhkb2FYUmxMVFV3ZTJKdmNtUmxjaTFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU5TbDlMbUl0TFhkb2FYUmxMVFF3ZTJKdmNtUmxjaTFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU5DbDlMbUl0TFhkb2FYUmxMVE13ZTJKdmNtUmxjaTFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU15bDlMbUl0TFhkb2FYUmxMVEl3ZTJKdmNtUmxjaTFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU1pbDlMbUl0TFhkb2FYUmxMVEV3ZTJKdmNtUmxjaTFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU1TbDlMbUl0TFhkb2FYUmxMVEExZTJKdmNtUmxjaTFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU1EVXBmUzVpTFMxM2FHbDBaUzB3TWpWN1ltOXlaR1Z5TFdOdmJHOXlPbWh6YkdFb01Dd3dKU3d4TURBbExDNHdNalVwZlM1aUxTMTNhR2wwWlMwd01USTFlMkp2Y21SbGNpMWpiMnh2Y2pwb2MyeGhLREFzTUNVc01UQXdKU3d1TURFeU5TbDlMbUl0TFdKc1lXTnJMVGt3ZTJKdmNtUmxjaTFqYjJ4dmNqcHlaMkpoS0RBc01Dd3dMQzQ1S1gwdVlpMHRZbXhoWTJzdE9EQjdZbTl5WkdWeUxXTnZiRzl5T25KblltRW9NQ3d3TERBc0xqZ3BmUzVpTFMxaWJHRmpheTAzTUh0aWIzSmtaWEl0WTI5c2IzSTZjbWRpWVNnd0xEQXNNQ3d1TnlsOUxtSXRMV0pzWVdOckxUWXdlMkp2Y21SbGNpMWpiMnh2Y2pweVoySmhLREFzTUN3d0xDNDJLWDB1WWkwdFlteGhZMnN0TlRCN1ltOXlaR1Z5TFdOdmJHOXlPbkpuWW1Fb01Dd3dMREFzTGpVcGZTNWlMUzFpYkdGamF5MDBNSHRpYjNKa1pYSXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VOQ2w5TG1JdExXSnNZV05yTFRNd2UySnZjbVJsY2kxamIyeHZjanB5WjJKaEtEQXNNQ3d3TEM0ektYMHVZaTB0WW14aFkyc3RNakI3WW05eVpHVnlMV052Ykc5eU9uSm5ZbUVvTUN3d0xEQXNMaklwZlM1aUxTMWliR0ZqYXkweE1IdGliM0prWlhJdFkyOXNiM0k2Y21kaVlTZ3dMREFzTUN3dU1TbDlMbUl0TFdKc1lXTnJMVEExZTJKdmNtUmxjaTFqYjJ4dmNqcHlaMkpoS0RBc01Dd3dMQzR3TlNsOUxtSXRMV0pzWVdOckxUQXlOWHRpYjNKa1pYSXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VNREkxS1gwdVlpMHRZbXhoWTJzdE1ERXlOWHRpYjNKa1pYSXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VNREV5TlNsOUxtSXRMV1JoY21zdGNtVmtlMkp2Y21SbGNpMWpiMnh2Y2pvalpUY3dOREJtZlM1aUxTMXlaV1I3WW05eVpHVnlMV052Ykc5eU9pTm1aalF4TXpaOUxtSXRMV3hwWjJoMExYSmxaSHRpYjNKa1pYSXRZMjlzYjNJNkkyWm1OekkxWTMwdVlpMHRiM0poYm1kbGUySnZjbVJsY2kxamIyeHZjam9qWm1ZMk16QXdmUzVpTFMxbmIyeGtlMkp2Y21SbGNpMWpiMnh2Y2pvalptWmlOekF3ZlM1aUxTMTVaV3hzYjNkN1ltOXlaR1Z5TFdOdmJHOXlPbWR2YkdSOUxtSXRMV3hwWjJoMExYbGxiR3h2ZDN0aWIzSmtaWEl0WTI5c2IzSTZJMlppWmpGaE9YMHVZaTB0Y0hWeWNHeGxlMkp2Y21SbGNpMWpiMnh2Y2pvak5XVXlZMkUxZlM1aUxTMXNhV2RvZEMxd2RYSndiR1Y3WW05eVpHVnlMV052Ykc5eU9pTmhORFl6WmpKOUxtSXRMV1JoY21zdGNHbHVhM3RpYjNKa1pYSXRZMjlzYjNJNkkyUTFNREE0Wm4wdVlpMHRhRzkwTFhCcGJtdDdZbTl5WkdWeUxXTnZiRzl5T2lObVpqUXhZalI5TG1JdExYQnBibXQ3WW05eVpHVnlMV052Ykc5eU9pTm1aamd3WTJOOUxtSXRMV3hwWjJoMExYQnBibXQ3WW05eVpHVnlMV052Ykc5eU9pTm1abUV6WkRkOUxtSXRMV1JoY21zdFozSmxaVzU3WW05eVpHVnlMV052Ykc5eU9pTXhNemMzTlRKOUxtSXRMV2R5WldWdWUySnZjbVJsY2kxamIyeHZjam9qTVRsaE9UYzBmUzVpTFMxc2FXZG9kQzFuY21WbGJudGliM0prWlhJdFkyOXNiM0k2SXpsbFpXSmpabjB1WWkwdGJtRjJlWHRpYjNKa1pYSXRZMjlzYjNJNkl6QXdNV0kwTkgwdVlpMHRaR0Z5YXkxaWJIVmxlMkp2Y21SbGNpMWpiMnh2Y2pvak1EQTBORGxsZlM1aUxTMWliSFZsZTJKdmNtUmxjaTFqYjJ4dmNqb2pNelUzWldSa2ZTNWlMUzFzYVdkb2RDMWliSFZsZTJKdmNtUmxjaTFqYjJ4dmNqb2pPVFpqWTJabWZTNWlMUzFzYVdkb2RHVnpkQzFpYkhWbGUySnZjbVJsY2kxamIyeHZjam9qWTJSbFkyWm1mUzVpTFMxM1lYTm9aV1F0WW14MVpYdGliM0prWlhJdFkyOXNiM0k2STJZMlptWm1aWDB1WWkwdGQyRnphR1ZrTFdkeVpXVnVlMkp2Y21SbGNpMWpiMnh2Y2pvalpUaG1aR1kxZlM1aUxTMTNZWE5vWldRdGVXVnNiRzkzZTJKdmNtUmxjaTFqYjJ4dmNqb2pabVptWTJWaWZTNWlMUzEzWVhOb1pXUXRjbVZrZTJKdmNtUmxjaTFqYjJ4dmNqb2pabVprWm1SbWZTNWlMUzEwY21GdWMzQmhjbVZ1ZEh0aWIzSmtaWEl0WTI5c2IzSTZkSEpoYm5Od1lYSmxiblI5TG1JdExXbHVhR1Z5YVhSN1ltOXlaR1Z5TFdOdmJHOXlPbWx1YUdWeWFYUjlMbUp5TUh0aWIzSmtaWEl0Y21Ga2FYVnpPakI5TG1KeU1YdGliM0prWlhJdGNtRmthWFZ6T2k0eE1qVnlaVzE5TG1KeU1udGliM0prWlhJdGNtRmthWFZ6T2k0eU5YSmxiWDB1WW5JemUySnZjbVJsY2kxeVlXUnBkWE02TGpWeVpXMTlMbUp5Tkh0aWIzSmtaWEl0Y21Ga2FYVnpPakZ5WlcxOUxtSnlMVEV3TUh0aWIzSmtaWEl0Y21Ga2FYVnpPakV3TUNWOUxtSnlMWEJwYkd4N1ltOXlaR1Z5TFhKaFpHbDFjem81T1RrNWNIaDlMbUp5TFMxaWIzUjBiMjE3WW05eVpHVnlMWFJ2Y0Mxc1pXWjBMWEpoWkdsMWN6b3dPMkp2Y21SbGNpMTBiM0F0Y21sbmFIUXRjbUZrYVhWek9qQjlMbUp5TFMxMGIzQjdZbTl5WkdWeUxXSnZkSFJ2YlMxeWFXZG9kQzF5WVdScGRYTTZNSDB1WW5JdExYSnBaMmgwTEM1aWNpMHRkRzl3ZTJKdmNtUmxjaTFpYjNSMGIyMHRiR1ZtZEMxeVlXUnBkWE02TUgwdVluSXRMWEpwWjJoMGUySnZjbVJsY2kxMGIzQXRiR1ZtZEMxeVlXUnBkWE02TUgwdVluSXRMV3hsWm5SN1ltOXlaR1Z5TFhSdmNDMXlhV2RvZEMxeVlXUnBkWE02TUR0aWIzSmtaWEl0WW05MGRHOXRMWEpwWjJoMExYSmhaR2wxY3pvd2ZTNWlMUzFrYjNSMFpXUjdZbTl5WkdWeUxYTjBlV3hsT21SdmRIUmxaSDB1WWkwdFpHRnphR1ZrZTJKdmNtUmxjaTF6ZEhsc1pUcGtZWE5vWldSOUxtSXRMWE52Ykdsa2UySnZjbVJsY2kxemRIbHNaVHB6YjJ4cFpIMHVZaTB0Ym05dVpYdGliM0prWlhJdGMzUjViR1U2Ym05dVpYMHVZbmN3ZTJKdmNtUmxjaTEzYVdSMGFEb3dmUzVpZHpGN1ltOXlaR1Z5TFhkcFpIUm9PaTR4TWpWeVpXMTlMbUozTW50aWIzSmtaWEl0ZDJsa2RHZzZMakkxY21WdGZTNWlkek43WW05eVpHVnlMWGRwWkhSb09pNDFjbVZ0ZlM1aWR6UjdZbTl5WkdWeUxYZHBaSFJvT2pGeVpXMTlMbUozTlh0aWIzSmtaWEl0ZDJsa2RHZzZNbkpsYlgwdVluUXRNSHRpYjNKa1pYSXRkRzl3TFhkcFpIUm9PakI5TG1KeUxUQjdZbTl5WkdWeUxYSnBaMmgwTFhkcFpIUm9PakI5TG1KaUxUQjdZbTl5WkdWeUxXSnZkSFJ2YlMxM2FXUjBhRG93ZlM1aWJDMHdlMkp2Y21SbGNpMXNaV1owTFhkcFpIUm9PakI5TG5Ob1lXUnZkeTB4ZTJKdmVDMXphR0ZrYjNjNk1DQXdJRFJ3ZUNBeWNIZ2djbWRpWVNnd0xEQXNNQ3d1TWlsOUxuTm9ZV1J2ZHkweWUySnZlQzF6YUdGa2IzYzZNQ0F3SURod2VDQXljSGdnY21kaVlTZ3dMREFzTUN3dU1pbDlMbk5vWVdSdmR5MHplMkp2ZUMxemFHRmtiM2M2TW5CNElESndlQ0EwY0hnZ01uQjRJSEpuWW1Fb01Dd3dMREFzTGpJcGZTNXphR0ZrYjNjdE5IdGliM2d0YzJoaFpHOTNPakp3ZUNBeWNIZ2dPSEI0SURBZ2NtZGlZU2d3TERBc01Dd3VNaWw5TG5Ob1lXUnZkeTAxZTJKdmVDMXphR0ZrYjNjNk5IQjRJRFJ3ZUNBNGNIZ2dNQ0J5WjJKaEtEQXNNQ3d3TEM0eUtYMHVjSEpsZTI5MlpYSm1iRzkzTFhnNllYVjBienR2ZG1WeVpteHZkeTE1T21ocFpHUmxianR2ZG1WeVpteHZkenB6WTNKdmJHeDlMblJ2Y0Mwd2UzUnZjRG93ZlM1eWFXZG9kQzB3ZTNKcFoyaDBPakI5TG1KdmRIUnZiUzB3ZTJKdmRIUnZiVG93ZlM1c1pXWjBMVEI3YkdWbWREb3dmUzUwYjNBdE1YdDBiM0E2TVhKbGJYMHVjbWxuYUhRdE1YdHlhV2RvZERveGNtVnRmUzVpYjNSMGIyMHRNWHRpYjNSMGIyMDZNWEpsYlgwdWJHVm1kQzB4ZTJ4bFpuUTZNWEpsYlgwdWRHOXdMVEo3ZEc5d09qSnlaVzE5TG5KcFoyaDBMVEo3Y21sbmFIUTZNbkpsYlgwdVltOTBkRzl0TFRKN1ltOTBkRzl0T2pKeVpXMTlMbXhsWm5RdE1udHNaV1owT2pKeVpXMTlMblJ2Y0MwdE1YdDBiM0E2TFRGeVpXMTlMbkpwWjJoMExTMHhlM0pwWjJoME9pMHhjbVZ0ZlM1aWIzUjBiMjB0TFRGN1ltOTBkRzl0T2kweGNtVnRmUzVzWldaMExTMHhlMnhsWm5RNkxURnlaVzE5TG5SdmNDMHRNbnQwYjNBNkxUSnlaVzE5TG5KcFoyaDBMUzB5ZTNKcFoyaDBPaTB5Y21WdGZTNWliM1IwYjIwdExUSjdZbTkwZEc5dE9pMHljbVZ0ZlM1c1pXWjBMUzB5ZTJ4bFpuUTZMVEp5WlcxOUxtRmljMjlzZFhSbExTMW1hV3hzZTNSdmNEb3dPM0pwWjJoME9qQTdZbTkwZEc5dE9qQTdiR1ZtZERvd2ZTNWpaanBoWm5SbGNpd3VZMlk2WW1WbWIzSmxlMk52Ym5SbGJuUTZYQ0lnWENJN1pHbHpjR3hoZVRwMFlXSnNaWDB1WTJZNllXWjBaWEo3WTJ4bFlYSTZZbTkwYUgwdVkyWjdLbnB2YjIwNk1YMHVZMng3WTJ4bFlYSTZiR1ZtZEgwdVkzSjdZMnhsWVhJNmNtbG5hSFI5TG1OaWUyTnNaV0Z5T21KdmRHaDlMbU51ZTJOc1pXRnlPbTV2Ym1WOUxtUnVlMlJwYzNCc1lYazZibTl1WlgwdVpHbDdaR2x6Y0d4aGVUcHBibXhwYm1WOUxtUmllMlJwYzNCc1lYazZZbXh2WTJ0OUxtUnBZbnRrYVhOd2JHRjVPbWx1YkdsdVpTMWliRzlqYTMwdVpHbDBlMlJwYzNCc1lYazZhVzVzYVc1bExYUmhZbXhsZlM1a2RIdGthWE53YkdGNU9uUmhZbXhsZlM1a2RHTjdaR2x6Y0d4aGVUcDBZV0pzWlMxalpXeHNmUzVrZEMxeWIzZDdaR2x6Y0d4aGVUcDBZV0pzWlMxeWIzZDlMbVIwTFhKdmR5MW5jbTkxY0h0a2FYTndiR0Y1T25SaFlteGxMWEp2ZHkxbmNtOTFjSDB1WkhRdFkyOXNkVzF1ZTJScGMzQnNZWGs2ZEdGaWJHVXRZMjlzZFcxdWZTNWtkQzFqYjJ4MWJXNHRaM0p2ZFhCN1pHbHpjR3hoZVRwMFlXSnNaUzFqYjJ4MWJXNHRaM0p2ZFhCOUxtUjBMUzFtYVhobFpIdDBZV0pzWlMxc1lYbHZkWFE2Wm1sNFpXUTdkMmxrZEdnNk1UQXdKWDB1Wm14bGVIdGthWE53YkdGNU9tWnNaWGg5TG1sdWJHbHVaUzFtYkdWNGUyUnBjM0JzWVhrNmFXNXNhVzVsTFdac1pYaDlMbVpzWlhndFlYVjBiM3RtYkdWNE9qRWdNU0JoZFhSdk8yMXBiaTEzYVdSMGFEb3dPMjFwYmkxb1pXbG5hSFE2TUgwdVpteGxlQzF1YjI1bGUyWnNaWGc2Ym05dVpYMHVabXhsZUMxamIyeDFiVzU3Wm14bGVDMWthWEpsWTNScGIyNDZZMjlzZFcxdWZTNW1iR1Y0TFhKdmQzdG1iR1Y0TFdScGNtVmpkR2x2YmpweWIzZDlMbVpzWlhndGQzSmhjSHRtYkdWNExYZHlZWEE2ZDNKaGNIMHVabXhsZUMxdWIzZHlZWEI3Wm14bGVDMTNjbUZ3T201dmQzSmhjSDB1Wm14bGVDMTNjbUZ3TFhKbGRtVnljMlY3Wm14bGVDMTNjbUZ3T25keVlYQXRjbVYyWlhKelpYMHVabXhsZUMxamIyeDFiVzR0Y21WMlpYSnpaWHRtYkdWNExXUnBjbVZqZEdsdmJqcGpiMngxYlc0dGNtVjJaWEp6WlgwdVpteGxlQzF5YjNjdGNtVjJaWEp6Wlh0bWJHVjRMV1JwY21WamRHbHZianB5YjNjdGNtVjJaWEp6WlgwdWFYUmxiWE10YzNSaGNuUjdZV3hwWjI0dGFYUmxiWE02Wm14bGVDMXpkR0Z5ZEgwdWFYUmxiWE10Wlc1a2UyRnNhV2R1TFdsMFpXMXpPbVpzWlhndFpXNWtmUzVwZEdWdGN5MWpaVzUwWlhKN1lXeHBaMjR0YVhSbGJYTTZZMlZ1ZEdWeWZTNXBkR1Z0Y3kxaVlYTmxiR2x1Wlh0aGJHbG5iaTFwZEdWdGN6cGlZWE5sYkdsdVpYMHVhWFJsYlhNdGMzUnlaWFJqYUh0aGJHbG5iaTFwZEdWdGN6cHpkSEpsZEdOb2ZTNXpaV3htTFhOMFlYSjBlMkZzYVdkdUxYTmxiR1k2Wm14bGVDMXpkR0Z5ZEgwdWMyVnNaaTFsYm1SN1lXeHBaMjR0YzJWc1pqcG1iR1Y0TFdWdVpIMHVjMlZzWmkxalpXNTBaWEo3WVd4cFoyNHRjMlZzWmpwalpXNTBaWEo5TG5ObGJHWXRZbUZ6Wld4cGJtVjdZV3hwWjI0dGMyVnNaanBpWVhObGJHbHVaWDB1YzJWc1ppMXpkSEpsZEdOb2UyRnNhV2R1TFhObGJHWTZjM1J5WlhSamFIMHVhblZ6ZEdsbWVTMXpkR0Z5ZEh0cWRYTjBhV1o1TFdOdmJuUmxiblE2Wm14bGVDMXpkR0Z5ZEgwdWFuVnpkR2xtZVMxbGJtUjdhblZ6ZEdsbWVTMWpiMjUwWlc1ME9tWnNaWGd0Wlc1a2ZTNXFkWE4wYVdaNUxXTmxiblJsY250cWRYTjBhV1o1TFdOdmJuUmxiblE2WTJWdWRHVnlmUzVxZFhOMGFXWjVMV0psZEhkbFpXNTdhblZ6ZEdsbWVTMWpiMjUwWlc1ME9uTndZV05sTFdKbGRIZGxaVzU5TG1wMWMzUnBabmt0WVhKdmRXNWtlMnAxYzNScFpua3RZMjl1ZEdWdWREcHpjR0ZqWlMxaGNtOTFibVI5TG1OdmJuUmxiblF0YzNSaGNuUjdZV3hwWjI0dFkyOXVkR1Z1ZERwbWJHVjRMWE4wWVhKMGZTNWpiMjUwWlc1MExXVnVaSHRoYkdsbmJpMWpiMjUwWlc1ME9tWnNaWGd0Wlc1a2ZTNWpiMjUwWlc1MExXTmxiblJsY250aGJHbG5iaTFqYjI1MFpXNTBPbU5sYm5SbGNuMHVZMjl1ZEdWdWRDMWlaWFIzWldWdWUyRnNhV2R1TFdOdmJuUmxiblE2YzNCaFkyVXRZbVYwZDJWbGJuMHVZMjl1ZEdWdWRDMWhjbTkxYm1SN1lXeHBaMjR0WTI5dWRHVnVkRHB6Y0dGalpTMWhjbTkxYm1SOUxtTnZiblJsYm5RdGMzUnlaWFJqYUh0aGJHbG5iaTFqYjI1MFpXNTBPbk4wY21WMFkyaDlMbTl5WkdWeUxUQjdiM0prWlhJNk1IMHViM0prWlhJdE1YdHZjbVJsY2pveGZTNXZjbVJsY2kweWUyOXlaR1Z5T2pKOUxtOXlaR1Z5TFRON2IzSmtaWEk2TTMwdWIzSmtaWEl0Tkh0dmNtUmxjam8wZlM1dmNtUmxjaTAxZTI5eVpHVnlPalY5TG05eVpHVnlMVFo3YjNKa1pYSTZObjB1YjNKa1pYSXROM3R2Y21SbGNqbzNmUzV2Y21SbGNpMDRlMjl5WkdWeU9qaDlMbTl5WkdWeUxXeGhjM1I3YjNKa1pYSTZPVGs1T1RsOUxtWnNaWGd0WjNKdmR5MHdlMlpzWlhndFozSnZkem93ZlM1bWJHVjRMV2R5YjNjdE1YdG1iR1Y0TFdkeWIzYzZNWDB1Wm14bGVDMXphSEpwYm1zdE1IdG1iR1Y0TFhOb2NtbHVhem93ZlM1bWJHVjRMWE5vY21sdWF5MHhlMlpzWlhndGMyaHlhVzVyT2pGOUxtWnNlMlpzYjJGME9teGxablI5TG1ac0xDNW1jbnRmWkdsemNHeGhlVHBwYm14cGJtVjlMbVp5ZTJac2IyRjBPbkpwWjJoMGZTNW1ibnRtYkc5aGREcHViMjVsZlM1ellXNXpMWE5sY21sbWUyWnZiblF0Wm1GdGFXeDVPaTFoY0hCc1pTMXplWE4wWlcwc1FteHBibXROWVdOVGVYTjBaVzFHYjI1MExHRjJaVzVwY2lCdVpYaDBMR0YyWlc1cGNpeG9aV3gyWlhScFkyRWdibVYxWlN4b1pXeDJaWFJwWTJFc2RXSjFiblIxTEhKdlltOTBieXh1YjNSdkxITmxaMjlsSUhWcExHRnlhV0ZzTEhOaGJuTXRjMlZ5YVdaOUxuTmxjbWxtZTJadmJuUXRabUZ0YVd4NU9tZGxiM0puYVdFc2RHbHRaWE1zYzJWeWFXWjlMbk41YzNSbGJTMXpZVzV6TFhObGNtbG1lMlp2Ym5RdFptRnRhV3g1T25OaGJuTXRjMlZ5YVdaOUxuTjVjM1JsYlMxelpYSnBabnRtYjI1MExXWmhiV2xzZVRwelpYSnBabjB1WTI5a1pTeGpiMlJsZTJadmJuUXRabUZ0YVd4NU9rTnZibk52YkdGekxHMXZibUZqYnl4dGIyNXZjM0JoWTJWOUxtTnZkWEpwWlhKN1ptOXVkQzFtWVcxcGJIazZRMjkxY21sbGNpQk9aWGgwTEdOdmRYSnBaWElzYlc5dWIzTndZV05sZlM1b1pXeDJaWFJwWTJGN1ptOXVkQzFtWVcxcGJIazZhR1ZzZG1WMGFXTmhJRzVsZFdVc2FHVnNkbVYwYVdOaExITmhibk10YzJWeWFXWjlMbUYyWlc1cGNudG1iMjUwTFdaaGJXbHNlVHBoZG1WdWFYSWdibVY0ZEN4aGRtVnVhWElzYzJGdWN5MXpaWEpwWm4wdVlYUm9aV3hoYzN0bWIyNTBMV1poYldsc2VUcGhkR2hsYkdGekxHZGxiM0puYVdFc2MyVnlhV1o5TG1kbGIzSm5hV0Y3Wm05dWRDMW1ZVzFwYkhrNloyVnZjbWRwWVN4elpYSnBabjB1ZEdsdFpYTjdabTl1ZEMxbVlXMXBiSGs2ZEdsdFpYTXNjMlZ5YVdaOUxtSnZaRzl1YVh0bWIyNTBMV1poYldsc2VUcENiMlJ2Ym1rZ1RWUXNjMlZ5YVdaOUxtTmhiR2x6ZEc5N1ptOXVkQzFtWVcxcGJIazZRMkZzYVhOMGJ5Qk5WQ3h6WlhKcFpuMHVaMkZ5WVcxdmJtUjdabTl1ZEMxbVlXMXBiSGs2WjJGeVlXMXZibVFzYzJWeWFXWjlMbUpoYzJ0bGNuWnBiR3hsZTJadmJuUXRabUZ0YVd4NU9tSmhjMnRsY25acGJHeGxMSE5sY21sbWZTNXBlMlp2Ym5RdGMzUjViR1U2YVhSaGJHbGpmUzVtY3kxdWIzSnRZV3g3Wm05dWRDMXpkSGxzWlRwdWIzSnRZV3g5TG01dmNtMWhiSHRtYjI1MExYZGxhV2RvZERvME1EQjlMbUo3Wm05dWRDMTNaV2xuYUhRNk56QXdmUzVtZHpGN1ptOXVkQzEzWldsbmFIUTZNVEF3ZlM1bWR6SjdabTl1ZEMxM1pXbG5hSFE2TWpBd2ZTNW1kek43Wm05dWRDMTNaV2xuYUhRNk16QXdmUzVtZHpSN1ptOXVkQzEzWldsbmFIUTZOREF3ZlM1bWR6VjdabTl1ZEMxM1pXbG5hSFE2TlRBd2ZTNW1kelo3Wm05dWRDMTNaV2xuYUhRNk5qQXdmUzVtZHpkN1ptOXVkQzEzWldsbmFIUTZOekF3ZlM1bWR6aDdabTl1ZEMxM1pXbG5hSFE2T0RBd2ZTNW1kemw3Wm05dWRDMTNaV2xuYUhRNk9UQXdmUzVwYm5CMWRDMXlaWE5sZEhzdGQyVmlhMmwwTFdGd2NHVmhjbUZ1WTJVNmJtOXVaVHN0Ylc5NkxXRndjR1ZoY21GdVkyVTZibTl1WlgwdVluVjBkRzl1TFhKbGMyVjBPam90Ylc5NkxXWnZZM1Z6TFdsdWJtVnlMQzVwYm5CMWRDMXlaWE5sZERvNkxXMXZlaTFtYjJOMWN5MXBibTVsY250aWIzSmtaWEk2TUR0d1lXUmthVzVuT2pCOUxtZ3hlMmhsYVdkb2REb3hjbVZ0ZlM1b01udG9aV2xuYUhRNk1uSmxiWDB1YURON2FHVnBaMmgwT2pSeVpXMTlMbWcwZTJobGFXZG9kRG80Y21WdGZTNW9OWHRvWldsbmFIUTZNVFp5WlcxOUxtZ3RNalY3YUdWcFoyaDBPakkxSlgwdWFDMDFNSHRvWldsbmFIUTZOVEFsZlM1b0xUYzFlMmhsYVdkb2REbzNOU1Y5TG1ndE1UQXdlMmhsYVdkb2REb3hNREFsZlM1dGFXNHRhQzB4TURCN2JXbHVMV2hsYVdkb2REb3hNREFsZlM1MmFDMHlOWHRvWldsbmFIUTZNalYyYUgwdWRtZ3ROVEI3YUdWcFoyaDBPalV3ZG1oOUxuWm9MVGMxZTJobGFXZG9kRG8zTlhab2ZTNTJhQzB4TURCN2FHVnBaMmgwT2pFd01IWm9mUzV0YVc0dGRtZ3RNVEF3ZTIxcGJpMW9aV2xuYUhRNk1UQXdkbWg5TG1ndFlYVjBiM3RvWldsbmFIUTZZWFYwYjMwdWFDMXBibWhsY21sMGUyaGxhV2RvZERwcGJtaGxjbWwwZlM1MGNtRmphMlZrZTJ4bGRIUmxjaTF6Y0dGamFXNW5PaTR4WlcxOUxuUnlZV05yWldRdGRHbG5hSFI3YkdWMGRHVnlMWE53WVdOcGJtYzZMUzR3TldWdGZTNTBjbUZqYTJWa0xXMWxaMkY3YkdWMGRHVnlMWE53WVdOcGJtYzZMakkxWlcxOUxteG9MWE52Ykdsa2UyeHBibVV0YUdWcFoyaDBPakY5TG14b0xYUnBkR3hsZTJ4cGJtVXRhR1ZwWjJoME9qRXVNalY5TG14b0xXTnZjSGw3YkdsdVpTMW9aV2xuYUhRNk1TNDFmUzVzYVc1cmUzUmxlSFF0WkdWamIzSmhkR2x2YmpwdWIyNWxmUzVzYVc1ckxDNXNhVzVyT21GamRHbDJaU3d1YkdsdWF6cG1iMk4xY3l3dWJHbHVhenBvYjNabGNpd3ViR2x1YXpwc2FXNXJMQzVzYVc1ck9uWnBjMmwwWldSN2RISmhibk5wZEdsdmJqcGpiMnh2Y2lBdU1UVnpJR1ZoYzJVdGFXNTlMbXhwYm1zNlptOWpkWE43YjNWMGJHbHVaVG94Y0hnZ1pHOTBkR1ZrSUdOMWNuSmxiblJEYjJ4dmNuMHViR2x6ZEh0c2FYTjBMWE4wZVd4bExYUjVjR1U2Ym05dVpYMHViWGN0TVRBd2UyMWhlQzEzYVdSMGFEb3hNREFsZlM1dGR6RjdiV0Y0TFhkcFpIUm9PakZ5WlcxOUxtMTNNbnR0WVhndGQybGtkR2c2TW5KbGJYMHViWGN6ZTIxaGVDMTNhV1IwYURvMGNtVnRmUzV0ZHpSN2JXRjRMWGRwWkhSb09qaHlaVzE5TG0xM05YdHRZWGd0ZDJsa2RHZzZNVFp5WlcxOUxtMTNObnR0WVhndGQybGtkR2c2TXpKeVpXMTlMbTEzTjN0dFlYZ3RkMmxrZEdnNk5EaHlaVzE5TG0xM09IdHRZWGd0ZDJsa2RHZzZOalJ5WlcxOUxtMTNPWHR0WVhndGQybGtkR2c2T1RaeVpXMTlMbTEzTFc1dmJtVjdiV0Y0TFhkcFpIUm9PbTV2Ym1WOUxuY3hlM2RwWkhSb09qRnlaVzE5TG5jeWUzZHBaSFJvT2pKeVpXMTlMbmN6ZTNkcFpIUm9PalJ5WlcxOUxuYzBlM2RwWkhSb09qaHlaVzE5TG5jMWUzZHBaSFJvT2pFMmNtVnRmUzUzTFRFd2UzZHBaSFJvT2pFd0pYMHVkeTB5TUh0M2FXUjBhRG95TUNWOUxuY3RNalY3ZDJsa2RHZzZNalVsZlM1M0xUTXdlM2RwWkhSb09qTXdKWDB1ZHkwek0zdDNhV1IwYURvek15VjlMbmN0TXpSN2QybGtkR2c2TXpRbGZTNTNMVFF3ZTNkcFpIUm9PalF3SlgwdWR5MDFNSHQzYVdSMGFEbzFNQ1Y5TG5jdE5qQjdkMmxrZEdnNk5qQWxmUzUzTFRjd2UzZHBaSFJvT2pjd0pYMHVkeTAzTlh0M2FXUjBhRG8zTlNWOUxuY3RPREI3ZDJsa2RHZzZPREFsZlM1M0xUa3dlM2RwWkhSb09qa3dKWDB1ZHkweE1EQjdkMmxrZEdnNk1UQXdKWDB1ZHkxMGFHbHlaSHQzYVdSMGFEb3pNeTR6TXpNek15VjlMbmN0ZEhkdkxYUm9hWEprYzN0M2FXUjBhRG8yTmk0Mk5qWTJOeVY5TG5jdFlYVjBiM3QzYVdSMGFEcGhkWFJ2ZlM1dmRtVnlabXh2ZHkxMmFYTnBZbXhsZTI5MlpYSm1iRzkzT25acGMybGliR1Y5TG05MlpYSm1iRzkzTFdocFpHUmxibnR2ZG1WeVpteHZkenBvYVdSa1pXNTlMbTkyWlhKbWJHOTNMWE5qY205c2JIdHZkbVZ5Wm14dmR6cHpZM0p2Ykd4OUxtOTJaWEptYkc5M0xXRjFkRzk3YjNabGNtWnNiM2M2WVhWMGIzMHViM1psY21ac2IzY3RlQzEyYVhOcFlteGxlMjkyWlhKbWJHOTNMWGc2ZG1semFXSnNaWDB1YjNabGNtWnNiM2N0ZUMxb2FXUmtaVzU3YjNabGNtWnNiM2N0ZURwb2FXUmtaVzU5TG05MlpYSm1iRzkzTFhndGMyTnliMnhzZTI5MlpYSm1iRzkzTFhnNmMyTnliMnhzZlM1dmRtVnlabXh2ZHkxNExXRjFkRzk3YjNabGNtWnNiM2N0ZURwaGRYUnZmUzV2ZG1WeVpteHZkeTE1TFhacGMybGliR1Y3YjNabGNtWnNiM2N0ZVRwMmFYTnBZbXhsZlM1dmRtVnlabXh2ZHkxNUxXaHBaR1JsYm50dmRtVnlabXh2ZHkxNU9taHBaR1JsYm4wdWIzWmxjbVpzYjNjdGVTMXpZM0p2Ykd4N2IzWmxjbVpzYjNjdGVUcHpZM0p2Ykd4OUxtOTJaWEptYkc5M0xYa3RZWFYwYjN0dmRtVnlabXh2ZHkxNU9tRjFkRzk5TG5OMFlYUnBZM3R3YjNOcGRHbHZianB6ZEdGMGFXTjlMbkpsYkdGMGFYWmxlM0J2YzJsMGFXOXVPbkpsYkdGMGFYWmxmUzVoWW5OdmJIVjBaWHR3YjNOcGRHbHZianBoWW5OdmJIVjBaWDB1Wm1sNFpXUjdjRzl6YVhScGIyNDZabWw0WldSOUxtOHRNVEF3ZTI5d1lXTnBkSGs2TVgwdWJ5MDVNSHR2Y0dGamFYUjVPaTQ1ZlM1dkxUZ3dlMjl3WVdOcGRIazZMamg5TG04dE56QjdiM0JoWTJsMGVUb3VOMzB1YnkwMk1IdHZjR0ZqYVhSNU9pNDJmUzV2TFRVd2UyOXdZV05wZEhrNkxqVjlMbTh0TkRCN2IzQmhZMmwwZVRvdU5IMHVieTB6TUh0dmNHRmphWFI1T2k0emZTNXZMVEl3ZTI5d1lXTnBkSGs2TGpKOUxtOHRNVEI3YjNCaFkybDBlVG91TVgwdWJ5MHdOWHR2Y0dGamFYUjVPaTR3TlgwdWJ5MHdNalY3YjNCaFkybDBlVG91TURJMWZTNXZMVEI3YjNCaFkybDBlVG93ZlM1eWIzUmhkR1V0TkRWN0xYZGxZbXRwZEMxMGNtRnVjMlp2Y20wNmNtOTBZWFJsS0RRMVpHVm5LVHQwY21GdWMyWnZjbTA2Y205MFlYUmxLRFExWkdWbktYMHVjbTkwWVhSbExUa3dleTEzWldKcmFYUXRkSEpoYm5ObWIzSnRPbkp2ZEdGMFpTZzVNR1JsWnlrN2RISmhibk5tYjNKdE9uSnZkR0YwWlNnNU1HUmxaeWw5TG5KdmRHRjBaUzB4TXpWN0xYZGxZbXRwZEMxMGNtRnVjMlp2Y20wNmNtOTBZWFJsS0RFek5XUmxaeWs3ZEhKaGJuTm1iM0p0T25KdmRHRjBaU2d4TXpWa1pXY3BmUzV5YjNSaGRHVXRNVGd3ZXkxM1pXSnJhWFF0ZEhKaGJuTm1iM0p0T25KdmRHRjBaU2d4T0RCa1pXY3BPM1J5WVc1elptOXliVHB5YjNSaGRHVW9NVGd3WkdWbktYMHVjbTkwWVhSbExUSXlOWHN0ZDJWaWEybDBMWFJ5WVc1elptOXliVHB5YjNSaGRHVW9NakkxWkdWbktUdDBjbUZ1YzJadmNtMDZjbTkwWVhSbEtESXlOV1JsWnlsOUxuSnZkR0YwWlMweU56QjdMWGRsWW10cGRDMTBjbUZ1YzJadmNtMDZjbTkwWVhSbEtESTNNR1JsWnlrN2RISmhibk5tYjNKdE9uSnZkR0YwWlNneU56QmtaV2NwZlM1eWIzUmhkR1V0TXpFMWV5MTNaV0pyYVhRdGRISmhibk5tYjNKdE9uSnZkR0YwWlNnek1UVmtaV2NwTzNSeVlXNXpabTl5YlRweWIzUmhkR1VvTXpFMVpHVm5LWDB1WW14aFkyc3RPVEI3WTI5c2IzSTZjbWRpWVNnd0xEQXNNQ3d1T1NsOUxtSnNZV05yTFRnd2UyTnZiRzl5T25KblltRW9NQ3d3TERBc0xqZ3BmUzVpYkdGamF5MDNNSHRqYjJ4dmNqcHlaMkpoS0RBc01Dd3dMQzQzS1gwdVlteGhZMnN0TmpCN1kyOXNiM0k2Y21kaVlTZ3dMREFzTUN3dU5pbDlMbUpzWVdOckxUVXdlMk52Ykc5eU9uSm5ZbUVvTUN3d0xEQXNMalVwZlM1aWJHRmpheTAwTUh0amIyeHZjanB5WjJKaEtEQXNNQ3d3TEM0MEtYMHVZbXhoWTJzdE16QjdZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VNeWw5TG1Kc1lXTnJMVEl3ZTJOdmJHOXlPbkpuWW1Fb01Dd3dMREFzTGpJcGZTNWliR0ZqYXkweE1IdGpiMnh2Y2pweVoySmhLREFzTUN3d0xDNHhLWDB1WW14aFkyc3RNRFY3WTI5c2IzSTZjbWRpWVNnd0xEQXNNQ3d1TURVcGZTNTNhR2wwWlMwNU1IdGpiMnh2Y2pwb2MyeGhLREFzTUNVc01UQXdKU3d1T1NsOUxuZG9hWFJsTFRnd2UyTnZiRzl5T21oemJHRW9NQ3d3SlN3eE1EQWxMQzQ0S1gwdWQyaHBkR1V0TnpCN1kyOXNiM0k2YUhOc1lTZ3dMREFsTERFd01DVXNMamNwZlM1M2FHbDBaUzAyTUh0amIyeHZjanBvYzJ4aEtEQXNNQ1VzTVRBd0pTd3VOaWw5TG5kb2FYUmxMVFV3ZTJOdmJHOXlPbWh6YkdFb01Dd3dKU3d4TURBbExDNDFLWDB1ZDJocGRHVXROREI3WTI5c2IzSTZhSE5zWVNnd0xEQWxMREV3TUNVc0xqUXBmUzUzYUdsMFpTMHpNSHRqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU15bDlMbmRvYVhSbExUSXdlMk52Ykc5eU9taHpiR0VvTUN3d0pTd3hNREFsTEM0eUtYMHVkMmhwZEdVdE1UQjdZMjlzYjNJNmFITnNZU2d3TERBbExERXdNQ1VzTGpFcGZTNWliR0ZqYTN0amIyeHZjam9qTURBd2ZTNXVaV0Z5TFdKc1lXTnJlMk52Ykc5eU9pTXhNVEY5TG1SaGNtc3RaM0poZVh0amIyeHZjam9qTXpNemZTNXRhV1F0WjNKaGVYdGpiMnh2Y2pvak5UVTFmUzVuY21GNWUyTnZiRzl5T2lNM056ZDlMbk5wYkhabGNudGpiMnh2Y2pvak9UazVmUzVzYVdkb2RDMXphV3gyWlhKN1kyOXNiM0k2STJGaFlYMHViVzl2YmkxbmNtRjVlMk52Ykc5eU9pTmpZMk45TG14cFoyaDBMV2R5WVhsN1kyOXNiM0k2STJWbFpYMHVibVZoY2kxM2FHbDBaWHRqYjJ4dmNqb2paalJtTkdZMGZTNTNhR2wwWlh0amIyeHZjam9qWm1abWZTNWtZWEpyTFhKbFpIdGpiMnh2Y2pvalpUY3dOREJtZlM1eVpXUjdZMjlzYjNJNkkyWm1OREV6Tm4wdWJHbG5hSFF0Y21Wa2UyTnZiRzl5T2lObVpqY3lOV045TG05eVlXNW5aWHRqYjJ4dmNqb2pabVkyTXpBd2ZTNW5iMnhrZTJOdmJHOXlPaU5tWm1JM01EQjlMbmxsYkd4dmQzdGpiMnh2Y2pwbmIyeGtmUzVzYVdkb2RDMTVaV3hzYjNkN1kyOXNiM0k2STJaaVpqRmhPWDB1Y0hWeWNHeGxlMk52Ykc5eU9pTTFaVEpqWVRWOUxteHBaMmgwTFhCMWNuQnNaWHRqYjJ4dmNqb2pZVFEyTTJZeWZTNWtZWEpyTFhCcGJtdDdZMjlzYjNJNkkyUTFNREE0Wm4wdWFHOTBMWEJwYm10N1kyOXNiM0k2STJabU5ERmlOSDB1Y0dsdWEzdGpiMnh2Y2pvalptWTRNR05qZlM1c2FXZG9kQzF3YVc1cmUyTnZiRzl5T2lObVptRXpaRGQ5TG1SaGNtc3RaM0psWlc1N1kyOXNiM0k2SXpFek56YzFNbjB1WjNKbFpXNTdZMjlzYjNJNkl6RTVZVGszTkgwdWJHbG5hSFF0WjNKbFpXNTdZMjlzYjNJNkl6bGxaV0pqWm4wdWJtRjJlWHRqYjJ4dmNqb2pNREF4WWpRMGZTNWtZWEpyTFdKc2RXVjdZMjlzYjNJNkl6QXdORFE1WlgwdVlteDFaWHRqYjJ4dmNqb2pNelUzWldSa2ZTNXNhV2RvZEMxaWJIVmxlMk52Ykc5eU9pTTVObU5qWm1aOUxteHBaMmgwWlhOMExXSnNkV1Y3WTI5c2IzSTZJMk5rWldObVpuMHVkMkZ6YUdWa0xXSnNkV1Y3WTI5c2IzSTZJMlkyWm1abVpYMHVkMkZ6YUdWa0xXZHlaV1Z1ZTJOdmJHOXlPaU5sT0daa1pqVjlMbmRoYzJobFpDMTVaV3hzYjNkN1kyOXNiM0k2STJabVptTmxZbjB1ZDJGemFHVmtMWEpsWkh0amIyeHZjam9qWm1aa1ptUm1mUzVqYjJ4dmNpMXBibWhsY21sMGUyTnZiRzl5T21sdWFHVnlhWFI5TG1KbkxXSnNZV05yTFRrd2UySmhZMnRuY205MWJtUXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VPU2w5TG1KbkxXSnNZV05yTFRnd2UySmhZMnRuY205MWJtUXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VPQ2w5TG1KbkxXSnNZV05yTFRjd2UySmhZMnRuY205MWJtUXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VOeWw5TG1KbkxXSnNZV05yTFRZd2UySmhZMnRuY205MWJtUXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VOaWw5TG1KbkxXSnNZV05yTFRVd2UySmhZMnRuY205MWJtUXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VOU2w5TG1KbkxXSnNZV05yTFRRd2UySmhZMnRuY205MWJtUXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VOQ2w5TG1KbkxXSnNZV05yTFRNd2UySmhZMnRuY205MWJtUXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VNeWw5TG1KbkxXSnNZV05yTFRJd2UySmhZMnRuY205MWJtUXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VNaWw5TG1KbkxXSnNZV05yTFRFd2UySmhZMnRuY205MWJtUXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VNU2w5TG1KbkxXSnNZV05yTFRBMWUySmhZMnRuY205MWJtUXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VNRFVwZlM1aVp5MTNhR2wwWlMwNU1IdGlZV05yWjNKdmRXNWtMV052Ykc5eU9taHpiR0VvTUN3d0pTd3hNREFsTEM0NUtYMHVZbWN0ZDJocGRHVXRPREI3WW1GamEyZHliM1Z1WkMxamIyeHZjanBvYzJ4aEtEQXNNQ1VzTVRBd0pTd3VPQ2w5TG1KbkxYZG9hWFJsTFRjd2UySmhZMnRuY205MWJtUXRZMjlzYjNJNmFITnNZU2d3TERBbExERXdNQ1VzTGpjcGZTNWlaeTEzYUdsMFpTMDJNSHRpWVdOclozSnZkVzVrTFdOdmJHOXlPbWh6YkdFb01Dd3dKU3d4TURBbExDNDJLWDB1WW1jdGQyaHBkR1V0TlRCN1ltRmphMmR5YjNWdVpDMWpiMnh2Y2pwb2MyeGhLREFzTUNVc01UQXdKU3d1TlNsOUxtSm5MWGRvYVhSbExUUXdlMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZhSE5zWVNnd0xEQWxMREV3TUNVc0xqUXBmUzVpWnkxM2FHbDBaUzB6TUh0aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T21oemJHRW9NQ3d3SlN3eE1EQWxMQzR6S1gwdVltY3RkMmhwZEdVdE1qQjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU1pbDlMbUpuTFhkb2FYUmxMVEV3ZTJKaFkydG5jbTkxYm1RdFkyOXNiM0k2YUhOc1lTZ3dMREFsTERFd01DVXNMakVwZlM1aVp5MWliR0ZqYTN0aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T2lNd01EQjlMbUpuTFc1bFlYSXRZbXhoWTJ0N1ltRmphMmR5YjNWdVpDMWpiMnh2Y2pvak1URXhmUzVpWnkxa1lYSnJMV2R5WVhsN1ltRmphMmR5YjNWdVpDMWpiMnh2Y2pvak16TXpmUzVpWnkxdGFXUXRaM0poZVh0aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T2lNMU5UVjlMbUpuTFdkeVlYbDdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2pOemMzZlM1aVp5MXphV3gyWlhKN1ltRmphMmR5YjNWdVpDMWpiMnh2Y2pvak9UazVmUzVpWnkxc2FXZG9kQzF6YVd4MlpYSjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2pZV0ZoZlM1aVp5MXRiMjl1TFdkeVlYbDdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2pZMk5qZlM1aVp5MXNhV2RvZEMxbmNtRjVlMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZJMlZsWlgwdVltY3RibVZoY2kxM2FHbDBaWHRpWVdOclozSnZkVzVrTFdOdmJHOXlPaU5tTkdZMFpqUjlMbUpuTFhkb2FYUmxlMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZJMlptWm4wdVltY3RkSEpoYm5Od1lYSmxiblI3WW1GamEyZHliM1Z1WkMxamIyeHZjanAwY21GdWMzQmhjbVZ1ZEgwdVltY3RaR0Z5YXkxeVpXUjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2paVGN3TkRCbWZTNWlaeTF5WldSN1ltRmphMmR5YjNWdVpDMWpiMnh2Y2pvalptWTBNVE0yZlM1aVp5MXNhV2RvZEMxeVpXUjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2pabVkzTWpWamZTNWlaeTF2Y21GdVoyVjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2pabVkyTXpBd2ZTNWlaeTFuYjJ4a2UySmhZMnRuY205MWJtUXRZMjlzYjNJNkkyWm1ZamN3TUgwdVltY3RlV1ZzYkc5M2UySmhZMnRuY205MWJtUXRZMjlzYjNJNloyOXNaSDB1WW1jdGJHbG5hSFF0ZVdWc2JHOTNlMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZJMlppWmpGaE9YMHVZbWN0Y0hWeWNHeGxlMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZJelZsTW1OaE5YMHVZbWN0YkdsbmFIUXRjSFZ5Y0d4bGUySmhZMnRuY205MWJtUXRZMjlzYjNJNkkyRTBOak5tTW4wdVltY3RaR0Z5YXkxd2FXNXJlMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZJMlExTURBNFpuMHVZbWN0YUc5MExYQnBibXQ3WW1GamEyZHliM1Z1WkMxamIyeHZjam9qWm1ZME1XSTBmUzVpWnkxd2FXNXJlMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZJMlptT0RCalkzMHVZbWN0YkdsbmFIUXRjR2x1YTN0aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T2lObVptRXpaRGQ5TG1KbkxXUmhjbXN0WjNKbFpXNTdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2pNVE0zTnpVeWZTNWlaeTFuY21WbGJudGlZV05yWjNKdmRXNWtMV052Ykc5eU9pTXhPV0U1TnpSOUxtSm5MV3hwWjJoMExXZHlaV1Z1ZTJKaFkydG5jbTkxYm1RdFkyOXNiM0k2SXpsbFpXSmpabjB1WW1jdGJtRjJlWHRpWVdOclozSnZkVzVrTFdOdmJHOXlPaU13TURGaU5EUjlMbUpuTFdSaGNtc3RZbXgxWlh0aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T2lNd01EUTBPV1Y5TG1KbkxXSnNkV1Y3WW1GamEyZHliM1Z1WkMxamIyeHZjam9qTXpVM1pXUmtmUzVpWnkxc2FXZG9kQzFpYkhWbGUySmhZMnRuY205MWJtUXRZMjlzYjNJNkl6azJZMk5tWm4wdVltY3RiR2xuYUhSbGMzUXRZbXgxWlh0aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T2lOalpHVmpabVo5TG1KbkxYZGhjMmhsWkMxaWJIVmxlMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZJMlkyWm1abVpYMHVZbWN0ZDJGemFHVmtMV2R5WldWdWUySmhZMnRuY205MWJtUXRZMjlzYjNJNkkyVTRabVJtTlgwdVltY3RkMkZ6YUdWa0xYbGxiR3h2ZDN0aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T2lObVptWmpaV0o5TG1KbkxYZGhjMmhsWkMxeVpXUjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2pabVprWm1SbWZTNWlaeTFwYm1obGNtbDBlMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZhVzVvWlhKcGRIMHVhRzkyWlhJdFlteGhZMnM2Wm05amRYTXNMbWh2ZG1WeUxXSnNZV05yT21odmRtVnllMk52Ykc5eU9pTXdNREI5TG1odmRtVnlMVzVsWVhJdFlteGhZMnM2Wm05amRYTXNMbWh2ZG1WeUxXNWxZWEl0WW14aFkyczZhRzkyWlhKN1kyOXNiM0k2SXpFeE1YMHVhRzkyWlhJdFpHRnlheTFuY21GNU9tWnZZM1Z6TEM1b2IzWmxjaTFrWVhKckxXZHlZWGs2YUc5MlpYSjdZMjlzYjNJNkl6TXpNMzB1YUc5MlpYSXRiV2xrTFdkeVlYazZabTlqZFhNc0xtaHZkbVZ5TFcxcFpDMW5jbUY1T21odmRtVnllMk52Ykc5eU9pTTFOVFY5TG1odmRtVnlMV2R5WVhrNlptOWpkWE1zTG1odmRtVnlMV2R5WVhrNmFHOTJaWEo3WTI5c2IzSTZJemMzTjMwdWFHOTJaWEl0YzJsc2RtVnlPbVp2WTNWekxDNW9iM1psY2kxemFXeDJaWEk2YUc5MlpYSjdZMjlzYjNJNkl6azVPWDB1YUc5MlpYSXRiR2xuYUhRdGMybHNkbVZ5T21adlkzVnpMQzVvYjNabGNpMXNhV2RvZEMxemFXeDJaWEk2YUc5MlpYSjdZMjlzYjNJNkkyRmhZWDB1YUc5MlpYSXRiVzl2YmkxbmNtRjVPbVp2WTNWekxDNW9iM1psY2kxdGIyOXVMV2R5WVhrNmFHOTJaWEo3WTI5c2IzSTZJMk5qWTMwdWFHOTJaWEl0YkdsbmFIUXRaM0poZVRwbWIyTjFjeXd1YUc5MlpYSXRiR2xuYUhRdFozSmhlVHBvYjNabGNudGpiMnh2Y2pvalpXVmxmUzVvYjNabGNpMXVaV0Z5TFhkb2FYUmxPbVp2WTNWekxDNW9iM1psY2kxdVpXRnlMWGRvYVhSbE9taHZkbVZ5ZTJOdmJHOXlPaU5tTkdZMFpqUjlMbWh2ZG1WeUxYZG9hWFJsT21adlkzVnpMQzVvYjNabGNpMTNhR2wwWlRwb2IzWmxjbnRqYjJ4dmNqb2pabVptZlM1b2IzWmxjaTFpYkdGamF5MDVNRHBtYjJOMWN5d3VhRzkyWlhJdFlteGhZMnN0T1RBNmFHOTJaWEo3WTI5c2IzSTZjbWRpWVNnd0xEQXNNQ3d1T1NsOUxtaHZkbVZ5TFdKc1lXTnJMVGd3T21adlkzVnpMQzVvYjNabGNpMWliR0ZqYXkwNE1EcG9iM1psY250amIyeHZjanB5WjJKaEtEQXNNQ3d3TEM0NEtYMHVhRzkyWlhJdFlteGhZMnN0TnpBNlptOWpkWE1zTG1odmRtVnlMV0pzWVdOckxUY3dPbWh2ZG1WeWUyTnZiRzl5T25KblltRW9NQ3d3TERBc0xqY3BmUzVvYjNabGNpMWliR0ZqYXkwMk1EcG1iMk4xY3l3dWFHOTJaWEl0WW14aFkyc3ROakE2YUc5MlpYSjdZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VOaWw5TG1odmRtVnlMV0pzWVdOckxUVXdPbVp2WTNWekxDNW9iM1psY2kxaWJHRmpheTAxTURwb2IzWmxjbnRqYjJ4dmNqcHlaMkpoS0RBc01Dd3dMQzQxS1gwdWFHOTJaWEl0WW14aFkyc3ROREE2Wm05amRYTXNMbWh2ZG1WeUxXSnNZV05yTFRRd09taHZkbVZ5ZTJOdmJHOXlPbkpuWW1Fb01Dd3dMREFzTGpRcGZTNW9iM1psY2kxaWJHRmpheTB6TURwbWIyTjFjeXd1YUc5MlpYSXRZbXhoWTJzdE16QTZhRzkyWlhKN1kyOXNiM0k2Y21kaVlTZ3dMREFzTUN3dU15bDlMbWh2ZG1WeUxXSnNZV05yTFRJd09tWnZZM1Z6TEM1b2IzWmxjaTFpYkdGamF5MHlNRHBvYjNabGNudGpiMnh2Y2pweVoySmhLREFzTUN3d0xDNHlLWDB1YUc5MlpYSXRZbXhoWTJzdE1UQTZabTlqZFhNc0xtaHZkbVZ5TFdKc1lXTnJMVEV3T21odmRtVnllMk52Ykc5eU9uSm5ZbUVvTUN3d0xEQXNMakVwZlM1b2IzWmxjaTEzYUdsMFpTMDVNRHBtYjJOMWN5d3VhRzkyWlhJdGQyaHBkR1V0T1RBNmFHOTJaWEo3WTI5c2IzSTZhSE5zWVNnd0xEQWxMREV3TUNVc0xqa3BmUzVvYjNabGNpMTNhR2wwWlMwNE1EcG1iMk4xY3l3dWFHOTJaWEl0ZDJocGRHVXRPREE2YUc5MlpYSjdZMjlzYjNJNmFITnNZU2d3TERBbExERXdNQ1VzTGpncGZTNW9iM1psY2kxM2FHbDBaUzAzTURwbWIyTjFjeXd1YUc5MlpYSXRkMmhwZEdVdE56QTZhRzkyWlhKN1kyOXNiM0k2YUhOc1lTZ3dMREFsTERFd01DVXNMamNwZlM1b2IzWmxjaTEzYUdsMFpTMDJNRHBtYjJOMWN5d3VhRzkyWlhJdGQyaHBkR1V0TmpBNmFHOTJaWEo3WTI5c2IzSTZhSE5zWVNnd0xEQWxMREV3TUNVc0xqWXBmUzVvYjNabGNpMTNhR2wwWlMwMU1EcG1iMk4xY3l3dWFHOTJaWEl0ZDJocGRHVXROVEE2YUc5MlpYSjdZMjlzYjNJNmFITnNZU2d3TERBbExERXdNQ1VzTGpVcGZTNW9iM1psY2kxM2FHbDBaUzAwTURwbWIyTjFjeXd1YUc5MlpYSXRkMmhwZEdVdE5EQTZhRzkyWlhKN1kyOXNiM0k2YUhOc1lTZ3dMREFsTERFd01DVXNMalFwZlM1b2IzWmxjaTEzYUdsMFpTMHpNRHBtYjJOMWN5d3VhRzkyWlhJdGQyaHBkR1V0TXpBNmFHOTJaWEo3WTI5c2IzSTZhSE5zWVNnd0xEQWxMREV3TUNVc0xqTXBmUzVvYjNabGNpMTNhR2wwWlMweU1EcG1iMk4xY3l3dWFHOTJaWEl0ZDJocGRHVXRNakE2YUc5MlpYSjdZMjlzYjNJNmFITnNZU2d3TERBbExERXdNQ1VzTGpJcGZTNW9iM1psY2kxM2FHbDBaUzB4TURwbWIyTjFjeXd1YUc5MlpYSXRkMmhwZEdVdE1UQTZhRzkyWlhKN1kyOXNiM0k2YUhOc1lTZ3dMREFsTERFd01DVXNMakVwZlM1b2IzWmxjaTFwYm1obGNtbDBPbVp2WTNWekxDNW9iM1psY2kxcGJtaGxjbWwwT21odmRtVnllMk52Ykc5eU9tbHVhR1Z5YVhSOUxtaHZkbVZ5TFdKbkxXSnNZV05yT21adlkzVnpMQzVvYjNabGNpMWlaeTFpYkdGamF6cG9iM1psY250aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T2lNd01EQjlMbWh2ZG1WeUxXSm5MVzVsWVhJdFlteGhZMnM2Wm05amRYTXNMbWh2ZG1WeUxXSm5MVzVsWVhJdFlteGhZMnM2YUc5MlpYSjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2pNVEV4ZlM1b2IzWmxjaTFpWnkxa1lYSnJMV2R5WVhrNlptOWpkWE1zTG1odmRtVnlMV0puTFdSaGNtc3RaM0poZVRwb2IzWmxjbnRpWVdOclozSnZkVzVrTFdOdmJHOXlPaU16TXpOOUxtaHZkbVZ5TFdKbkxXMXBaQzFuY21GNU9tWnZZM1Z6TEM1b2IzWmxjaTFpWnkxdGFXUXRaM0poZVRwb2IzWmxjbnRpWVdOclozSnZkVzVrTFdOdmJHOXlPaU0xTlRWOUxtaHZkbVZ5TFdKbkxXZHlZWGs2Wm05amRYTXNMbWh2ZG1WeUxXSm5MV2R5WVhrNmFHOTJaWEo3WW1GamEyZHliM1Z1WkMxamIyeHZjam9qTnpjM2ZTNW9iM1psY2kxaVp5MXphV3gyWlhJNlptOWpkWE1zTG1odmRtVnlMV0puTFhOcGJIWmxjanBvYjNabGNudGlZV05yWjNKdmRXNWtMV052Ykc5eU9pTTVPVGw5TG1odmRtVnlMV0puTFd4cFoyaDBMWE5wYkhabGNqcG1iMk4xY3l3dWFHOTJaWEl0WW1jdGJHbG5hSFF0YzJsc2RtVnlPbWh2ZG1WeWUySmhZMnRuY205MWJtUXRZMjlzYjNJNkkyRmhZWDB1YUc5MlpYSXRZbWN0Ylc5dmJpMW5jbUY1T21adlkzVnpMQzVvYjNabGNpMWlaeTF0YjI5dUxXZHlZWGs2YUc5MlpYSjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2pZMk5qZlM1b2IzWmxjaTFpWnkxc2FXZG9kQzFuY21GNU9tWnZZM1Z6TEM1b2IzWmxjaTFpWnkxc2FXZG9kQzFuY21GNU9taHZkbVZ5ZTJKaFkydG5jbTkxYm1RdFkyOXNiM0k2STJWbFpYMHVhRzkyWlhJdFltY3RibVZoY2kxM2FHbDBaVHBtYjJOMWN5d3VhRzkyWlhJdFltY3RibVZoY2kxM2FHbDBaVHBvYjNabGNudGlZV05yWjNKdmRXNWtMV052Ykc5eU9pTm1OR1kwWmpSOUxtaHZkbVZ5TFdKbkxYZG9hWFJsT21adlkzVnpMQzVvYjNabGNpMWlaeTEzYUdsMFpUcG9iM1psY250aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T2lObVptWjlMbWh2ZG1WeUxXSm5MWFJ5WVc1emNHRnlaVzUwT21adlkzVnpMQzVvYjNabGNpMWlaeTEwY21GdWMzQmhjbVZ1ZERwb2IzWmxjbnRpWVdOclozSnZkVzVrTFdOdmJHOXlPblJ5WVc1emNHRnlaVzUwZlM1b2IzWmxjaTFpWnkxaWJHRmpheTA1TURwbWIyTjFjeXd1YUc5MlpYSXRZbWN0WW14aFkyc3RPVEE2YUc5MlpYSjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqcHlaMkpoS0RBc01Dd3dMQzQ1S1gwdWFHOTJaWEl0WW1jdFlteGhZMnN0T0RBNlptOWpkWE1zTG1odmRtVnlMV0puTFdKc1lXTnJMVGd3T21odmRtVnllMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZjbWRpWVNnd0xEQXNNQ3d1T0NsOUxtaHZkbVZ5TFdKbkxXSnNZV05yTFRjd09tWnZZM1Z6TEM1b2IzWmxjaTFpWnkxaWJHRmpheTAzTURwb2IzWmxjbnRpWVdOclozSnZkVzVrTFdOdmJHOXlPbkpuWW1Fb01Dd3dMREFzTGpjcGZTNW9iM1psY2kxaVp5MWliR0ZqYXkwMk1EcG1iMk4xY3l3dWFHOTJaWEl0WW1jdFlteGhZMnN0TmpBNmFHOTJaWEo3WW1GamEyZHliM1Z1WkMxamIyeHZjanB5WjJKaEtEQXNNQ3d3TEM0MktYMHVhRzkyWlhJdFltY3RZbXhoWTJzdE5UQTZabTlqZFhNc0xtaHZkbVZ5TFdKbkxXSnNZV05yTFRVd09taHZkbVZ5ZTJKaFkydG5jbTkxYm1RdFkyOXNiM0k2Y21kaVlTZ3dMREFzTUN3dU5TbDlMbWh2ZG1WeUxXSm5MV0pzWVdOckxUUXdPbVp2WTNWekxDNW9iM1psY2kxaVp5MWliR0ZqYXkwME1EcG9iM1psY250aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T25KblltRW9NQ3d3TERBc0xqUXBmUzVvYjNabGNpMWlaeTFpYkdGamF5MHpNRHBtYjJOMWN5d3VhRzkyWlhJdFltY3RZbXhoWTJzdE16QTZhRzkyWlhKN1ltRmphMmR5YjNWdVpDMWpiMnh2Y2pweVoySmhLREFzTUN3d0xDNHpLWDB1YUc5MlpYSXRZbWN0WW14aFkyc3RNakE2Wm05amRYTXNMbWh2ZG1WeUxXSm5MV0pzWVdOckxUSXdPbWh2ZG1WeWUySmhZMnRuY205MWJtUXRZMjlzYjNJNmNtZGlZU2d3TERBc01Dd3VNaWw5TG1odmRtVnlMV0puTFdKc1lXTnJMVEV3T21adlkzVnpMQzVvYjNabGNpMWlaeTFpYkdGamF5MHhNRHBvYjNabGNudGlZV05yWjNKdmRXNWtMV052Ykc5eU9uSm5ZbUVvTUN3d0xEQXNMakVwZlM1b2IzWmxjaTFpWnkxM2FHbDBaUzA1TURwbWIyTjFjeXd1YUc5MlpYSXRZbWN0ZDJocGRHVXRPVEE2YUc5MlpYSjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqcG9jMnhoS0RBc01DVXNNVEF3SlN3dU9TbDlMbWh2ZG1WeUxXSm5MWGRvYVhSbExUZ3dPbVp2WTNWekxDNW9iM1psY2kxaVp5MTNhR2wwWlMwNE1EcG9iM1psY250aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T21oemJHRW9NQ3d3SlN3eE1EQWxMQzQ0S1gwdWFHOTJaWEl0WW1jdGQyaHBkR1V0TnpBNlptOWpkWE1zTG1odmRtVnlMV0puTFhkb2FYUmxMVGN3T21odmRtVnllMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZhSE5zWVNnd0xEQWxMREV3TUNVc0xqY3BmUzVvYjNabGNpMWlaeTEzYUdsMFpTMDJNRHBtYjJOMWN5d3VhRzkyWlhJdFltY3RkMmhwZEdVdE5qQTZhRzkyWlhKN1ltRmphMmR5YjNWdVpDMWpiMnh2Y2pwb2MyeGhLREFzTUNVc01UQXdKU3d1TmlsOUxtaHZkbVZ5TFdKbkxYZG9hWFJsTFRVd09tWnZZM1Z6TEM1b2IzWmxjaTFpWnkxM2FHbDBaUzAxTURwb2IzWmxjbnRpWVdOclozSnZkVzVrTFdOdmJHOXlPbWh6YkdFb01Dd3dKU3d4TURBbExDNDFLWDB1YUc5MlpYSXRZbWN0ZDJocGRHVXROREE2Wm05amRYTXNMbWh2ZG1WeUxXSm5MWGRvYVhSbExUUXdPbWh2ZG1WeWUySmhZMnRuY205MWJtUXRZMjlzYjNJNmFITnNZU2d3TERBbExERXdNQ1VzTGpRcGZTNW9iM1psY2kxaVp5MTNhR2wwWlMwek1EcG1iMk4xY3l3dWFHOTJaWEl0WW1jdGQyaHBkR1V0TXpBNmFHOTJaWEo3WW1GamEyZHliM1Z1WkMxamIyeHZjanBvYzJ4aEtEQXNNQ1VzTVRBd0pTd3VNeWw5TG1odmRtVnlMV0puTFhkb2FYUmxMVEl3T21adlkzVnpMQzVvYjNabGNpMWlaeTEzYUdsMFpTMHlNRHBvYjNabGNudGlZV05yWjNKdmRXNWtMV052Ykc5eU9taHpiR0VvTUN3d0pTd3hNREFsTEM0eUtYMHVhRzkyWlhJdFltY3RkMmhwZEdVdE1UQTZabTlqZFhNc0xtaHZkbVZ5TFdKbkxYZG9hWFJsTFRFd09taHZkbVZ5ZTJKaFkydG5jbTkxYm1RdFkyOXNiM0k2YUhOc1lTZ3dMREFsTERFd01DVXNMakVwZlM1b2IzWmxjaTFrWVhKckxYSmxaRHBtYjJOMWN5d3VhRzkyWlhJdFpHRnlheTF5WldRNmFHOTJaWEo3WTI5c2IzSTZJMlUzTURRd1puMHVhRzkyWlhJdGNtVmtPbVp2WTNWekxDNW9iM1psY2kxeVpXUTZhRzkyWlhKN1kyOXNiM0k2STJabU5ERXpObjB1YUc5MlpYSXRiR2xuYUhRdGNtVmtPbVp2WTNWekxDNW9iM1psY2kxc2FXZG9kQzF5WldRNmFHOTJaWEo3WTI5c2IzSTZJMlptTnpJMVkzMHVhRzkyWlhJdGIzSmhibWRsT21adlkzVnpMQzVvYjNabGNpMXZjbUZ1WjJVNmFHOTJaWEo3WTI5c2IzSTZJMlptTmpNd01IMHVhRzkyWlhJdFoyOXNaRHBtYjJOMWN5d3VhRzkyWlhJdFoyOXNaRHBvYjNabGNudGpiMnh2Y2pvalptWmlOekF3ZlM1b2IzWmxjaTE1Wld4c2IzYzZabTlqZFhNc0xtaHZkbVZ5TFhsbGJHeHZkenBvYjNabGNudGpiMnh2Y2pwbmIyeGtmUzVvYjNabGNpMXNhV2RvZEMxNVpXeHNiM2M2Wm05amRYTXNMbWh2ZG1WeUxXeHBaMmgwTFhsbGJHeHZkenBvYjNabGNudGpiMnh2Y2pvalptSm1NV0U1ZlM1b2IzWmxjaTF3ZFhKd2JHVTZabTlqZFhNc0xtaHZkbVZ5TFhCMWNuQnNaVHBvYjNabGNudGpiMnh2Y2pvak5XVXlZMkUxZlM1b2IzWmxjaTFzYVdkb2RDMXdkWEp3YkdVNlptOWpkWE1zTG1odmRtVnlMV3hwWjJoMExYQjFjbkJzWlRwb2IzWmxjbnRqYjJ4dmNqb2pZVFEyTTJZeWZTNW9iM1psY2kxa1lYSnJMWEJwYm1zNlptOWpkWE1zTG1odmRtVnlMV1JoY21zdGNHbHVhenBvYjNabGNudGpiMnh2Y2pvalpEVXdNRGhtZlM1b2IzWmxjaTFvYjNRdGNHbHVhenBtYjJOMWN5d3VhRzkyWlhJdGFHOTBMWEJwYm1zNmFHOTJaWEo3WTI5c2IzSTZJMlptTkRGaU5IMHVhRzkyWlhJdGNHbHVhenBtYjJOMWN5d3VhRzkyWlhJdGNHbHVhenBvYjNabGNudGpiMnh2Y2pvalptWTRNR05qZlM1b2IzWmxjaTFzYVdkb2RDMXdhVzVyT21adlkzVnpMQzVvYjNabGNpMXNhV2RvZEMxd2FXNXJPbWh2ZG1WeWUyTnZiRzl5T2lObVptRXpaRGQ5TG1odmRtVnlMV1JoY21zdFozSmxaVzQ2Wm05amRYTXNMbWh2ZG1WeUxXUmhjbXN0WjNKbFpXNDZhRzkyWlhKN1kyOXNiM0k2SXpFek56YzFNbjB1YUc5MlpYSXRaM0psWlc0NlptOWpkWE1zTG1odmRtVnlMV2R5WldWdU9taHZkbVZ5ZTJOdmJHOXlPaU14T1dFNU56UjlMbWh2ZG1WeUxXeHBaMmgwTFdkeVpXVnVPbVp2WTNWekxDNW9iM1psY2kxc2FXZG9kQzFuY21WbGJqcG9iM1psY250amIyeHZjam9qT1dWbFltTm1mUzVvYjNabGNpMXVZWFo1T21adlkzVnpMQzVvYjNabGNpMXVZWFo1T21odmRtVnllMk52Ykc5eU9pTXdNREZpTkRSOUxtaHZkbVZ5TFdSaGNtc3RZbXgxWlRwbWIyTjFjeXd1YUc5MlpYSXRaR0Z5YXkxaWJIVmxPbWh2ZG1WeWUyTnZiRzl5T2lNd01EUTBPV1Y5TG1odmRtVnlMV0pzZFdVNlptOWpkWE1zTG1odmRtVnlMV0pzZFdVNmFHOTJaWEo3WTI5c2IzSTZJek0xTjJWa1pIMHVhRzkyWlhJdGJHbG5hSFF0WW14MVpUcG1iMk4xY3l3dWFHOTJaWEl0YkdsbmFIUXRZbXgxWlRwb2IzWmxjbnRqYjJ4dmNqb2pPVFpqWTJabWZTNW9iM1psY2kxc2FXZG9kR1Z6ZEMxaWJIVmxPbVp2WTNWekxDNW9iM1psY2kxc2FXZG9kR1Z6ZEMxaWJIVmxPbWh2ZG1WeWUyTnZiRzl5T2lOalpHVmpabVo5TG1odmRtVnlMWGRoYzJobFpDMWliSFZsT21adlkzVnpMQzVvYjNabGNpMTNZWE5vWldRdFlteDFaVHBvYjNabGNudGpiMnh2Y2pvalpqWm1abVpsZlM1b2IzWmxjaTEzWVhOb1pXUXRaM0psWlc0NlptOWpkWE1zTG1odmRtVnlMWGRoYzJobFpDMW5jbVZsYmpwb2IzWmxjbnRqYjJ4dmNqb2paVGhtWkdZMWZTNW9iM1psY2kxM1lYTm9aV1F0ZVdWc2JHOTNPbVp2WTNWekxDNW9iM1psY2kxM1lYTm9aV1F0ZVdWc2JHOTNPbWh2ZG1WeWUyTnZiRzl5T2lObVptWmpaV0o5TG1odmRtVnlMWGRoYzJobFpDMXlaV1E2Wm05amRYTXNMbWh2ZG1WeUxYZGhjMmhsWkMxeVpXUTZhRzkyWlhKN1kyOXNiM0k2STJabVpHWmtabjB1YUc5MlpYSXRZbWN0WkdGeWF5MXlaV1E2Wm05amRYTXNMbWh2ZG1WeUxXSm5MV1JoY21zdGNtVmtPbWh2ZG1WeWUySmhZMnRuY205MWJtUXRZMjlzYjNJNkkyVTNNRFF3Wm4wdWFHOTJaWEl0WW1jdGNtVmtPbVp2WTNWekxDNW9iM1psY2kxaVp5MXlaV1E2YUc5MlpYSjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2pabVkwTVRNMmZTNW9iM1psY2kxaVp5MXNhV2RvZEMxeVpXUTZabTlqZFhNc0xtaHZkbVZ5TFdKbkxXeHBaMmgwTFhKbFpEcG9iM1psY250aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T2lObVpqY3lOV045TG1odmRtVnlMV0puTFc5eVlXNW5aVHBtYjJOMWN5d3VhRzkyWlhJdFltY3RiM0poYm1kbE9taHZkbVZ5ZTJKaFkydG5jbTkxYm1RdFkyOXNiM0k2STJabU5qTXdNSDB1YUc5MlpYSXRZbWN0WjI5c1pEcG1iMk4xY3l3dWFHOTJaWEl0WW1jdFoyOXNaRHBvYjNabGNudGlZV05yWjNKdmRXNWtMV052Ykc5eU9pTm1abUkzTURCOUxtaHZkbVZ5TFdKbkxYbGxiR3h2ZHpwbWIyTjFjeXd1YUc5MlpYSXRZbWN0ZVdWc2JHOTNPbWh2ZG1WeWUySmhZMnRuY205MWJtUXRZMjlzYjNJNloyOXNaSDB1YUc5MlpYSXRZbWN0YkdsbmFIUXRlV1ZzYkc5M09tWnZZM1Z6TEM1b2IzWmxjaTFpWnkxc2FXZG9kQzE1Wld4c2IzYzZhRzkyWlhKN1ltRmphMmR5YjNWdVpDMWpiMnh2Y2pvalptSm1NV0U1ZlM1b2IzWmxjaTFpWnkxd2RYSndiR1U2Wm05amRYTXNMbWh2ZG1WeUxXSm5MWEIxY25Cc1pUcG9iM1psY250aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T2lNMVpUSmpZVFY5TG1odmRtVnlMV0puTFd4cFoyaDBMWEIxY25Cc1pUcG1iMk4xY3l3dWFHOTJaWEl0WW1jdGJHbG5hSFF0Y0hWeWNHeGxPbWh2ZG1WeWUySmhZMnRuY205MWJtUXRZMjlzYjNJNkkyRTBOak5tTW4wdWFHOTJaWEl0WW1jdFpHRnlheTF3YVc1ck9tWnZZM1Z6TEM1b2IzWmxjaTFpWnkxa1lYSnJMWEJwYm1zNmFHOTJaWEo3WW1GamEyZHliM1Z1WkMxamIyeHZjam9qWkRVd01EaG1mUzVvYjNabGNpMWlaeTFvYjNRdGNHbHVhenBtYjJOMWN5d3VhRzkyWlhJdFltY3RhRzkwTFhCcGJtczZhRzkyWlhKN1ltRmphMmR5YjNWdVpDMWpiMnh2Y2pvalptWTBNV0kwZlM1b2IzWmxjaTFpWnkxd2FXNXJPbVp2WTNWekxDNW9iM1psY2kxaVp5MXdhVzVyT21odmRtVnllMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZJMlptT0RCalkzMHVhRzkyWlhJdFltY3RiR2xuYUhRdGNHbHVhenBtYjJOMWN5d3VhRzkyWlhJdFltY3RiR2xuYUhRdGNHbHVhenBvYjNabGNudGlZV05yWjNKdmRXNWtMV052Ykc5eU9pTm1abUV6WkRkOUxtaHZkbVZ5TFdKbkxXUmhjbXN0WjNKbFpXNDZabTlqZFhNc0xtaHZkbVZ5TFdKbkxXUmhjbXN0WjNKbFpXNDZhRzkyWlhKN1ltRmphMmR5YjNWdVpDMWpiMnh2Y2pvak1UTTNOelV5ZlM1b2IzWmxjaTFpWnkxbmNtVmxianBtYjJOMWN5d3VhRzkyWlhJdFltY3RaM0psWlc0NmFHOTJaWEo3WW1GamEyZHliM1Z1WkMxamIyeHZjam9qTVRsaE9UYzBmUzVvYjNabGNpMWlaeTFzYVdkb2RDMW5jbVZsYmpwbWIyTjFjeXd1YUc5MlpYSXRZbWN0YkdsbmFIUXRaM0psWlc0NmFHOTJaWEo3WW1GamEyZHliM1Z1WkMxamIyeHZjam9qT1dWbFltTm1mUzVvYjNabGNpMWlaeTF1WVhaNU9tWnZZM1Z6TEM1b2IzWmxjaTFpWnkxdVlYWjVPbWh2ZG1WeWUySmhZMnRuY205MWJtUXRZMjlzYjNJNkl6QXdNV0kwTkgwdWFHOTJaWEl0WW1jdFpHRnlheTFpYkhWbE9tWnZZM1Z6TEM1b2IzWmxjaTFpWnkxa1lYSnJMV0pzZFdVNmFHOTJaWEo3WW1GamEyZHliM1Z1WkMxamIyeHZjam9qTURBME5EbGxmUzVvYjNabGNpMWlaeTFpYkhWbE9tWnZZM1Z6TEM1b2IzWmxjaTFpWnkxaWJIVmxPbWh2ZG1WeWUySmhZMnRuY205MWJtUXRZMjlzYjNJNkl6TTFOMlZrWkgwdWFHOTJaWEl0WW1jdGJHbG5hSFF0WW14MVpUcG1iMk4xY3l3dWFHOTJaWEl0WW1jdGJHbG5hSFF0WW14MVpUcG9iM1psY250aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T2lNNU5tTmpabVo5TG1odmRtVnlMV0puTFd4cFoyaDBaWE4wTFdKc2RXVTZabTlqZFhNc0xtaHZkbVZ5TFdKbkxXeHBaMmgwWlhOMExXSnNkV1U2YUc5MlpYSjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2pZMlJsWTJabWZTNW9iM1psY2kxaVp5MTNZWE5vWldRdFlteDFaVHBtYjJOMWN5d3VhRzkyWlhJdFltY3RkMkZ6YUdWa0xXSnNkV1U2YUc5MlpYSjdZbUZqYTJkeWIzVnVaQzFqYjJ4dmNqb2paalptWm1abGZTNW9iM1psY2kxaVp5MTNZWE5vWldRdFozSmxaVzQ2Wm05amRYTXNMbWh2ZG1WeUxXSm5MWGRoYzJobFpDMW5jbVZsYmpwb2IzWmxjbnRpWVdOclozSnZkVzVrTFdOdmJHOXlPaU5sT0daa1pqVjlMbWh2ZG1WeUxXSm5MWGRoYzJobFpDMTVaV3hzYjNjNlptOWpkWE1zTG1odmRtVnlMV0puTFhkaGMyaGxaQzE1Wld4c2IzYzZhRzkyWlhKN1ltRmphMmR5YjNWdVpDMWpiMnh2Y2pvalptWm1ZMlZpZlM1b2IzWmxjaTFpWnkxM1lYTm9aV1F0Y21Wa09tWnZZM1Z6TEM1b2IzWmxjaTFpWnkxM1lYTm9aV1F0Y21Wa09taHZkbVZ5ZTJKaFkydG5jbTkxYm1RdFkyOXNiM0k2STJabVpHWmtabjB1YUc5MlpYSXRZbWN0YVc1b1pYSnBkRHBtYjJOMWN5d3VhRzkyWlhJdFltY3RhVzVvWlhKcGREcG9iM1psY250aVlXTnJaM0p2ZFc1a0xXTnZiRzl5T21sdWFHVnlhWFI5TG5CaE1IdHdZV1JrYVc1bk9qQjlMbkJoTVh0d1lXUmthVzVuT2k0eU5YSmxiWDB1Y0dFeWUzQmhaR1JwYm1jNkxqVnlaVzE5TG5CaE0zdHdZV1JrYVc1bk9qRnlaVzE5TG5CaE5IdHdZV1JrYVc1bk9qSnlaVzE5TG5CaE5YdHdZV1JrYVc1bk9qUnlaVzE5TG5CaE5udHdZV1JrYVc1bk9qaHlaVzE5TG5CaE4zdHdZV1JrYVc1bk9qRTJjbVZ0ZlM1d2JEQjdjR0ZrWkdsdVp5MXNaV1owT2pCOUxuQnNNWHR3WVdSa2FXNW5MV3hsWm5RNkxqSTFjbVZ0ZlM1d2JESjdjR0ZrWkdsdVp5MXNaV1owT2k0MWNtVnRmUzV3YkRON2NHRmtaR2x1Wnkxc1pXWjBPakZ5WlcxOUxuQnNOSHR3WVdSa2FXNW5MV3hsWm5RNk1uSmxiWDB1Y0d3MWUzQmhaR1JwYm1jdGJHVm1kRG8wY21WdGZTNXdiRFo3Y0dGa1pHbHVaeTFzWldaME9qaHlaVzE5TG5Cc04zdHdZV1JrYVc1bkxXeGxablE2TVRaeVpXMTlMbkJ5TUh0d1lXUmthVzVuTFhKcFoyaDBPakI5TG5CeU1YdHdZV1JrYVc1bkxYSnBaMmgwT2k0eU5YSmxiWDB1Y0hJeWUzQmhaR1JwYm1jdGNtbG5hSFE2TGpWeVpXMTlMbkJ5TTN0d1lXUmthVzVuTFhKcFoyaDBPakZ5WlcxOUxuQnlOSHR3WVdSa2FXNW5MWEpwWjJoME9qSnlaVzE5TG5CeU5YdHdZV1JrYVc1bkxYSnBaMmgwT2pSeVpXMTlMbkJ5Tm50d1lXUmthVzVuTFhKcFoyaDBPamh5WlcxOUxuQnlOM3R3WVdSa2FXNW5MWEpwWjJoME9qRTJjbVZ0ZlM1d1lqQjdjR0ZrWkdsdVp5MWliM1IwYjIwNk1IMHVjR0l4ZTNCaFpHUnBibWN0WW05MGRHOXRPaTR5TlhKbGJYMHVjR0l5ZTNCaFpHUnBibWN0WW05MGRHOXRPaTQxY21WdGZTNXdZak43Y0dGa1pHbHVaeTFpYjNSMGIyMDZNWEpsYlgwdWNHSTBlM0JoWkdScGJtY3RZbTkwZEc5dE9qSnlaVzE5TG5CaU5YdHdZV1JrYVc1bkxXSnZkSFJ2YlRvMGNtVnRmUzV3WWpaN2NHRmtaR2x1WnkxaWIzUjBiMjA2T0hKbGJYMHVjR0kzZTNCaFpHUnBibWN0WW05MGRHOXRPakUyY21WdGZTNXdkREI3Y0dGa1pHbHVaeTEwYjNBNk1IMHVjSFF4ZTNCaFpHUnBibWN0ZEc5d09pNHlOWEpsYlgwdWNIUXllM0JoWkdScGJtY3RkRzl3T2k0MWNtVnRmUzV3ZERON2NHRmtaR2x1WnkxMGIzQTZNWEpsYlgwdWNIUTBlM0JoWkdScGJtY3RkRzl3T2pKeVpXMTlMbkIwTlh0d1lXUmthVzVuTFhSdmNEbzBjbVZ0ZlM1d2REWjdjR0ZrWkdsdVp5MTBiM0E2T0hKbGJYMHVjSFEzZTNCaFpHUnBibWN0ZEc5d09qRTJjbVZ0ZlM1d2RqQjdjR0ZrWkdsdVp5MTBiM0E2TUR0d1lXUmthVzVuTFdKdmRIUnZiVG93ZlM1d2RqRjdjR0ZrWkdsdVp5MTBiM0E2TGpJMWNtVnRPM0JoWkdScGJtY3RZbTkwZEc5dE9pNHlOWEpsYlgwdWNIWXllM0JoWkdScGJtY3RkRzl3T2k0MWNtVnRPM0JoWkdScGJtY3RZbTkwZEc5dE9pNDFjbVZ0ZlM1d2RqTjdjR0ZrWkdsdVp5MTBiM0E2TVhKbGJUdHdZV1JrYVc1bkxXSnZkSFJ2YlRveGNtVnRmUzV3ZGpSN2NHRmtaR2x1WnkxMGIzQTZNbkpsYlR0d1lXUmthVzVuTFdKdmRIUnZiVG95Y21WdGZTNXdkalY3Y0dGa1pHbHVaeTEwYjNBNk5ISmxiVHR3WVdSa2FXNW5MV0p2ZEhSdmJUbzBjbVZ0ZlM1d2RqWjdjR0ZrWkdsdVp5MTBiM0E2T0hKbGJUdHdZV1JrYVc1bkxXSnZkSFJ2YlRvNGNtVnRmUzV3ZGpkN2NHRmtaR2x1WnkxMGIzQTZNVFp5WlcwN2NHRmtaR2x1WnkxaWIzUjBiMjA2TVRaeVpXMTlMbkJvTUh0d1lXUmthVzVuTFd4bFpuUTZNRHR3WVdSa2FXNW5MWEpwWjJoME9qQjlMbkJvTVh0d1lXUmthVzVuTFd4bFpuUTZMakkxY21WdE8zQmhaR1JwYm1jdGNtbG5hSFE2TGpJMWNtVnRmUzV3YURKN2NHRmtaR2x1Wnkxc1pXWjBPaTQxY21WdE8zQmhaR1JwYm1jdGNtbG5hSFE2TGpWeVpXMTlMbkJvTTN0d1lXUmthVzVuTFd4bFpuUTZNWEpsYlR0d1lXUmthVzVuTFhKcFoyaDBPakZ5WlcxOUxuQm9OSHR3WVdSa2FXNW5MV3hsWm5RNk1uSmxiVHR3WVdSa2FXNW5MWEpwWjJoME9qSnlaVzE5TG5Cb05YdHdZV1JrYVc1bkxXeGxablE2TkhKbGJUdHdZV1JrYVc1bkxYSnBaMmgwT2pSeVpXMTlMbkJvTm50d1lXUmthVzVuTFd4bFpuUTZPSEpsYlR0d1lXUmthVzVuTFhKcFoyaDBPamh5WlcxOUxuQm9OM3R3WVdSa2FXNW5MV3hsWm5RNk1UWnlaVzA3Y0dGa1pHbHVaeTF5YVdkb2REb3hObkpsYlgwdWJXRXdlMjFoY21kcGJqb3dmUzV0WVRGN2JXRnlaMmx1T2k0eU5YSmxiWDB1YldFeWUyMWhjbWRwYmpvdU5YSmxiWDB1YldFemUyMWhjbWRwYmpveGNtVnRmUzV0WVRSN2JXRnlaMmx1T2pKeVpXMTlMbTFoTlh0dFlYSm5hVzQ2TkhKbGJYMHViV0UyZTIxaGNtZHBiam80Y21WdGZTNXRZVGQ3YldGeVoybHVPakUyY21WdGZTNXRiREI3YldGeVoybHVMV3hsWm5RNk1IMHViV3d4ZTIxaGNtZHBiaTFzWldaME9pNHlOWEpsYlgwdWJXd3llMjFoY21kcGJpMXNaV1owT2k0MWNtVnRmUzV0YkRON2JXRnlaMmx1TFd4bFpuUTZNWEpsYlgwdWJXdzBlMjFoY21kcGJpMXNaV1owT2pKeVpXMTlMbTFzTlh0dFlYSm5hVzR0YkdWbWREbzBjbVZ0ZlM1dGJEWjdiV0Z5WjJsdUxXeGxablE2T0hKbGJYMHViV3czZTIxaGNtZHBiaTFzWldaME9qRTJjbVZ0ZlM1dGNqQjdiV0Z5WjJsdUxYSnBaMmgwT2pCOUxtMXlNWHR0WVhKbmFXNHRjbWxuYUhRNkxqSTFjbVZ0ZlM1dGNqSjdiV0Z5WjJsdUxYSnBaMmgwT2k0MWNtVnRmUzV0Y2pON2JXRnlaMmx1TFhKcFoyaDBPakZ5WlcxOUxtMXlOSHR0WVhKbmFXNHRjbWxuYUhRNk1uSmxiWDB1YlhJMWUyMWhjbWRwYmkxeWFXZG9kRG8wY21WdGZTNXRjalo3YldGeVoybHVMWEpwWjJoME9qaHlaVzE5TG0xeU4zdHRZWEpuYVc0dGNtbG5hSFE2TVRaeVpXMTlMbTFpTUh0dFlYSm5hVzR0WW05MGRHOXRPakI5TG0xaU1YdHRZWEpuYVc0dFltOTBkRzl0T2k0eU5YSmxiWDB1YldJeWUyMWhjbWRwYmkxaWIzUjBiMjA2TGpWeVpXMTlMbTFpTTN0dFlYSm5hVzR0WW05MGRHOXRPakZ5WlcxOUxtMWlOSHR0WVhKbmFXNHRZbTkwZEc5dE9qSnlaVzE5TG0xaU5YdHRZWEpuYVc0dFltOTBkRzl0T2pSeVpXMTlMbTFpTm50dFlYSm5hVzR0WW05MGRHOXRPamh5WlcxOUxtMWlOM3R0WVhKbmFXNHRZbTkwZEc5dE9qRTJjbVZ0ZlM1dGREQjdiV0Z5WjJsdUxYUnZjRG93ZlM1dGRERjdiV0Z5WjJsdUxYUnZjRG91TWpWeVpXMTlMbTEwTW50dFlYSm5hVzR0ZEc5d09pNDFjbVZ0ZlM1dGRETjdiV0Z5WjJsdUxYUnZjRG94Y21WdGZTNXRkRFI3YldGeVoybHVMWFJ2Y0RveWNtVnRmUzV0ZERWN2JXRnlaMmx1TFhSdmNEbzBjbVZ0ZlM1dGREWjdiV0Z5WjJsdUxYUnZjRG80Y21WdGZTNXRkRGQ3YldGeVoybHVMWFJ2Y0RveE5uSmxiWDB1YlhZd2UyMWhjbWRwYmkxMGIzQTZNRHR0WVhKbmFXNHRZbTkwZEc5dE9qQjlMbTEyTVh0dFlYSm5hVzR0ZEc5d09pNHlOWEpsYlR0dFlYSm5hVzR0WW05MGRHOXRPaTR5TlhKbGJYMHViWFl5ZTIxaGNtZHBiaTEwYjNBNkxqVnlaVzA3YldGeVoybHVMV0p2ZEhSdmJUb3VOWEpsYlgwdWJYWXplMjFoY21kcGJpMTBiM0E2TVhKbGJUdHRZWEpuYVc0dFltOTBkRzl0T2pGeVpXMTlMbTEyTkh0dFlYSm5hVzR0ZEc5d09qSnlaVzA3YldGeVoybHVMV0p2ZEhSdmJUb3ljbVZ0ZlM1dGRqVjdiV0Z5WjJsdUxYUnZjRG8wY21WdE8yMWhjbWRwYmkxaWIzUjBiMjA2TkhKbGJYMHViWFkyZTIxaGNtZHBiaTEwYjNBNk9ISmxiVHR0WVhKbmFXNHRZbTkwZEc5dE9qaHlaVzE5TG0xMk4zdHRZWEpuYVc0dGRHOXdPakUyY21WdE8yMWhjbWRwYmkxaWIzUjBiMjA2TVRaeVpXMTlMbTFvTUh0dFlYSm5hVzR0YkdWbWREb3dPMjFoY21kcGJpMXlhV2RvZERvd2ZTNXRhREY3YldGeVoybHVMV3hsWm5RNkxqSTFjbVZ0TzIxaGNtZHBiaTF5YVdkb2REb3VNalZ5WlcxOUxtMW9NbnR0WVhKbmFXNHRiR1ZtZERvdU5YSmxiVHR0WVhKbmFXNHRjbWxuYUhRNkxqVnlaVzE5TG0xb00zdHRZWEpuYVc0dGJHVm1kRG94Y21WdE8yMWhjbWRwYmkxeWFXZG9kRG94Y21WdGZTNXRhRFI3YldGeVoybHVMV3hsWm5RNk1uSmxiVHR0WVhKbmFXNHRjbWxuYUhRNk1uSmxiWDB1YldnMWUyMWhjbWRwYmkxc1pXWjBPalJ5WlcwN2JXRnlaMmx1TFhKcFoyaDBPalJ5WlcxOUxtMW9ObnR0WVhKbmFXNHRiR1ZtZERvNGNtVnRPMjFoY21kcGJpMXlhV2RvZERvNGNtVnRmUzV0YURkN2JXRnlaMmx1TFd4bFpuUTZNVFp5WlcwN2JXRnlaMmx1TFhKcFoyaDBPakUyY21WdGZTNXVZVEY3YldGeVoybHVPaTB1TWpWeVpXMTlMbTVoTW50dFlYSm5hVzQ2TFM0MWNtVnRmUzV1WVRON2JXRnlaMmx1T2kweGNtVnRmUzV1WVRSN2JXRnlaMmx1T2kweWNtVnRmUzV1WVRWN2JXRnlaMmx1T2kwMGNtVnRmUzV1WVRaN2JXRnlaMmx1T2kwNGNtVnRmUzV1WVRkN2JXRnlaMmx1T2kweE5uSmxiWDB1Ym13eGUyMWhjbWRwYmkxc1pXWjBPaTB1TWpWeVpXMTlMbTVzTW50dFlYSm5hVzR0YkdWbWREb3RMalZ5WlcxOUxtNXNNM3R0WVhKbmFXNHRiR1ZtZERvdE1YSmxiWDB1Ym13MGUyMWhjbWRwYmkxc1pXWjBPaTB5Y21WdGZTNXViRFY3YldGeVoybHVMV3hsWm5RNkxUUnlaVzE5TG01c05udHRZWEpuYVc0dGJHVm1kRG90T0hKbGJYMHVibXczZTIxaGNtZHBiaTFzWldaME9pMHhObkpsYlgwdWJuSXhlMjFoY21kcGJpMXlhV2RvZERvdExqSTFjbVZ0ZlM1dWNqSjdiV0Z5WjJsdUxYSnBaMmgwT2kwdU5YSmxiWDB1Ym5JemUyMWhjbWRwYmkxeWFXZG9kRG90TVhKbGJYMHVibkkwZTIxaGNtZHBiaTF5YVdkb2REb3RNbkpsYlgwdWJuSTFlMjFoY21kcGJpMXlhV2RvZERvdE5ISmxiWDB1Ym5JMmUyMWhjbWRwYmkxeWFXZG9kRG90T0hKbGJYMHVibkkzZTIxaGNtZHBiaTF5YVdkb2REb3RNVFp5WlcxOUxtNWlNWHR0WVhKbmFXNHRZbTkwZEc5dE9pMHVNalZ5WlcxOUxtNWlNbnR0WVhKbmFXNHRZbTkwZEc5dE9pMHVOWEpsYlgwdWJtSXplMjFoY21kcGJpMWliM1IwYjIwNkxURnlaVzE5TG01aU5IdHRZWEpuYVc0dFltOTBkRzl0T2kweWNtVnRmUzV1WWpWN2JXRnlaMmx1TFdKdmRIUnZiVG90TkhKbGJYMHVibUkyZTIxaGNtZHBiaTFpYjNSMGIyMDZMVGh5WlcxOUxtNWlOM3R0WVhKbmFXNHRZbTkwZEc5dE9pMHhObkpsYlgwdWJuUXhlMjFoY21kcGJpMTBiM0E2TFM0eU5YSmxiWDB1Ym5ReWUyMWhjbWRwYmkxMGIzQTZMUzQxY21WdGZTNXVkRE43YldGeVoybHVMWFJ2Y0RvdE1YSmxiWDB1Ym5RMGUyMWhjbWRwYmkxMGIzQTZMVEp5WlcxOUxtNTBOWHR0WVhKbmFXNHRkRzl3T2kwMGNtVnRmUzV1ZERaN2JXRnlaMmx1TFhSdmNEb3RPSEpsYlgwdWJuUTNlMjFoY21kcGJpMTBiM0E2TFRFMmNtVnRmUzVqYjJ4c1lYQnpaWHRpYjNKa1pYSXRZMjlzYkdGd2MyVTZZMjlzYkdGd2MyVTdZbTl5WkdWeUxYTndZV05wYm1jNk1IMHVjM1J5YVhCbFpDMHRiR2xuYUhRdGMybHNkbVZ5T201MGFDMWphR2xzWkNodlpHUXBlMkpoWTJ0bmNtOTFibVF0WTI5c2IzSTZJMkZoWVgwdWMzUnlhWEJsWkMwdGJXOXZiaTFuY21GNU9tNTBhQzFqYUdsc1pDaHZaR1FwZTJKaFkydG5jbTkxYm1RdFkyOXNiM0k2STJOalkzMHVjM1J5YVhCbFpDMHRiR2xuYUhRdFozSmhlVHB1ZEdndFkyaHBiR1FvYjJSa0tYdGlZV05yWjNKdmRXNWtMV052Ykc5eU9pTmxaV1Y5TG5OMGNtbHdaV1F0TFc1bFlYSXRkMmhwZEdVNmJuUm9MV05vYVd4a0tHOWtaQ2w3WW1GamEyZHliM1Z1WkMxamIyeHZjam9qWmpSbU5HWTBmUzV6ZEhKcGNHVXRiR2xuYUhRNmJuUm9MV05vYVd4a0tHOWtaQ2w3WW1GamEyZHliM1Z1WkMxamIyeHZjanBvYzJ4aEtEQXNNQ1VzTVRBd0pTd3VNU2w5TG5OMGNtbHdaUzFrWVhKck9tNTBhQzFqYUdsc1pDaHZaR1FwZTJKaFkydG5jbTkxYm1RdFkyOXNiM0k2Y21kaVlTZ3dMREFzTUN3dU1TbDlMbk4wY21sclpYdDBaWGgwTFdSbFkyOXlZWFJwYjI0NmJHbHVaUzEwYUhKdmRXZG9mUzUxYm1SbGNteHBibVY3ZEdWNGRDMWtaV052Y21GMGFXOXVPblZ1WkdWeWJHbHVaWDB1Ym04dGRXNWtaWEpzYVc1bGUzUmxlSFF0WkdWamIzSmhkR2x2YmpwdWIyNWxmUzUwYkh0MFpYaDBMV0ZzYVdkdU9teGxablI5TG5SeWUzUmxlSFF0WVd4cFoyNDZjbWxuYUhSOUxuUmplM1JsZUhRdFlXeHBaMjQ2WTJWdWRHVnlmUzUwYW50MFpYaDBMV0ZzYVdkdU9tcDFjM1JwWm5sOUxuUjBZM3QwWlhoMExYUnlZVzV6Wm05eWJUcGpZWEJwZEdGc2FYcGxmUzUwZEd4N2RHVjRkQzEwY21GdWMyWnZjbTA2Ykc5M1pYSmpZWE5sZlM1MGRIVjdkR1Y0ZEMxMGNtRnVjMlp2Y20wNmRYQndaWEpqWVhObGZTNTBkRzU3ZEdWNGRDMTBjbUZ1YzJadmNtMDZibTl1WlgwdVppMDJMQzVtTFdobFlXUnNhVzVsZTJadmJuUXRjMmw2WlRvMmNtVnRmUzVtTFRVc0xtWXRjM1ZpYUdWaFpHeHBibVY3Wm05dWRDMXphWHBsT2pWeVpXMTlMbVl4ZTJadmJuUXRjMmw2WlRvemNtVnRmUzVtTW50bWIyNTBMWE5wZW1VNk1pNHlOWEpsYlgwdVpqTjdabTl1ZEMxemFYcGxPakV1TlhKbGJYMHVaalI3Wm05dWRDMXphWHBsT2pFdU1qVnlaVzE5TG1ZMWUyWnZiblF0YzJsNlpUb3hjbVZ0ZlM1bU5udG1iMjUwTFhOcGVtVTZMamczTlhKbGJYMHVaamQ3Wm05dWRDMXphWHBsT2k0M05YSmxiWDB1YldWaGMzVnlaWHR0WVhndGQybGtkR2c2TXpCbGJYMHViV1ZoYzNWeVpTMTNhV1JsZTIxaGVDMTNhV1IwYURvek5HVnRmUzV0WldGemRYSmxMVzVoY25KdmQzdHRZWGd0ZDJsa2RHZzZNakJsYlgwdWFXNWtaVzUwZTNSbGVIUXRhVzVrWlc1ME9qRmxiVHR0WVhKbmFXNHRkRzl3T2pBN2JXRnlaMmx1TFdKdmRIUnZiVG93ZlM1emJXRnNiQzFqWVhCemUyWnZiblF0ZG1GeWFXRnVkRHB6YldGc2JDMWpZWEJ6ZlM1MGNuVnVZMkYwWlh0M2FHbDBaUzF6Y0dGalpUcHViM2R5WVhBN2IzWmxjbVpzYjNjNmFHbGtaR1Z1TzNSbGVIUXRiM1psY21ac2IzYzZaV3hzYVhCemFYTjlMbTkyWlhKbWJHOTNMV052Ym5SaGFXNWxjbnR2ZG1WeVpteHZkeTE1T25OamNtOXNiSDB1WTJWdWRHVnllMjFoY21kcGJpMXNaV1owT21GMWRHOTlMbU5sYm5SbGNpd3ViWEl0WVhWMGIzdHRZWEpuYVc0dGNtbG5hSFE2WVhWMGIzMHViV3d0WVhWMGIzdHRZWEpuYVc0dGJHVm1kRHBoZFhSdmZTNWpiR2x3ZTNCdmMybDBhVzl1T21acGVHVmtJV2x0Y0c5eWRHRnVkRHRmY0c5emFYUnBiMjQ2WVdKemIyeDFkR1VoYVcxd2IzSjBZVzUwTzJOc2FYQTZjbVZqZENneGNIZ2dNWEI0SURGd2VDQXhjSGdwTzJOc2FYQTZjbVZqZENneGNIZ3NNWEI0TERGd2VDd3hjSGdwZlM1M2N5MXViM0p0WVd4N2QyaHBkR1V0YzNCaFkyVTZibTl5YldGc2ZTNXViM2R5WVhCN2QyaHBkR1V0YzNCaFkyVTZibTkzY21Gd2ZTNXdjbVY3ZDJocGRHVXRjM0JoWTJVNmNISmxmUzUyTFdKaGMyVjdkbVZ5ZEdsallXd3RZV3hwWjI0NlltRnpaV3hwYm1WOUxuWXRiV2xrZTNabGNuUnBZMkZzTFdGc2FXZHVPbTFwWkdSc1pYMHVkaTEwYjNCN2RtVnlkR2xqWVd3dFlXeHBaMjQ2ZEc5d2ZTNTJMV0owYlh0MlpYSjBhV05oYkMxaGJHbG5ianBpYjNSMGIyMTlMbVJwYlh0dmNHRmphWFI1T2pGOUxtUnBiU3d1WkdsdE9tWnZZM1Z6TEM1a2FXMDZhRzkyWlhKN2RISmhibk5wZEdsdmJqcHZjR0ZqYVhSNUlDNHhOWE1nWldGelpTMXBibjB1WkdsdE9tWnZZM1Z6TEM1a2FXMDZhRzkyWlhKN2IzQmhZMmwwZVRvdU5YMHVaR2x0T21GamRHbDJaWHR2Y0dGamFYUjVPaTQ0TzNSeVlXNXphWFJwYjI0NmIzQmhZMmwwZVNBdU1UVnpJR1ZoYzJVdGIzVjBmUzVuYkc5M0xDNW5iRzkzT21adlkzVnpMQzVuYkc5M09taHZkbVZ5ZTNSeVlXNXphWFJwYjI0NmIzQmhZMmwwZVNBdU1UVnpJR1ZoYzJVdGFXNTlMbWRzYjNjNlptOWpkWE1zTG1kc2IzYzZhRzkyWlhKN2IzQmhZMmwwZVRveGZTNW9hV1JsTFdOb2FXeGtJQzVqYUdsc1pIdHZjR0ZqYVhSNU9qQTdkSEpoYm5OcGRHbHZianB2Y0dGamFYUjVJQzR4TlhNZ1pXRnpaUzFwYm4wdWFHbGtaUzFqYUdsc1pEcGhZM1JwZG1VZ0xtTm9hV3hrTEM1b2FXUmxMV05vYVd4a09tWnZZM1Z6SUM1amFHbHNaQ3d1YUdsa1pTMWphR2xzWkRwb2IzWmxjaUF1WTJocGJHUjdiM0JoWTJsMGVUb3hPM1J5WVc1emFYUnBiMjQ2YjNCaFkybDBlU0F1TVRWeklHVmhjMlV0YVc1OUxuVnVaR1Z5YkdsdVpTMW9iM1psY2pwbWIyTjFjeXd1ZFc1a1pYSnNhVzVsTFdodmRtVnlPbWh2ZG1WeWUzUmxlSFF0WkdWamIzSmhkR2x2YmpwMWJtUmxjbXhwYm1WOUxtZHliM2Q3TFcxdmVpMXZjM2d0Wm05dWRDMXpiVzl2ZEdocGJtYzZaM0poZVhOallXeGxPeTEzWldKcmFYUXRZbUZqYTJaaFkyVXRkbWx6YVdKcGJHbDBlVHBvYVdSa1pXNDdZbUZqYTJaaFkyVXRkbWx6YVdKcGJHbDBlVHBvYVdSa1pXNDdMWGRsWW10cGRDMTBjbUZ1YzJadmNtMDZkSEpoYm5Oc1lYUmxXaWd3S1R0MGNtRnVjMlp2Y20wNmRISmhibk5zWVhSbFdpZ3dLVHQwY21GdWMybDBhVzl1T2kxM1pXSnJhWFF0ZEhKaGJuTm1iM0p0SUM0eU5YTWdaV0Z6WlMxdmRYUTdkSEpoYm5OcGRHbHZianAwY21GdWMyWnZjbTBnTGpJMWN5QmxZWE5sTFc5MWREdDBjbUZ1YzJsMGFXOXVPblJ5WVc1elptOXliU0F1TWpWeklHVmhjMlV0YjNWMExDMTNaV0pyYVhRdGRISmhibk5tYjNKdElDNHlOWE1nWldGelpTMXZkWFI5TG1keWIzYzZabTlqZFhNc0xtZHliM2M2YUc5MlpYSjdMWGRsWW10cGRDMTBjbUZ1YzJadmNtMDZjMk5oYkdVb01TNHdOU2s3ZEhKaGJuTm1iM0p0T25OallXeGxLREV1TURVcGZTNW5jbTkzT21GamRHbDJaWHN0ZDJWaWEybDBMWFJ5WVc1elptOXliVHB6WTJGc1pTZ3VPU2s3ZEhKaGJuTm1iM0p0T25OallXeGxLQzQ1S1gwdVozSnZkeTFzWVhKblpYc3RiVzk2TFc5emVDMW1iMjUwTFhOdGIyOTBhR2x1WnpwbmNtRjVjMk5oYkdVN0xYZGxZbXRwZEMxaVlXTnJabUZqWlMxMmFYTnBZbWxzYVhSNU9taHBaR1JsYmp0aVlXTnJabUZqWlMxMmFYTnBZbWxzYVhSNU9taHBaR1JsYmpzdGQyVmlhMmwwTFhSeVlXNXpabTl5YlRwMGNtRnVjMnhoZEdWYUtEQXBPM1J5WVc1elptOXliVHAwY21GdWMyeGhkR1ZhS0RBcE8zUnlZVzV6YVhScGIyNDZMWGRsWW10cGRDMTBjbUZ1YzJadmNtMGdMakkxY3lCbFlYTmxMV2x1TFc5MWREdDBjbUZ1YzJsMGFXOXVPblJ5WVc1elptOXliU0F1TWpWeklHVmhjMlV0YVc0dGIzVjBPM1J5WVc1emFYUnBiMjQ2ZEhKaGJuTm1iM0p0SUM0eU5YTWdaV0Z6WlMxcGJpMXZkWFFzTFhkbFltdHBkQzEwY21GdWMyWnZjbTBnTGpJMWN5QmxZWE5sTFdsdUxXOTFkSDB1WjNKdmR5MXNZWEpuWlRwbWIyTjFjeXd1WjNKdmR5MXNZWEpuWlRwb2IzWmxjbnN0ZDJWaWEybDBMWFJ5WVc1elptOXliVHB6WTJGc1pTZ3hMaklwTzNSeVlXNXpabTl5YlRwelkyRnNaU2d4TGpJcGZTNW5jbTkzTFd4aGNtZGxPbUZqZEdsMlpYc3RkMlZpYTJsMExYUnlZVzV6Wm05eWJUcHpZMkZzWlNndU9UVXBPM1J5WVc1elptOXliVHB6WTJGc1pTZ3VPVFVwZlM1d2IybHVkR1Z5T21odmRtVnlMQzV6YUdGa2IzY3RhRzkyWlhKN1kzVnljMjl5T25CdmFXNTBaWEo5TG5Ob1lXUnZkeTFvYjNabGNudHdiM05wZEdsdmJqcHlaV3hoZEdsMlpUdDBjbUZ1YzJsMGFXOXVPbUZzYkNBdU5YTWdZM1ZpYVdNdFltVjZhV1Z5S0M0eE5qVXNMamcwTEM0ME5Dd3hLWDB1YzJoaFpHOTNMV2h2ZG1WeU9tRm1kR1Z5ZTJOdmJuUmxiblE2WENKY0lqdGliM2d0YzJoaFpHOTNPakFnTUNBeE5uQjRJREp3ZUNCeVoySmhLREFzTUN3d0xDNHlLVHRpYjNKa1pYSXRjbUZrYVhWek9tbHVhR1Z5YVhRN2IzQmhZMmwwZVRvd08zQnZjMmwwYVc5dU9tRmljMjlzZFhSbE8zUnZjRG93TzJ4bFpuUTZNRHQzYVdSMGFEb3hNREFsTzJobGFXZG9kRG94TURBbE8zb3RhVzVrWlhnNkxURTdkSEpoYm5OcGRHbHZianB2Y0dGamFYUjVJQzQxY3lCamRXSnBZeTFpWlhwcFpYSW9MakUyTlN3dU9EUXNMalEwTERFcGZTNXphR0ZrYjNjdGFHOTJaWEk2Wm05amRYTTZZV1owWlhJc0xuTm9ZV1J2ZHkxb2IzWmxjanBvYjNabGNqcGhablJsY250dmNHRmphWFI1T2pGOUxtSm5MV0Z1YVcxaGRHVXNMbUpuTFdGdWFXMWhkR1U2Wm05amRYTXNMbUpuTFdGdWFXMWhkR1U2YUc5MlpYSjdkSEpoYm5OcGRHbHZianBpWVdOclozSnZkVzVrTFdOdmJHOXlJQzR4TlhNZ1pXRnpaUzFwYmkxdmRYUjlMbm90TUh0NkxXbHVaR1Y0T2pCOUxub3RNWHQ2TFdsdVpHVjRPakY5TG5vdE1udDZMV2x1WkdWNE9qSjlMbm90TTN0NkxXbHVaR1Y0T2pOOUxub3ROSHQ2TFdsdVpHVjRPalI5TG5vdE5YdDZMV2x1WkdWNE9qVjlMbm90T1RrNWUzb3RhVzVrWlhnNk9UazVmUzU2TFRrNU9UbDdlaTFwYm1SbGVEbzVPVGs1ZlM1NkxXMWhlSHQ2TFdsdVpHVjRPakl4TkRjME9ETTJORGQ5TG5vdGFXNW9aWEpwZEh0NkxXbHVaR1Y0T21sdWFHVnlhWFI5TG5vdGFXNXBkR2xoYkh0NkxXbHVaR1Y0T21GMWRHOTlMbm90ZFc1elpYUjdlaTFwYm1SbGVEcDFibk5sZEgwdWJtVnpkR1ZrTFdOdmNIa3RiR2x1WlMxb1pXbG5hSFFnYjJ3c0xtNWxjM1JsWkMxamIzQjVMV3hwYm1VdGFHVnBaMmgwSUhBc0xtNWxjM1JsWkMxamIzQjVMV3hwYm1VdGFHVnBaMmgwSUhWc2UyeHBibVV0YUdWcFoyaDBPakV1TlgwdWJtVnpkR1ZrTFdobFlXUnNhVzVsTFd4cGJtVXRhR1ZwWjJoMElHZ3hMQzV1WlhOMFpXUXRhR1ZoWkd4cGJtVXRiR2x1WlMxb1pXbG5hSFFnYURJc0xtNWxjM1JsWkMxb1pXRmtiR2x1WlMxc2FXNWxMV2hsYVdkb2RDQm9NeXd1Ym1WemRHVmtMV2hsWVdSc2FXNWxMV3hwYm1VdGFHVnBaMmgwSUdnMExDNXVaWE4wWldRdGFHVmhaR3hwYm1VdGJHbHVaUzFvWldsbmFIUWdhRFVzTG01bGMzUmxaQzFvWldGa2JHbHVaUzFzYVc1bExXaGxhV2RvZENCb05udHNhVzVsTFdobGFXZG9kRG94TGpJMWZTNXVaWE4wWldRdGJHbHpkQzF5WlhObGRDQnZiQ3d1Ym1WemRHVmtMV3hwYzNRdGNtVnpaWFFnZFd4N2NHRmtaR2x1Wnkxc1pXWjBPakE3YldGeVoybHVMV3hsWm5RNk1EdHNhWE4wTFhOMGVXeGxMWFI1Y0dVNmJtOXVaWDB1Ym1WemRHVmtMV052Y0hrdGFXNWtaVzUwSUhBcmNIdDBaWGgwTFdsdVpHVnVkRG94WlcwN2JXRnlaMmx1TFhSdmNEb3dPMjFoY21kcGJpMWliM1IwYjIwNk1IMHVibVZ6ZEdWa0xXTnZjSGt0YzJWd1lYSmhkRzl5SUhBcmNIdHRZWEpuYVc0dGRHOXdPakV1TldWdGZTNXVaWE4wWldRdGFXMW5JR2x0WjN0M2FXUjBhRG94TURBbE8yMWhlQzEzYVdSMGFEb3hNREFsTzJScGMzQnNZWGs2WW14dlkydDlMbTVsYzNSbFpDMXNhVzVyY3lCaGUyTnZiRzl5T2lNek5UZGxaR1E3ZEhKaGJuTnBkR2x2YmpwamIyeHZjaUF1TVRWeklHVmhjMlV0YVc1OUxtNWxjM1JsWkMxc2FXNXJjeUJoT21adlkzVnpMQzV1WlhOMFpXUXRiR2x1YTNNZ1lUcG9iM1psY250amIyeHZjam9qT1RaalkyWm1PM1J5WVc1emFYUnBiMjQ2WTI5c2IzSWdMakUxY3lCbFlYTmxMV2x1ZlM1a1pXSjFaeUFxZTI5MWRHeHBibVU2TVhCNElITnZiR2xrSUdkdmJHUjlMbVJsWW5WbkxYZG9hWFJsSUNwN2IzVjBiR2x1WlRveGNIZ2djMjlzYVdRZ0kyWm1abjB1WkdWaWRXY3RZbXhoWTJzZ0tudHZkWFJzYVc1bE9qRndlQ0J6YjJ4cFpDQWpNREF3ZlM1a1pXSjFaeTFuY21sa2UySmhZMnRuY205MWJtUTZkSEpoYm5Od1lYSmxiblFnZFhKc0tHUmhkR0U2YVcxaFoyVXZjRzVuTzJKaGMyVTJOQ3hwVmtKUFVuY3dTMGRuYjBGQlFVRk9VMVZvUlZWblFVRkJRV2RCUVVGQlNVTkJXVUZCUVVSRlJEYzJURUZCUVVGR1JXeEZVVlpTTkVGWFVFRkRPVGN2T1hnd1pVTnpRVVZRWjNkQlZreHphR1J3UlU1SmVHTkJRVUZCUVZOVlZrOVNTelZEV1VsSlBTa2djbVZ3WldGMElEQWdNSDB1WkdWaWRXY3RaM0pwWkMweE5udGlZV05yWjNKdmRXNWtPblJ5WVc1emNHRnlaVzUwSUhWeWJDaGtZWFJoT21sdFlXZGxMM0J1Wnp0aVlYTmxOalFzYVZaQ1QxSjNNRXRIWjI5QlFVRkJUbE5WYUVWVlowRkJRVUpCUVVGQlFWRkRRVmxCUVVGQlpqZ3ZPV2hCUVVGQlRXdHNSVkZXVWpSQlYwOW5RMHg2TDJJd1pYQkJZVFpWUjNWQ1QzRlJTRTlSU0V4VlowWkZSRzVCWW1OQ1dqUlZSM2RFVDJ0cFEyNXJTV2hrWjA1blRuaEJXVUZwV1d4RUt6aHpSWFZ2T0VGQlFVRkJVMVZXVDFKTE5VTlpTVWs5S1NCeVpYQmxZWFFnTUNBd2ZTNWtaV0oxWnkxbmNtbGtMVGd0YzI5c2FXUjdZbUZqYTJkeWIzVnVaRG9qWm1abUlIVnliQ2hrWVhSaE9tbHRZV2RsTDJkcFpqdGlZWE5sTmpRc1VqQnNSMDlFWkdoRFFVRkpRVkJGUVVGQlJIY3ZkMFI0THk4dkx5OTNRVUZCUTNkQlFVRkJRVU5CUVVsQlFVRkRSRnBSZG1kaFpXSXZiSGhpUVVsTFFUaDVNRUZQZHowOUtTQnlaWEJsWVhRZ01DQXdmUzVrWldKMVp5MW5jbWxrTFRFMkxYTnZiR2xrZTJKaFkydG5jbTkxYm1RNkkyWm1aaUIxY213b1pHRjBZVHBwYldGblpTOW5hV1k3WW1GelpUWTBMRkl3YkVkUFJHUm9SVUZCVVVGUVJVRkJRVVIzTDNkRWVDOTRXSGt2THk4dkwzbDNRVUZCUVVGRlFVRlJRVUZCUTBsYWVWQkxZMnRaUkZGR2MySTJXbkZFT0RWcVdqSXJRbXQzYVZKR1MyVm9hSEZSUTFGblJFaGpaM2RGUWxGQk55a2djbVZ3WldGMElEQWdNSDFBYldWa2FXRWdjMk55WldWdUlHRnVaQ0FvYldsdUxYZHBaSFJvT2pNd1pXMHBleTVoYzNCbFkzUXRjbUYwYVc4dGJuTjdhR1ZwWjJoME9qQTdjRzl6YVhScGIyNDZjbVZzWVhScGRtVjlMbUZ6Y0dWamRDMXlZWFJwYnkwdE1UWjRPUzF1YzN0d1lXUmthVzVuTFdKdmRIUnZiVG8xTmk0eU5TVjlMbUZ6Y0dWamRDMXlZWFJwYnkwdE9YZ3hOaTF1YzN0d1lXUmthVzVuTFdKdmRIUnZiVG94TnpjdU56Y2xmUzVoYzNCbFkzUXRjbUYwYVc4dExUUjRNeTF1YzN0d1lXUmthVzVuTFdKdmRIUnZiVG8zTlNWOUxtRnpjR1ZqZEMxeVlYUnBieTB0TTNnMExXNXplM0JoWkdScGJtY3RZbTkwZEc5dE9qRXpNeTR6TXlWOUxtRnpjR1ZqZEMxeVlYUnBieTB0Tm5nMExXNXplM0JoWkdScGJtY3RZbTkwZEc5dE9qWTJMallsZlM1aGMzQmxZM1F0Y21GMGFXOHRMVFI0TmkxdWMzdHdZV1JrYVc1bkxXSnZkSFJ2YlRveE5UQWxmUzVoYzNCbFkzUXRjbUYwYVc4dExUaDROUzF1YzN0d1lXUmthVzVuTFdKdmRIUnZiVG8yTWk0MUpYMHVZWE53WldOMExYSmhkR2x2TFMwMWVEZ3Ribk43Y0dGa1pHbHVaeTFpYjNSMGIyMDZNVFl3SlgwdVlYTndaV04wTFhKaGRHbHZMUzAzZURVdGJuTjdjR0ZrWkdsdVp5MWliM1IwYjIwNk56RXVORElsZlM1aGMzQmxZM1F0Y21GMGFXOHRMVFY0TnkxdWMzdHdZV1JrYVc1bkxXSnZkSFJ2YlRveE5EQWxmUzVoYzNCbFkzUXRjbUYwYVc4dExURjRNUzF1YzN0d1lXUmthVzVuTFdKdmRIUnZiVG94TURBbGZTNWhjM0JsWTNRdGNtRjBhVzh0TFc5aWFtVmpkQzF1YzN0d2IzTnBkR2x2YmpwaFluTnZiSFYwWlR0MGIzQTZNRHR5YVdkb2REb3dPMkp2ZEhSdmJUb3dPMnhsWm5RNk1EdDNhV1IwYURveE1EQWxPMmhsYVdkb2REb3hNREFsTzNvdGFXNWtaWGc2TVRBd2ZTNWpiM1psY2kxdWMzdGlZV05yWjNKdmRXNWtMWE5wZW1VNlkyOTJaWEloYVcxd2IzSjBZVzUwZlM1amIyNTBZV2x1TFc1emUySmhZMnRuY205MWJtUXRjMmw2WlRwamIyNTBZV2x1SVdsdGNHOXlkR0Z1ZEgwdVltY3RZMlZ1ZEdWeUxXNXplMkpoWTJ0bmNtOTFibVF0Y0c5emFYUnBiMjQ2TlRBbGZTNWlaeTFqWlc1MFpYSXRibk1zTG1KbkxYUnZjQzF1YzN0aVlXTnJaM0p2ZFc1a0xYSmxjR1ZoZERwdWJ5MXlaWEJsWVhSOUxtSm5MWFJ2Y0MxdWMzdGlZV05yWjNKdmRXNWtMWEJ2YzJsMGFXOXVPblJ2Y0gwdVltY3RjbWxuYUhRdGJuTjdZbUZqYTJkeWIzVnVaQzF3YjNOcGRHbHZiam94TURBbGZTNWlaeTFpYjNSMGIyMHRibk1zTG1KbkxYSnBaMmgwTFc1emUySmhZMnRuY205MWJtUXRjbVZ3WldGME9tNXZMWEpsY0dWaGRIMHVZbWN0WW05MGRHOXRMVzV6ZTJKaFkydG5jbTkxYm1RdGNHOXphWFJwYjI0NlltOTBkRzl0ZlM1aVp5MXNaV1owTFc1emUySmhZMnRuY205MWJtUXRjbVZ3WldGME9tNXZMWEpsY0dWaGREdGlZV05yWjNKdmRXNWtMWEJ2YzJsMGFXOXVPakI5TG05MWRHeHBibVV0Ym5ON2IzVjBiR2x1WlRveGNIZ2djMjlzYVdSOUxtOTFkR3hwYm1VdGRISmhibk53WVhKbGJuUXRibk43YjNWMGJHbHVaVG94Y0hnZ2MyOXNhV1FnZEhKaGJuTndZWEpsYm5SOUxtOTFkR3hwYm1VdE1DMXVjM3R2ZFhSc2FXNWxPakI5TG1KaExXNXplMkp2Y21SbGNpMXpkSGxzWlRwemIyeHBaRHRpYjNKa1pYSXRkMmxrZEdnNk1YQjRmUzVpZEMxdWMzdGliM0prWlhJdGRHOXdMWE4wZVd4bE9uTnZiR2xrTzJKdmNtUmxjaTEwYjNBdGQybGtkR2c2TVhCNGZTNWljaTF1YzN0aWIzSmtaWEl0Y21sbmFIUXRjM1I1YkdVNmMyOXNhV1E3WW05eVpHVnlMWEpwWjJoMExYZHBaSFJvT2pGd2VIMHVZbUl0Ym5ON1ltOXlaR1Z5TFdKdmRIUnZiUzF6ZEhsc1pUcHpiMnhwWkR0aWIzSmtaWEl0WW05MGRHOXRMWGRwWkhSb09qRndlSDB1WW13dGJuTjdZbTl5WkdWeUxXeGxablF0YzNSNWJHVTZjMjlzYVdRN1ltOXlaR1Z5TFd4bFpuUXRkMmxrZEdnNk1YQjRmUzVpYmkxdWMzdGliM0prWlhJdGMzUjViR1U2Ym05dVpUdGliM0prWlhJdGQybGtkR2c2TUgwdVluSXdMVzV6ZTJKdmNtUmxjaTF5WVdScGRYTTZNSDB1WW5JeExXNXplMkp2Y21SbGNpMXlZV1JwZFhNNkxqRXlOWEpsYlgwdVluSXlMVzV6ZTJKdmNtUmxjaTF5WVdScGRYTTZMakkxY21WdGZTNWljak10Ym5ON1ltOXlaR1Z5TFhKaFpHbDFjem91TlhKbGJYMHVZbkkwTFc1emUySnZjbVJsY2kxeVlXUnBkWE02TVhKbGJYMHVZbkl0TVRBd0xXNXplMkp2Y21SbGNpMXlZV1JwZFhNNk1UQXdKWDB1WW5JdGNHbHNiQzF1YzN0aWIzSmtaWEl0Y21Ga2FYVnpPams1T1Rsd2VIMHVZbkl0TFdKdmRIUnZiUzF1YzN0aWIzSmtaWEl0ZEc5d0xXeGxablF0Y21Ga2FYVnpPakE3WW05eVpHVnlMWFJ2Y0MxeWFXZG9kQzF5WVdScGRYTTZNSDB1WW5JdExYUnZjQzF1YzN0aWIzSmtaWEl0WW05MGRHOXRMWEpwWjJoMExYSmhaR2wxY3pvd2ZTNWljaTB0Y21sbmFIUXRibk1zTG1KeUxTMTBiM0F0Ym5ON1ltOXlaR1Z5TFdKdmRIUnZiUzFzWldaMExYSmhaR2wxY3pvd2ZTNWljaTB0Y21sbmFIUXRibk43WW05eVpHVnlMWFJ2Y0Mxc1pXWjBMWEpoWkdsMWN6b3dmUzVpY2kwdGJHVm1kQzF1YzN0aWIzSmtaWEl0ZEc5d0xYSnBaMmgwTFhKaFpHbDFjem93TzJKdmNtUmxjaTFpYjNSMGIyMHRjbWxuYUhRdGNtRmthWFZ6T2pCOUxtSXRMV1J2ZEhSbFpDMXVjM3RpYjNKa1pYSXRjM1I1YkdVNlpHOTBkR1ZrZlM1aUxTMWtZWE5vWldRdGJuTjdZbTl5WkdWeUxYTjBlV3hsT21SaGMyaGxaSDB1WWkwdGMyOXNhV1F0Ym5ON1ltOXlaR1Z5TFhOMGVXeGxPbk52Ykdsa2ZTNWlMUzF1YjI1bExXNXplMkp2Y21SbGNpMXpkSGxzWlRwdWIyNWxmUzVpZHpBdGJuTjdZbTl5WkdWeUxYZHBaSFJvT2pCOUxtSjNNUzF1YzN0aWIzSmtaWEl0ZDJsa2RHZzZMakV5TlhKbGJYMHVZbmN5TFc1emUySnZjbVJsY2kxM2FXUjBhRG91TWpWeVpXMTlMbUozTXkxdWMzdGliM0prWlhJdGQybGtkR2c2TGpWeVpXMTlMbUozTkMxdWMzdGliM0prWlhJdGQybGtkR2c2TVhKbGJYMHVZbmMxTFc1emUySnZjbVJsY2kxM2FXUjBhRG95Y21WdGZTNWlkQzB3TFc1emUySnZjbVJsY2kxMGIzQXRkMmxrZEdnNk1IMHVZbkl0TUMxdWMzdGliM0prWlhJdGNtbG5hSFF0ZDJsa2RHZzZNSDB1WW1JdE1DMXVjM3RpYjNKa1pYSXRZbTkwZEc5dExYZHBaSFJvT2pCOUxtSnNMVEF0Ym5ON1ltOXlaR1Z5TFd4bFpuUXRkMmxrZEdnNk1IMHVjMmhoWkc5M0xURXRibk43WW05NExYTm9ZV1J2ZHpvd0lEQWdOSEI0SURKd2VDQnlaMkpoS0RBc01Dd3dMQzR5S1gwdWMyaGhaRzkzTFRJdGJuTjdZbTk0TFhOb1lXUnZkem93SURBZ09IQjRJREp3ZUNCeVoySmhLREFzTUN3d0xDNHlLWDB1YzJoaFpHOTNMVE10Ym5ON1ltOTRMWE5vWVdSdmR6b3ljSGdnTW5CNElEUndlQ0F5Y0hnZ2NtZGlZU2d3TERBc01Dd3VNaWw5TG5Ob1lXUnZkeTAwTFc1emUySnZlQzF6YUdGa2IzYzZNbkI0SURKd2VDQTRjSGdnTUNCeVoySmhLREFzTUN3d0xDNHlLWDB1YzJoaFpHOTNMVFV0Ym5ON1ltOTRMWE5vWVdSdmR6bzBjSGdnTkhCNElEaHdlQ0F3SUhKblltRW9NQ3d3TERBc0xqSXBmUzUwYjNBdE1DMXVjM3QwYjNBNk1IMHViR1ZtZEMwd0xXNXplMnhsWm5RNk1IMHVjbWxuYUhRdE1DMXVjM3R5YVdkb2REb3dmUzVpYjNSMGIyMHRNQzF1YzN0aWIzUjBiMjA2TUgwdWRHOXdMVEV0Ym5ON2RHOXdPakZ5WlcxOUxteGxablF0TVMxdWMzdHNaV1owT2pGeVpXMTlMbkpwWjJoMExURXRibk43Y21sbmFIUTZNWEpsYlgwdVltOTBkRzl0TFRFdGJuTjdZbTkwZEc5dE9qRnlaVzE5TG5SdmNDMHlMVzV6ZTNSdmNEb3ljbVZ0ZlM1c1pXWjBMVEl0Ym5ON2JHVm1kRG95Y21WdGZTNXlhV2RvZEMweUxXNXplM0pwWjJoME9qSnlaVzE5TG1KdmRIUnZiUzB5TFc1emUySnZkSFJ2YlRveWNtVnRmUzUwYjNBdExURXRibk43ZEc5d09pMHhjbVZ0ZlM1eWFXZG9kQzB0TVMxdWMzdHlhV2RvZERvdE1YSmxiWDB1WW05MGRHOXRMUzB4TFc1emUySnZkSFJ2YlRvdE1YSmxiWDB1YkdWbWRDMHRNUzF1YzN0c1pXWjBPaTB4Y21WdGZTNTBiM0F0TFRJdGJuTjdkRzl3T2kweWNtVnRmUzV5YVdkb2RDMHRNaTF1YzN0eWFXZG9kRG90TW5KbGJYMHVZbTkwZEc5dExTMHlMVzV6ZTJKdmRIUnZiVG90TW5KbGJYMHViR1ZtZEMwdE1pMXVjM3RzWldaME9pMHljbVZ0ZlM1aFluTnZiSFYwWlMwdFptbHNiQzF1YzN0MGIzQTZNRHR5YVdkb2REb3dPMkp2ZEhSdmJUb3dPMnhsWm5RNk1IMHVZMnd0Ym5ON1kyeGxZWEk2YkdWbWRIMHVZM0l0Ym5ON1kyeGxZWEk2Y21sbmFIUjlMbU5pTFc1emUyTnNaV0Z5T21KdmRHaDlMbU51TFc1emUyTnNaV0Z5T201dmJtVjlMbVJ1TFc1emUyUnBjM0JzWVhrNmJtOXVaWDB1WkdrdGJuTjdaR2x6Y0d4aGVUcHBibXhwYm1WOUxtUmlMVzV6ZTJScGMzQnNZWGs2WW14dlkydDlMbVJwWWkxdWMzdGthWE53YkdGNU9tbHViR2x1WlMxaWJHOWphMzB1WkdsMExXNXplMlJwYzNCc1lYazZhVzVzYVc1bExYUmhZbXhsZlM1a2RDMXVjM3RrYVhOd2JHRjVPblJoWW14bGZTNWtkR010Ym5ON1pHbHpjR3hoZVRwMFlXSnNaUzFqWld4c2ZTNWtkQzF5YjNjdGJuTjdaR2x6Y0d4aGVUcDBZV0pzWlMxeWIzZDlMbVIwTFhKdmR5MW5jbTkxY0MxdWMzdGthWE53YkdGNU9uUmhZbXhsTFhKdmR5MW5jbTkxY0gwdVpIUXRZMjlzZFcxdUxXNXplMlJwYzNCc1lYazZkR0ZpYkdVdFkyOXNkVzF1ZlM1a2RDMWpiMngxYlc0dFozSnZkWEF0Ym5ON1pHbHpjR3hoZVRwMFlXSnNaUzFqYjJ4MWJXNHRaM0p2ZFhCOUxtUjBMUzFtYVhobFpDMXVjM3QwWVdKc1pTMXNZWGx2ZFhRNlptbDRaV1E3ZDJsa2RHZzZNVEF3SlgwdVpteGxlQzF1YzN0a2FYTndiR0Y1T21ac1pYaDlMbWx1YkdsdVpTMW1iR1Y0TFc1emUyUnBjM0JzWVhrNmFXNXNhVzVsTFdac1pYaDlMbVpzWlhndFlYVjBieTF1YzN0bWJHVjRPakVnTVNCaGRYUnZPMjFwYmkxM2FXUjBhRG93TzIxcGJpMW9aV2xuYUhRNk1IMHVabXhsZUMxdWIyNWxMVzV6ZTJac1pYZzZibTl1WlgwdVpteGxlQzFqYjJ4MWJXNHRibk43Wm14bGVDMWthWEpsWTNScGIyNDZZMjlzZFcxdWZTNW1iR1Y0TFhKdmR5MXVjM3RtYkdWNExXUnBjbVZqZEdsdmJqcHliM2Q5TG1ac1pYZ3RkM0poY0MxdWMzdG1iR1Y0TFhkeVlYQTZkM0poY0gwdVpteGxlQzF1YjNkeVlYQXRibk43Wm14bGVDMTNjbUZ3T201dmQzSmhjSDB1Wm14bGVDMTNjbUZ3TFhKbGRtVnljMlV0Ym5ON1pteGxlQzEzY21Gd09uZHlZWEF0Y21WMlpYSnpaWDB1Wm14bGVDMWpiMngxYlc0dGNtVjJaWEp6WlMxdWMzdG1iR1Y0TFdScGNtVmpkR2x2YmpwamIyeDFiVzR0Y21WMlpYSnpaWDB1Wm14bGVDMXliM2N0Y21WMlpYSnpaUzF1YzN0bWJHVjRMV1JwY21WamRHbHZianB5YjNjdGNtVjJaWEp6WlgwdWFYUmxiWE10YzNSaGNuUXRibk43WVd4cFoyNHRhWFJsYlhNNlpteGxlQzF6ZEdGeWRIMHVhWFJsYlhNdFpXNWtMVzV6ZTJGc2FXZHVMV2wwWlcxek9tWnNaWGd0Wlc1a2ZTNXBkR1Z0Y3kxalpXNTBaWEl0Ym5ON1lXeHBaMjR0YVhSbGJYTTZZMlZ1ZEdWeWZTNXBkR1Z0Y3kxaVlYTmxiR2x1WlMxdWMzdGhiR2xuYmkxcGRHVnRjenBpWVhObGJHbHVaWDB1YVhSbGJYTXRjM1J5WlhSamFDMXVjM3RoYkdsbmJpMXBkR1Z0Y3pwemRISmxkR05vZlM1elpXeG1MWE4wWVhKMExXNXplMkZzYVdkdUxYTmxiR1k2Wm14bGVDMXpkR0Z5ZEgwdWMyVnNaaTFsYm1RdGJuTjdZV3hwWjI0dGMyVnNaanBtYkdWNExXVnVaSDB1YzJWc1ppMWpaVzUwWlhJdGJuTjdZV3hwWjI0dGMyVnNaanBqWlc1MFpYSjlMbk5sYkdZdFltRnpaV3hwYm1VdGJuTjdZV3hwWjI0dGMyVnNaanBpWVhObGJHbHVaWDB1YzJWc1ppMXpkSEpsZEdOb0xXNXplMkZzYVdkdUxYTmxiR1k2YzNSeVpYUmphSDB1YW5WemRHbG1lUzF6ZEdGeWRDMXVjM3RxZFhOMGFXWjVMV052Ym5SbGJuUTZabXhsZUMxemRHRnlkSDB1YW5WemRHbG1lUzFsYm1RdGJuTjdhblZ6ZEdsbWVTMWpiMjUwWlc1ME9tWnNaWGd0Wlc1a2ZTNXFkWE4wYVdaNUxXTmxiblJsY2kxdWMzdHFkWE4wYVdaNUxXTnZiblJsYm5RNlkyVnVkR1Z5ZlM1cWRYTjBhV1o1TFdKbGRIZGxaVzR0Ym5ON2FuVnpkR2xtZVMxamIyNTBaVzUwT25Od1lXTmxMV0psZEhkbFpXNTlMbXAxYzNScFpua3RZWEp2ZFc1a0xXNXplMnAxYzNScFpua3RZMjl1ZEdWdWREcHpjR0ZqWlMxaGNtOTFibVI5TG1OdmJuUmxiblF0YzNSaGNuUXRibk43WVd4cFoyNHRZMjl1ZEdWdWREcG1iR1Y0TFhOMFlYSjBmUzVqYjI1MFpXNTBMV1Z1WkMxdWMzdGhiR2xuYmkxamIyNTBaVzUwT21ac1pYZ3RaVzVrZlM1amIyNTBaVzUwTFdObGJuUmxjaTF1YzN0aGJHbG5iaTFqYjI1MFpXNTBPbU5sYm5SbGNuMHVZMjl1ZEdWdWRDMWlaWFIzWldWdUxXNXplMkZzYVdkdUxXTnZiblJsYm5RNmMzQmhZMlV0WW1WMGQyVmxibjB1WTI5dWRHVnVkQzFoY205MWJtUXRibk43WVd4cFoyNHRZMjl1ZEdWdWREcHpjR0ZqWlMxaGNtOTFibVI5TG1OdmJuUmxiblF0YzNSeVpYUmphQzF1YzN0aGJHbG5iaTFqYjI1MFpXNTBPbk4wY21WMFkyaDlMbTl5WkdWeUxUQXRibk43YjNKa1pYSTZNSDB1YjNKa1pYSXRNUzF1YzN0dmNtUmxjam94ZlM1dmNtUmxjaTB5TFc1emUyOXlaR1Z5T2pKOUxtOXlaR1Z5TFRNdGJuTjdiM0prWlhJNk0zMHViM0prWlhJdE5DMXVjM3R2Y21SbGNqbzBmUzV2Y21SbGNpMDFMVzV6ZTI5eVpHVnlPalY5TG05eVpHVnlMVFl0Ym5ON2IzSmtaWEk2Tm4wdWIzSmtaWEl0TnkxdWMzdHZjbVJsY2pvM2ZTNXZjbVJsY2kwNExXNXplMjl5WkdWeU9qaDlMbTl5WkdWeUxXeGhjM1F0Ym5ON2IzSmtaWEk2T1RrNU9UbDlMbVpzWlhndFozSnZkeTB3TFc1emUyWnNaWGd0WjNKdmR6b3dmUzVtYkdWNExXZHliM2N0TVMxdWMzdG1iR1Y0TFdkeWIzYzZNWDB1Wm14bGVDMXphSEpwYm1zdE1DMXVjM3RtYkdWNExYTm9jbWx1YXpvd2ZTNW1iR1Y0TFhOb2NtbHVheTB4TFc1emUyWnNaWGd0YzJoeWFXNXJPakY5TG1ac0xXNXplMlpzYjJGME9teGxablI5TG1ac0xXNXpMQzVtY2kxdWMzdGZaR2x6Y0d4aGVUcHBibXhwYm1WOUxtWnlMVzV6ZTJac2IyRjBPbkpwWjJoMGZTNW1iaTF1YzN0bWJHOWhkRHB1YjI1bGZTNXBMVzV6ZTJadmJuUXRjM1I1YkdVNmFYUmhiR2xqZlM1bWN5MXViM0p0WVd3dGJuTjdabTl1ZEMxemRIbHNaVHB1YjNKdFlXeDlMbTV2Y20xaGJDMXVjM3RtYjI1MExYZGxhV2RvZERvME1EQjlMbUl0Ym5ON1ptOXVkQzEzWldsbmFIUTZOekF3ZlM1bWR6RXRibk43Wm05dWRDMTNaV2xuYUhRNk1UQXdmUzVtZHpJdGJuTjdabTl1ZEMxM1pXbG5hSFE2TWpBd2ZTNW1kek10Ym5ON1ptOXVkQzEzWldsbmFIUTZNekF3ZlM1bWR6UXRibk43Wm05dWRDMTNaV2xuYUhRNk5EQXdmUzVtZHpVdGJuTjdabTl1ZEMxM1pXbG5hSFE2TlRBd2ZTNW1kell0Ym5ON1ptOXVkQzEzWldsbmFIUTZOakF3ZlM1bWR6Y3Ribk43Wm05dWRDMTNaV2xuYUhRNk56QXdmUzVtZHpndGJuTjdabTl1ZEMxM1pXbG5hSFE2T0RBd2ZTNW1kemt0Ym5ON1ptOXVkQzEzWldsbmFIUTZPVEF3ZlM1b01TMXVjM3RvWldsbmFIUTZNWEpsYlgwdWFESXRibk43YUdWcFoyaDBPakp5WlcxOUxtZ3pMVzV6ZTJobGFXZG9kRG8wY21WdGZTNW9OQzF1YzN0b1pXbG5hSFE2T0hKbGJYMHVhRFV0Ym5ON2FHVnBaMmgwT2pFMmNtVnRmUzVvTFRJMUxXNXplMmhsYVdkb2REb3lOU1Y5TG1ndE5UQXRibk43YUdWcFoyaDBPalV3SlgwdWFDMDNOUzF1YzN0b1pXbG5hSFE2TnpVbGZTNW9MVEV3TUMxdWMzdG9aV2xuYUhRNk1UQXdKWDB1YldsdUxXZ3RNVEF3TFc1emUyMXBiaTFvWldsbmFIUTZNVEF3SlgwdWRtZ3RNalV0Ym5ON2FHVnBaMmgwT2pJMWRtaDlMblpvTFRVd0xXNXplMmhsYVdkb2REbzFNSFpvZlM1MmFDMDNOUzF1YzN0b1pXbG5hSFE2TnpWMmFIMHVkbWd0TVRBd0xXNXplMmhsYVdkb2REb3hNREIyYUgwdWJXbHVMWFpvTFRFd01DMXVjM3R0YVc0dGFHVnBaMmgwT2pFd01IWm9mUzVvTFdGMWRHOHRibk43YUdWcFoyaDBPbUYxZEc5OUxtZ3RhVzVvWlhKcGRDMXVjM3RvWldsbmFIUTZhVzVvWlhKcGRIMHVkSEpoWTJ0bFpDMXVjM3RzWlhSMFpYSXRjM0JoWTJsdVp6b3VNV1Z0ZlM1MGNtRmphMlZrTFhScFoyaDBMVzV6ZTJ4bGRIUmxjaTF6Y0dGamFXNW5PaTB1TURWbGJYMHVkSEpoWTJ0bFpDMXRaV2RoTFc1emUyeGxkSFJsY2kxemNHRmphVzVuT2k0eU5XVnRmUzVzYUMxemIyeHBaQzF1YzN0c2FXNWxMV2hsYVdkb2REb3hmUzVzYUMxMGFYUnNaUzF1YzN0c2FXNWxMV2hsYVdkb2REb3hMakkxZlM1c2FDMWpiM0I1TFc1emUyeHBibVV0YUdWcFoyaDBPakV1TlgwdWJYY3RNVEF3TFc1emUyMWhlQzEzYVdSMGFEb3hNREFsZlM1dGR6RXRibk43YldGNExYZHBaSFJvT2pGeVpXMTlMbTEzTWkxdWMzdHRZWGd0ZDJsa2RHZzZNbkpsYlgwdWJYY3pMVzV6ZTIxaGVDMTNhV1IwYURvMGNtVnRmUzV0ZHpRdGJuTjdiV0Y0TFhkcFpIUm9Pamh5WlcxOUxtMTNOUzF1YzN0dFlYZ3RkMmxrZEdnNk1UWnlaVzE5TG0xM05pMXVjM3R0WVhndGQybGtkR2c2TXpKeVpXMTlMbTEzTnkxdWMzdHRZWGd0ZDJsa2RHZzZORGh5WlcxOUxtMTNPQzF1YzN0dFlYZ3RkMmxrZEdnNk5qUnlaVzE5TG0xM09TMXVjM3R0WVhndGQybGtkR2c2T1RaeVpXMTlMbTEzTFc1dmJtVXRibk43YldGNExYZHBaSFJvT201dmJtVjlMbmN4TFc1emUzZHBaSFJvT2pGeVpXMTlMbmN5TFc1emUzZHBaSFJvT2pKeVpXMTlMbmN6TFc1emUzZHBaSFJvT2pSeVpXMTlMbmMwTFc1emUzZHBaSFJvT2poeVpXMTlMbmMxTFc1emUzZHBaSFJvT2pFMmNtVnRmUzUzTFRFd0xXNXplM2RwWkhSb09qRXdKWDB1ZHkweU1DMXVjM3QzYVdSMGFEb3lNQ1Y5TG5jdE1qVXRibk43ZDJsa2RHZzZNalVsZlM1M0xUTXdMVzV6ZTNkcFpIUm9Pak13SlgwdWR5MHpNeTF1YzN0M2FXUjBhRG96TXlWOUxuY3RNelF0Ym5ON2QybGtkR2c2TXpRbGZTNTNMVFF3TFc1emUzZHBaSFJvT2pRd0pYMHVkeTAxTUMxdWMzdDNhV1IwYURvMU1DVjlMbmN0TmpBdGJuTjdkMmxrZEdnNk5qQWxmUzUzTFRjd0xXNXplM2RwWkhSb09qY3dKWDB1ZHkwM05TMXVjM3QzYVdSMGFEbzNOU1Y5TG5jdE9EQXRibk43ZDJsa2RHZzZPREFsZlM1M0xUa3dMVzV6ZTNkcFpIUm9Pamt3SlgwdWR5MHhNREF0Ym5ON2QybGtkR2c2TVRBd0pYMHVkeTEwYUdseVpDMXVjM3QzYVdSMGFEb3pNeTR6TXpNek15VjlMbmN0ZEhkdkxYUm9hWEprY3kxdWMzdDNhV1IwYURvMk5pNDJOalkyTnlWOUxuY3RZWFYwYnkxdWMzdDNhV1IwYURwaGRYUnZmUzV2ZG1WeVpteHZkeTEyYVhOcFlteGxMVzV6ZTI5MlpYSm1iRzkzT25acGMybGliR1Y5TG05MlpYSm1iRzkzTFdocFpHUmxiaTF1YzN0dmRtVnlabXh2ZHpwb2FXUmtaVzU5TG05MlpYSm1iRzkzTFhOamNtOXNiQzF1YzN0dmRtVnlabXh2ZHpwelkzSnZiR3g5TG05MlpYSm1iRzkzTFdGMWRHOHRibk43YjNabGNtWnNiM2M2WVhWMGIzMHViM1psY21ac2IzY3RlQzEyYVhOcFlteGxMVzV6ZTI5MlpYSm1iRzkzTFhnNmRtbHphV0pzWlgwdWIzWmxjbVpzYjNjdGVDMW9hV1JrWlc0dGJuTjdiM1psY21ac2IzY3RlRHBvYVdSa1pXNTlMbTkyWlhKbWJHOTNMWGd0YzJOeWIyeHNMVzV6ZTI5MlpYSm1iRzkzTFhnNmMyTnliMnhzZlM1dmRtVnlabXh2ZHkxNExXRjFkRzh0Ym5ON2IzWmxjbVpzYjNjdGVEcGhkWFJ2ZlM1dmRtVnlabXh2ZHkxNUxYWnBjMmxpYkdVdGJuTjdiM1psY21ac2IzY3RlVHAyYVhOcFlteGxmUzV2ZG1WeVpteHZkeTE1TFdocFpHUmxiaTF1YzN0dmRtVnlabXh2ZHkxNU9taHBaR1JsYm4wdWIzWmxjbVpzYjNjdGVTMXpZM0p2Ykd3dGJuTjdiM1psY21ac2IzY3RlVHB6WTNKdmJHeDlMbTkyWlhKbWJHOTNMWGt0WVhWMGJ5MXVjM3R2ZG1WeVpteHZkeTE1T21GMWRHOTlMbk4wWVhScFl5MXVjM3R3YjNOcGRHbHZianB6ZEdGMGFXTjlMbkpsYkdGMGFYWmxMVzV6ZTNCdmMybDBhVzl1T25KbGJHRjBhWFpsZlM1aFluTnZiSFYwWlMxdWMzdHdiM05wZEdsdmJqcGhZbk52YkhWMFpYMHVabWw0WldRdGJuTjdjRzl6YVhScGIyNDZabWw0WldSOUxuSnZkR0YwWlMwME5TMXVjM3N0ZDJWaWEybDBMWFJ5WVc1elptOXliVHB5YjNSaGRHVW9ORFZrWldjcE8zUnlZVzV6Wm05eWJUcHliM1JoZEdVb05EVmtaV2NwZlM1eWIzUmhkR1V0T1RBdGJuTjdMWGRsWW10cGRDMTBjbUZ1YzJadmNtMDZjbTkwWVhSbEtEa3daR1ZuS1R0MGNtRnVjMlp2Y20wNmNtOTBZWFJsS0Rrd1pHVm5LWDB1Y205MFlYUmxMVEV6TlMxdWMzc3RkMlZpYTJsMExYUnlZVzV6Wm05eWJUcHliM1JoZEdVb01UTTFaR1ZuS1R0MGNtRnVjMlp2Y20wNmNtOTBZWFJsS0RFek5XUmxaeWw5TG5KdmRHRjBaUzB4T0RBdGJuTjdMWGRsWW10cGRDMTBjbUZ1YzJadmNtMDZjbTkwWVhSbEtERTRNR1JsWnlrN2RISmhibk5tYjNKdE9uSnZkR0YwWlNneE9EQmtaV2NwZlM1eWIzUmhkR1V0TWpJMUxXNXpleTEzWldKcmFYUXRkSEpoYm5ObWIzSnRPbkp2ZEdGMFpTZ3lNalZrWldjcE8zUnlZVzV6Wm05eWJUcHliM1JoZEdVb01qSTFaR1ZuS1gwdWNtOTBZWFJsTFRJM01DMXVjM3N0ZDJWaWEybDBMWFJ5WVc1elptOXliVHB5YjNSaGRHVW9NamN3WkdWbktUdDBjbUZ1YzJadmNtMDZjbTkwWVhSbEtESTNNR1JsWnlsOUxuSnZkR0YwWlMwek1UVXRibk43TFhkbFltdHBkQzEwY21GdWMyWnZjbTA2Y205MFlYUmxLRE14TldSbFp5azdkSEpoYm5ObWIzSnRPbkp2ZEdGMFpTZ3pNVFZrWldjcGZTNXdZVEF0Ym5ON2NHRmtaR2x1Wnpvd2ZTNXdZVEV0Ym5ON2NHRmtaR2x1WnpvdU1qVnlaVzE5TG5CaE1pMXVjM3R3WVdSa2FXNW5PaTQxY21WdGZTNXdZVE10Ym5ON2NHRmtaR2x1WnpveGNtVnRmUzV3WVRRdGJuTjdjR0ZrWkdsdVp6b3ljbVZ0ZlM1d1lUVXRibk43Y0dGa1pHbHVaem8wY21WdGZTNXdZVFl0Ym5ON2NHRmtaR2x1WnpvNGNtVnRmUzV3WVRjdGJuTjdjR0ZrWkdsdVp6b3hObkpsYlgwdWNHd3dMVzV6ZTNCaFpHUnBibWN0YkdWbWREb3dmUzV3YkRFdGJuTjdjR0ZrWkdsdVp5MXNaV1owT2k0eU5YSmxiWDB1Y0d3eUxXNXplM0JoWkdScGJtY3RiR1ZtZERvdU5YSmxiWDB1Y0d3ekxXNXplM0JoWkdScGJtY3RiR1ZtZERveGNtVnRmUzV3YkRRdGJuTjdjR0ZrWkdsdVp5MXNaV1owT2pKeVpXMTlMbkJzTlMxdWMzdHdZV1JrYVc1bkxXeGxablE2TkhKbGJYMHVjR3cyTFc1emUzQmhaR1JwYm1jdGJHVm1kRG80Y21WdGZTNXdiRGN0Ym5ON2NHRmtaR2x1Wnkxc1pXWjBPakUyY21WdGZTNXdjakF0Ym5ON2NHRmtaR2x1WnkxeWFXZG9kRG93ZlM1d2NqRXRibk43Y0dGa1pHbHVaeTF5YVdkb2REb3VNalZ5WlcxOUxuQnlNaTF1YzN0d1lXUmthVzVuTFhKcFoyaDBPaTQxY21WdGZTNXdjak10Ym5ON2NHRmtaR2x1WnkxeWFXZG9kRG94Y21WdGZTNXdjalF0Ym5ON2NHRmtaR2x1WnkxeWFXZG9kRG95Y21WdGZTNXdjalV0Ym5ON2NHRmtaR2x1WnkxeWFXZG9kRG8wY21WdGZTNXdjall0Ym5ON2NHRmtaR2x1WnkxeWFXZG9kRG80Y21WdGZTNXdjamN0Ym5ON2NHRmtaR2x1WnkxeWFXZG9kRG94Tm5KbGJYMHVjR0l3TFc1emUzQmhaR1JwYm1jdFltOTBkRzl0T2pCOUxuQmlNUzF1YzN0d1lXUmthVzVuTFdKdmRIUnZiVG91TWpWeVpXMTlMbkJpTWkxdWMzdHdZV1JrYVc1bkxXSnZkSFJ2YlRvdU5YSmxiWDB1Y0dJekxXNXplM0JoWkdScGJtY3RZbTkwZEc5dE9qRnlaVzE5TG5CaU5DMXVjM3R3WVdSa2FXNW5MV0p2ZEhSdmJUb3ljbVZ0ZlM1d1lqVXRibk43Y0dGa1pHbHVaeTFpYjNSMGIyMDZOSEpsYlgwdWNHSTJMVzV6ZTNCaFpHUnBibWN0WW05MGRHOXRPamh5WlcxOUxuQmlOeTF1YzN0d1lXUmthVzVuTFdKdmRIUnZiVG94Tm5KbGJYMHVjSFF3TFc1emUzQmhaR1JwYm1jdGRHOXdPakI5TG5CME1TMXVjM3R3WVdSa2FXNW5MWFJ2Y0RvdU1qVnlaVzE5TG5CME1pMXVjM3R3WVdSa2FXNW5MWFJ2Y0RvdU5YSmxiWDB1Y0hRekxXNXplM0JoWkdScGJtY3RkRzl3T2pGeVpXMTlMbkIwTkMxdWMzdHdZV1JrYVc1bkxYUnZjRG95Y21WdGZTNXdkRFV0Ym5ON2NHRmtaR2x1WnkxMGIzQTZOSEpsYlgwdWNIUTJMVzV6ZTNCaFpHUnBibWN0ZEc5d09qaHlaVzE5TG5CME55MXVjM3R3WVdSa2FXNW5MWFJ2Y0RveE5uSmxiWDB1Y0hZd0xXNXplM0JoWkdScGJtY3RkRzl3T2pBN2NHRmtaR2x1WnkxaWIzUjBiMjA2TUgwdWNIWXhMVzV6ZTNCaFpHUnBibWN0ZEc5d09pNHlOWEpsYlR0d1lXUmthVzVuTFdKdmRIUnZiVG91TWpWeVpXMTlMbkIyTWkxdWMzdHdZV1JrYVc1bkxYUnZjRG91TlhKbGJUdHdZV1JrYVc1bkxXSnZkSFJ2YlRvdU5YSmxiWDB1Y0hZekxXNXplM0JoWkdScGJtY3RkRzl3T2pGeVpXMDdjR0ZrWkdsdVp5MWliM1IwYjIwNk1YSmxiWDB1Y0hZMExXNXplM0JoWkdScGJtY3RkRzl3T2pKeVpXMDdjR0ZrWkdsdVp5MWliM1IwYjIwNk1uSmxiWDB1Y0hZMUxXNXplM0JoWkdScGJtY3RkRzl3T2pSeVpXMDdjR0ZrWkdsdVp5MWliM1IwYjIwNk5ISmxiWDB1Y0hZMkxXNXplM0JoWkdScGJtY3RkRzl3T2poeVpXMDdjR0ZrWkdsdVp5MWliM1IwYjIwNk9ISmxiWDB1Y0hZM0xXNXplM0JoWkdScGJtY3RkRzl3T2pFMmNtVnRPM0JoWkdScGJtY3RZbTkwZEc5dE9qRTJjbVZ0ZlM1d2FEQXRibk43Y0dGa1pHbHVaeTFzWldaME9qQTdjR0ZrWkdsdVp5MXlhV2RvZERvd2ZTNXdhREV0Ym5ON2NHRmtaR2x1Wnkxc1pXWjBPaTR5TlhKbGJUdHdZV1JrYVc1bkxYSnBaMmgwT2k0eU5YSmxiWDB1Y0dneUxXNXplM0JoWkdScGJtY3RiR1ZtZERvdU5YSmxiVHR3WVdSa2FXNW5MWEpwWjJoME9pNDFjbVZ0ZlM1d2FETXRibk43Y0dGa1pHbHVaeTFzWldaME9qRnlaVzA3Y0dGa1pHbHVaeTF5YVdkb2REb3hjbVZ0ZlM1d2FEUXRibk43Y0dGa1pHbHVaeTFzWldaME9qSnlaVzA3Y0dGa1pHbHVaeTF5YVdkb2REb3ljbVZ0ZlM1d2FEVXRibk43Y0dGa1pHbHVaeTFzWldaME9qUnlaVzA3Y0dGa1pHbHVaeTF5YVdkb2REbzBjbVZ0ZlM1d2FEWXRibk43Y0dGa1pHbHVaeTFzWldaME9qaHlaVzA3Y0dGa1pHbHVaeTF5YVdkb2REbzRjbVZ0ZlM1d2FEY3Ribk43Y0dGa1pHbHVaeTFzWldaME9qRTJjbVZ0TzNCaFpHUnBibWN0Y21sbmFIUTZNVFp5WlcxOUxtMWhNQzF1YzN0dFlYSm5hVzQ2TUgwdWJXRXhMVzV6ZTIxaGNtZHBiam91TWpWeVpXMTlMbTFoTWkxdWMzdHRZWEpuYVc0NkxqVnlaVzE5TG0xaE15MXVjM3R0WVhKbmFXNDZNWEpsYlgwdWJXRTBMVzV6ZTIxaGNtZHBiam95Y21WdGZTNXRZVFV0Ym5ON2JXRnlaMmx1T2pSeVpXMTlMbTFoTmkxdWMzdHRZWEpuYVc0Nk9ISmxiWDB1YldFM0xXNXplMjFoY21kcGJqb3hObkpsYlgwdWJXd3dMVzV6ZTIxaGNtZHBiaTFzWldaME9qQjlMbTFzTVMxdWMzdHRZWEpuYVc0dGJHVm1kRG91TWpWeVpXMTlMbTFzTWkxdWMzdHRZWEpuYVc0dGJHVm1kRG91TlhKbGJYMHViV3d6TFc1emUyMWhjbWRwYmkxc1pXWjBPakZ5WlcxOUxtMXNOQzF1YzN0dFlYSm5hVzR0YkdWbWREb3ljbVZ0ZlM1dGJEVXRibk43YldGeVoybHVMV3hsWm5RNk5ISmxiWDB1Yld3MkxXNXplMjFoY21kcGJpMXNaV1owT2poeVpXMTlMbTFzTnkxdWMzdHRZWEpuYVc0dGJHVm1kRG94Tm5KbGJYMHViWEl3TFc1emUyMWhjbWRwYmkxeWFXZG9kRG93ZlM1dGNqRXRibk43YldGeVoybHVMWEpwWjJoME9pNHlOWEpsYlgwdWJYSXlMVzV6ZTIxaGNtZHBiaTF5YVdkb2REb3VOWEpsYlgwdWJYSXpMVzV6ZTIxaGNtZHBiaTF5YVdkb2REb3hjbVZ0ZlM1dGNqUXRibk43YldGeVoybHVMWEpwWjJoME9qSnlaVzE5TG0xeU5TMXVjM3R0WVhKbmFXNHRjbWxuYUhRNk5ISmxiWDB1YlhJMkxXNXplMjFoY21kcGJpMXlhV2RvZERvNGNtVnRmUzV0Y2pjdGJuTjdiV0Z5WjJsdUxYSnBaMmgwT2pFMmNtVnRmUzV0WWpBdGJuTjdiV0Z5WjJsdUxXSnZkSFJ2YlRvd2ZTNXRZakV0Ym5ON2JXRnlaMmx1TFdKdmRIUnZiVG91TWpWeVpXMTlMbTFpTWkxdWMzdHRZWEpuYVc0dFltOTBkRzl0T2k0MWNtVnRmUzV0WWpNdGJuTjdiV0Z5WjJsdUxXSnZkSFJ2YlRveGNtVnRmUzV0WWpRdGJuTjdiV0Z5WjJsdUxXSnZkSFJ2YlRveWNtVnRmUzV0WWpVdGJuTjdiV0Z5WjJsdUxXSnZkSFJ2YlRvMGNtVnRmUzV0WWpZdGJuTjdiV0Z5WjJsdUxXSnZkSFJ2YlRvNGNtVnRmUzV0WWpjdGJuTjdiV0Z5WjJsdUxXSnZkSFJ2YlRveE5uSmxiWDB1YlhRd0xXNXplMjFoY21kcGJpMTBiM0E2TUgwdWJYUXhMVzV6ZTIxaGNtZHBiaTEwYjNBNkxqSTFjbVZ0ZlM1dGRESXRibk43YldGeVoybHVMWFJ2Y0RvdU5YSmxiWDB1YlhRekxXNXplMjFoY21kcGJpMTBiM0E2TVhKbGJYMHViWFEwTFc1emUyMWhjbWRwYmkxMGIzQTZNbkpsYlgwdWJYUTFMVzV6ZTIxaGNtZHBiaTEwYjNBNk5ISmxiWDB1YlhRMkxXNXplMjFoY21kcGJpMTBiM0E2T0hKbGJYMHViWFEzTFc1emUyMWhjbWRwYmkxMGIzQTZNVFp5WlcxOUxtMTJNQzF1YzN0dFlYSm5hVzR0ZEc5d09qQTdiV0Z5WjJsdUxXSnZkSFJ2YlRvd2ZTNXRkakV0Ym5ON2JXRnlaMmx1TFhSdmNEb3VNalZ5WlcwN2JXRnlaMmx1TFdKdmRIUnZiVG91TWpWeVpXMTlMbTEyTWkxdWMzdHRZWEpuYVc0dGRHOXdPaTQxY21WdE8yMWhjbWRwYmkxaWIzUjBiMjA2TGpWeVpXMTlMbTEyTXkxdWMzdHRZWEpuYVc0dGRHOXdPakZ5WlcwN2JXRnlaMmx1TFdKdmRIUnZiVG94Y21WdGZTNXRkalF0Ym5ON2JXRnlaMmx1TFhSdmNEb3ljbVZ0TzIxaGNtZHBiaTFpYjNSMGIyMDZNbkpsYlgwdWJYWTFMVzV6ZTIxaGNtZHBiaTEwYjNBNk5ISmxiVHR0WVhKbmFXNHRZbTkwZEc5dE9qUnlaVzE5TG0xMk5pMXVjM3R0WVhKbmFXNHRkRzl3T2poeVpXMDdiV0Z5WjJsdUxXSnZkSFJ2YlRvNGNtVnRmUzV0ZGpjdGJuTjdiV0Z5WjJsdUxYUnZjRG94Tm5KbGJUdHRZWEpuYVc0dFltOTBkRzl0T2pFMmNtVnRmUzV0YURBdGJuTjdiV0Z5WjJsdUxXeGxablE2TUR0dFlYSm5hVzR0Y21sbmFIUTZNSDB1YldneExXNXplMjFoY21kcGJpMXNaV1owT2k0eU5YSmxiVHR0WVhKbmFXNHRjbWxuYUhRNkxqSTFjbVZ0ZlM1dGFESXRibk43YldGeVoybHVMV3hsWm5RNkxqVnlaVzA3YldGeVoybHVMWEpwWjJoME9pNDFjbVZ0ZlM1dGFETXRibk43YldGeVoybHVMV3hsWm5RNk1YSmxiVHR0WVhKbmFXNHRjbWxuYUhRNk1YSmxiWDB1YldnMExXNXplMjFoY21kcGJpMXNaV1owT2pKeVpXMDdiV0Z5WjJsdUxYSnBaMmgwT2pKeVpXMTlMbTFvTlMxdWMzdHRZWEpuYVc0dGJHVm1kRG8wY21WdE8yMWhjbWRwYmkxeWFXZG9kRG8wY21WdGZTNXRhRFl0Ym5ON2JXRnlaMmx1TFd4bFpuUTZPSEpsYlR0dFlYSm5hVzR0Y21sbmFIUTZPSEpsYlgwdWJXZzNMVzV6ZTIxaGNtZHBiaTFzWldaME9qRTJjbVZ0TzIxaGNtZHBiaTF5YVdkb2REb3hObkpsYlgwdWJtRXhMVzV6ZTIxaGNtZHBiam90TGpJMWNtVnRmUzV1WVRJdGJuTjdiV0Z5WjJsdU9pMHVOWEpsYlgwdWJtRXpMVzV6ZTIxaGNtZHBiam90TVhKbGJYMHVibUUwTFc1emUyMWhjbWRwYmpvdE1uSmxiWDB1Ym1FMUxXNXplMjFoY21kcGJqb3ROSEpsYlgwdWJtRTJMVzV6ZTIxaGNtZHBiam90T0hKbGJYMHVibUUzTFc1emUyMWhjbWRwYmpvdE1UWnlaVzE5TG01c01TMXVjM3R0WVhKbmFXNHRiR1ZtZERvdExqSTFjbVZ0ZlM1dWJESXRibk43YldGeVoybHVMV3hsWm5RNkxTNDFjbVZ0ZlM1dWJETXRibk43YldGeVoybHVMV3hsWm5RNkxURnlaVzE5TG01c05DMXVjM3R0WVhKbmFXNHRiR1ZtZERvdE1uSmxiWDB1Ym13MUxXNXplMjFoY21kcGJpMXNaV1owT2kwMGNtVnRmUzV1YkRZdGJuTjdiV0Z5WjJsdUxXeGxablE2TFRoeVpXMTlMbTVzTnkxdWMzdHRZWEpuYVc0dGJHVm1kRG90TVRaeVpXMTlMbTV5TVMxdWMzdHRZWEpuYVc0dGNtbG5hSFE2TFM0eU5YSmxiWDB1Ym5JeUxXNXplMjFoY21kcGJpMXlhV2RvZERvdExqVnlaVzE5TG01eU15MXVjM3R0WVhKbmFXNHRjbWxuYUhRNkxURnlaVzE5TG01eU5DMXVjM3R0WVhKbmFXNHRjbWxuYUhRNkxUSnlaVzE5TG01eU5TMXVjM3R0WVhKbmFXNHRjbWxuYUhRNkxUUnlaVzE5TG01eU5pMXVjM3R0WVhKbmFXNHRjbWxuYUhRNkxUaHlaVzE5TG01eU55MXVjM3R0WVhKbmFXNHRjbWxuYUhRNkxURTJjbVZ0ZlM1dVlqRXRibk43YldGeVoybHVMV0p2ZEhSdmJUb3RMakkxY21WdGZTNXVZakl0Ym5ON2JXRnlaMmx1TFdKdmRIUnZiVG90TGpWeVpXMTlMbTVpTXkxdWMzdHRZWEpuYVc0dFltOTBkRzl0T2kweGNtVnRmUzV1WWpRdGJuTjdiV0Z5WjJsdUxXSnZkSFJ2YlRvdE1uSmxiWDB1Ym1JMUxXNXplMjFoY21kcGJpMWliM1IwYjIwNkxUUnlaVzE5TG01aU5pMXVjM3R0WVhKbmFXNHRZbTkwZEc5dE9pMDRjbVZ0ZlM1dVlqY3Ribk43YldGeVoybHVMV0p2ZEhSdmJUb3RNVFp5WlcxOUxtNTBNUzF1YzN0dFlYSm5hVzR0ZEc5d09pMHVNalZ5WlcxOUxtNTBNaTF1YzN0dFlYSm5hVzR0ZEc5d09pMHVOWEpsYlgwdWJuUXpMVzV6ZTIxaGNtZHBiaTEwYjNBNkxURnlaVzE5TG01ME5DMXVjM3R0WVhKbmFXNHRkRzl3T2kweWNtVnRmUzV1ZERVdGJuTjdiV0Z5WjJsdUxYUnZjRG90TkhKbGJYMHViblEyTFc1emUyMWhjbWRwYmkxMGIzQTZMVGh5WlcxOUxtNTBOeTF1YzN0dFlYSm5hVzR0ZEc5d09pMHhObkpsYlgwdWMzUnlhV3RsTFc1emUzUmxlSFF0WkdWamIzSmhkR2x2Ympwc2FXNWxMWFJvY205MVoyaDlMblZ1WkdWeWJHbHVaUzF1YzN0MFpYaDBMV1JsWTI5eVlYUnBiMjQ2ZFc1a1pYSnNhVzVsZlM1dWJ5MTFibVJsY214cGJtVXRibk43ZEdWNGRDMWtaV052Y21GMGFXOXVPbTV2Ym1WOUxuUnNMVzV6ZTNSbGVIUXRZV3hwWjI0NmJHVm1kSDB1ZEhJdGJuTjdkR1Y0ZEMxaGJHbG5ianB5YVdkb2RIMHVkR010Ym5ON2RHVjRkQzFoYkdsbmJqcGpaVzUwWlhKOUxuUnFMVzV6ZTNSbGVIUXRZV3hwWjI0NmFuVnpkR2xtZVgwdWRIUmpMVzV6ZTNSbGVIUXRkSEpoYm5ObWIzSnRPbU5oY0dsMFlXeHBlbVY5TG5SMGJDMXVjM3QwWlhoMExYUnlZVzV6Wm05eWJUcHNiM2RsY21OaGMyVjlMblIwZFMxdWMzdDBaWGgwTFhSeVlXNXpabTl5YlRwMWNIQmxjbU5oYzJWOUxuUjBiaTF1YzN0MFpYaDBMWFJ5WVc1elptOXliVHB1YjI1bGZTNW1MVFl0Ym5Nc0xtWXRhR1ZoWkd4cGJtVXRibk43Wm05dWRDMXphWHBsT2paeVpXMTlMbVl0TlMxdWN5d3VaaTF6ZFdKb1pXRmtiR2x1WlMxdWMzdG1iMjUwTFhOcGVtVTZOWEpsYlgwdVpqRXRibk43Wm05dWRDMXphWHBsT2pOeVpXMTlMbVl5TFc1emUyWnZiblF0YzJsNlpUb3lMakkxY21WdGZTNW1NeTF1YzN0bWIyNTBMWE5wZW1VNk1TNDFjbVZ0ZlM1bU5DMXVjM3RtYjI1MExYTnBlbVU2TVM0eU5YSmxiWDB1WmpVdGJuTjdabTl1ZEMxemFYcGxPakZ5WlcxOUxtWTJMVzV6ZTJadmJuUXRjMmw2WlRvdU9EYzFjbVZ0ZlM1bU55MXVjM3RtYjI1MExYTnBlbVU2TGpjMWNtVnRmUzV0WldGemRYSmxMVzV6ZTIxaGVDMTNhV1IwYURvek1HVnRmUzV0WldGemRYSmxMWGRwWkdVdGJuTjdiV0Y0TFhkcFpIUm9Pak0wWlcxOUxtMWxZWE4xY21VdGJtRnljbTkzTFc1emUyMWhlQzEzYVdSMGFEb3lNR1Z0ZlM1cGJtUmxiblF0Ym5ON2RHVjRkQzFwYm1SbGJuUTZNV1Z0TzIxaGNtZHBiaTEwYjNBNk1EdHRZWEpuYVc0dFltOTBkRzl0T2pCOUxuTnRZV3hzTFdOaGNITXRibk43Wm05dWRDMTJZWEpwWVc1ME9uTnRZV3hzTFdOaGNITjlMblJ5ZFc1allYUmxMVzV6ZTNkb2FYUmxMWE53WVdObE9tNXZkM0poY0R0dmRtVnlabXh2ZHpwb2FXUmtaVzQ3ZEdWNGRDMXZkbVZ5Wm14dmR6cGxiR3hwY0hOcGMzMHVZMlZ1ZEdWeUxXNXplMjFoY21kcGJpMXNaV1owT21GMWRHOTlMbU5sYm5SbGNpMXVjeXd1YlhJdFlYVjBieTF1YzN0dFlYSm5hVzR0Y21sbmFIUTZZWFYwYjMwdWJXd3RZWFYwYnkxdWMzdHRZWEpuYVc0dGJHVm1kRHBoZFhSdmZTNWpiR2x3TFc1emUzQnZjMmwwYVc5dU9tWnBlR1ZrSVdsdGNHOXlkR0Z1ZER0ZmNHOXphWFJwYjI0NllXSnpiMngxZEdVaGFXMXdiM0owWVc1ME8yTnNhWEE2Y21WamRDZ3hjSGdnTVhCNElERndlQ0F4Y0hncE8yTnNhWEE2Y21WamRDZ3hjSGdzTVhCNExERndlQ3d4Y0hncGZTNTNjeTF1YjNKdFlXd3Ribk43ZDJocGRHVXRjM0JoWTJVNmJtOXliV0ZzZlM1dWIzZHlZWEF0Ym5ON2QyaHBkR1V0YzNCaFkyVTZibTkzY21Gd2ZTNXdjbVV0Ym5ON2QyaHBkR1V0YzNCaFkyVTZjSEpsZlM1MkxXSmhjMlV0Ym5ON2RtVnlkR2xqWVd3dFlXeHBaMjQ2WW1GelpXeHBibVY5TG5ZdGJXbGtMVzV6ZTNabGNuUnBZMkZzTFdGc2FXZHVPbTFwWkdSc1pYMHVkaTEwYjNBdGJuTjdkbVZ5ZEdsallXd3RZV3hwWjI0NmRHOXdmUzUyTFdKMGJTMXVjM3QyWlhKMGFXTmhiQzFoYkdsbmJqcGliM1IwYjIxOWZVQnRaV1JwWVNCelkzSmxaVzRnWVc1a0lDaHRhVzR0ZDJsa2RHZzZNekJsYlNrZ1lXNWtJQ2h0WVhndGQybGtkR2c2TmpCbGJTbDdMbUZ6Y0dWamRDMXlZWFJwYnkxdGUyaGxhV2RvZERvd08zQnZjMmwwYVc5dU9uSmxiR0YwYVhabGZTNWhjM0JsWTNRdGNtRjBhVzh0TFRFMmVEa3RiWHR3WVdSa2FXNW5MV0p2ZEhSdmJUbzFOaTR5TlNWOUxtRnpjR1ZqZEMxeVlYUnBieTB0T1hneE5pMXRlM0JoWkdScGJtY3RZbTkwZEc5dE9qRTNOeTQzTnlWOUxtRnpjR1ZqZEMxeVlYUnBieTB0TkhnekxXMTdjR0ZrWkdsdVp5MWliM1IwYjIwNk56VWxmUzVoYzNCbFkzUXRjbUYwYVc4dExUTjROQzF0ZTNCaFpHUnBibWN0WW05MGRHOXRPakV6TXk0ek15VjlMbUZ6Y0dWamRDMXlZWFJwYnkwdE5uZzBMVzE3Y0dGa1pHbHVaeTFpYjNSMGIyMDZOall1TmlWOUxtRnpjR1ZqZEMxeVlYUnBieTB0TkhnMkxXMTdjR0ZrWkdsdVp5MWliM1IwYjIwNk1UVXdKWDB1WVhOd1pXTjBMWEpoZEdsdkxTMDRlRFV0Ylh0d1lXUmthVzVuTFdKdmRIUnZiVG8yTWk0MUpYMHVZWE53WldOMExYSmhkR2x2TFMwMWVEZ3RiWHR3WVdSa2FXNW5MV0p2ZEhSdmJUb3hOakFsZlM1aGMzQmxZM1F0Y21GMGFXOHRMVGQ0TlMxdGUzQmhaR1JwYm1jdFltOTBkRzl0T2pjeExqUXlKWDB1WVhOd1pXTjBMWEpoZEdsdkxTMDFlRGN0Ylh0d1lXUmthVzVuTFdKdmRIUnZiVG94TkRBbGZTNWhjM0JsWTNRdGNtRjBhVzh0TFRGNE1TMXRlM0JoWkdScGJtY3RZbTkwZEc5dE9qRXdNQ1Y5TG1GemNHVmpkQzF5WVhScGJ5MHRiMkpxWldOMExXMTdjRzl6YVhScGIyNDZZV0p6YjJ4MWRHVTdkRzl3T2pBN2NtbG5hSFE2TUR0aWIzUjBiMjA2TUR0c1pXWjBPakE3ZDJsa2RHZzZNVEF3SlR0b1pXbG5hSFE2TVRBd0pUdDZMV2x1WkdWNE9qRXdNSDB1WTI5MlpYSXRiWHRpWVdOclozSnZkVzVrTFhOcGVtVTZZMjkyWlhJaGFXMXdiM0owWVc1MGZTNWpiMjUwWVdsdUxXMTdZbUZqYTJkeWIzVnVaQzF6YVhwbE9tTnZiblJoYVc0aGFXMXdiM0owWVc1MGZTNWlaeTFqWlc1MFpYSXRiWHRpWVdOclozSnZkVzVrTFhCdmMybDBhVzl1T2pVd0pYMHVZbWN0WTJWdWRHVnlMVzBzTG1KbkxYUnZjQzF0ZTJKaFkydG5jbTkxYm1RdGNtVndaV0YwT201dkxYSmxjR1ZoZEgwdVltY3RkRzl3TFcxN1ltRmphMmR5YjNWdVpDMXdiM05wZEdsdmJqcDBiM0I5TG1KbkxYSnBaMmgwTFcxN1ltRmphMmR5YjNWdVpDMXdiM05wZEdsdmJqb3hNREFsZlM1aVp5MWliM1IwYjIwdGJTd3VZbWN0Y21sbmFIUXRiWHRpWVdOclozSnZkVzVrTFhKbGNHVmhkRHB1YnkxeVpYQmxZWFI5TG1KbkxXSnZkSFJ2YlMxdGUySmhZMnRuY205MWJtUXRjRzl6YVhScGIyNDZZbTkwZEc5dGZTNWlaeTFzWldaMExXMTdZbUZqYTJkeWIzVnVaQzF5WlhCbFlYUTZibTh0Y21Wd1pXRjBPMkpoWTJ0bmNtOTFibVF0Y0c5emFYUnBiMjQ2TUgwdWIzVjBiR2x1WlMxdGUyOTFkR3hwYm1VNk1YQjRJSE52Ykdsa2ZTNXZkWFJzYVc1bExYUnlZVzV6Y0dGeVpXNTBMVzE3YjNWMGJHbHVaVG94Y0hnZ2MyOXNhV1FnZEhKaGJuTndZWEpsYm5SOUxtOTFkR3hwYm1VdE1DMXRlMjkxZEd4cGJtVTZNSDB1WW1FdGJYdGliM0prWlhJdGMzUjViR1U2YzI5c2FXUTdZbTl5WkdWeUxYZHBaSFJvT2pGd2VIMHVZblF0Ylh0aWIzSmtaWEl0ZEc5d0xYTjBlV3hsT25OdmJHbGtPMkp2Y21SbGNpMTBiM0F0ZDJsa2RHZzZNWEI0ZlM1aWNpMXRlMkp2Y21SbGNpMXlhV2RvZEMxemRIbHNaVHB6YjJ4cFpEdGliM0prWlhJdGNtbG5hSFF0ZDJsa2RHZzZNWEI0ZlM1aVlpMXRlMkp2Y21SbGNpMWliM1IwYjIwdGMzUjViR1U2YzI5c2FXUTdZbTl5WkdWeUxXSnZkSFJ2YlMxM2FXUjBhRG94Y0hoOUxtSnNMVzE3WW05eVpHVnlMV3hsWm5RdGMzUjViR1U2YzI5c2FXUTdZbTl5WkdWeUxXeGxablF0ZDJsa2RHZzZNWEI0ZlM1aWJpMXRlMkp2Y21SbGNpMXpkSGxzWlRwdWIyNWxPMkp2Y21SbGNpMTNhV1IwYURvd2ZTNWljakF0Ylh0aWIzSmtaWEl0Y21Ga2FYVnpPakI5TG1KeU1TMXRlMkp2Y21SbGNpMXlZV1JwZFhNNkxqRXlOWEpsYlgwdVluSXlMVzE3WW05eVpHVnlMWEpoWkdsMWN6b3VNalZ5WlcxOUxtSnlNeTF0ZTJKdmNtUmxjaTF5WVdScGRYTTZMalZ5WlcxOUxtSnlOQzF0ZTJKdmNtUmxjaTF5WVdScGRYTTZNWEpsYlgwdVluSXRNVEF3TFcxN1ltOXlaR1Z5TFhKaFpHbDFjem94TURBbGZTNWljaTF3YVd4c0xXMTdZbTl5WkdWeUxYSmhaR2wxY3pvNU9UazVjSGg5TG1KeUxTMWliM1IwYjIwdGJYdGliM0prWlhJdGRHOXdMV3hsWm5RdGNtRmthWFZ6T2pBN1ltOXlaR1Z5TFhSdmNDMXlhV2RvZEMxeVlXUnBkWE02TUgwdVluSXRMWFJ2Y0MxdGUySnZjbVJsY2kxaWIzUjBiMjB0Y21sbmFIUXRjbUZrYVhWek9qQjlMbUp5TFMxeWFXZG9kQzF0TEM1aWNpMHRkRzl3TFcxN1ltOXlaR1Z5TFdKdmRIUnZiUzFzWldaMExYSmhaR2wxY3pvd2ZTNWljaTB0Y21sbmFIUXRiWHRpYjNKa1pYSXRkRzl3TFd4bFpuUXRjbUZrYVhWek9qQjlMbUp5TFMxc1pXWjBMVzE3WW05eVpHVnlMWFJ2Y0MxeWFXZG9kQzF5WVdScGRYTTZNRHRpYjNKa1pYSXRZbTkwZEc5dExYSnBaMmgwTFhKaFpHbDFjem93ZlM1aUxTMWtiM1IwWldRdGJYdGliM0prWlhJdGMzUjViR1U2Wkc5MGRHVmtmUzVpTFMxa1lYTm9aV1F0Ylh0aWIzSmtaWEl0YzNSNWJHVTZaR0Z6YUdWa2ZTNWlMUzF6YjJ4cFpDMXRlMkp2Y21SbGNpMXpkSGxzWlRwemIyeHBaSDB1WWkwdGJtOXVaUzF0ZTJKdmNtUmxjaTF6ZEhsc1pUcHViMjVsZlM1aWR6QXRiWHRpYjNKa1pYSXRkMmxrZEdnNk1IMHVZbmN4TFcxN1ltOXlaR1Z5TFhkcFpIUm9PaTR4TWpWeVpXMTlMbUozTWkxdGUySnZjbVJsY2kxM2FXUjBhRG91TWpWeVpXMTlMbUozTXkxdGUySnZjbVJsY2kxM2FXUjBhRG91TlhKbGJYMHVZbmMwTFcxN1ltOXlaR1Z5TFhkcFpIUm9PakZ5WlcxOUxtSjNOUzF0ZTJKdmNtUmxjaTEzYVdSMGFEb3ljbVZ0ZlM1aWRDMHdMVzE3WW05eVpHVnlMWFJ2Y0MxM2FXUjBhRG93ZlM1aWNpMHdMVzE3WW05eVpHVnlMWEpwWjJoMExYZHBaSFJvT2pCOUxtSmlMVEF0Ylh0aWIzSmtaWEl0WW05MGRHOXRMWGRwWkhSb09qQjlMbUpzTFRBdGJYdGliM0prWlhJdGJHVm1kQzEzYVdSMGFEb3dmUzV6YUdGa2IzY3RNUzF0ZTJKdmVDMXphR0ZrYjNjNk1DQXdJRFJ3ZUNBeWNIZ2djbWRpWVNnd0xEQXNNQ3d1TWlsOUxuTm9ZV1J2ZHkweUxXMTdZbTk0TFhOb1lXUnZkem93SURBZ09IQjRJREp3ZUNCeVoySmhLREFzTUN3d0xDNHlLWDB1YzJoaFpHOTNMVE10Ylh0aWIzZ3RjMmhoWkc5M09qSndlQ0F5Y0hnZ05IQjRJREp3ZUNCeVoySmhLREFzTUN3d0xDNHlLWDB1YzJoaFpHOTNMVFF0Ylh0aWIzZ3RjMmhoWkc5M09qSndlQ0F5Y0hnZ09IQjRJREFnY21kaVlTZ3dMREFzTUN3dU1pbDlMbk5vWVdSdmR5MDFMVzE3WW05NExYTm9ZV1J2ZHpvMGNIZ2dOSEI0SURod2VDQXdJSEpuWW1Fb01Dd3dMREFzTGpJcGZTNTBiM0F0TUMxdGUzUnZjRG93ZlM1c1pXWjBMVEF0Ylh0c1pXWjBPakI5TG5KcFoyaDBMVEF0Ylh0eWFXZG9kRG93ZlM1aWIzUjBiMjB0TUMxdGUySnZkSFJ2YlRvd2ZTNTBiM0F0TVMxdGUzUnZjRG94Y21WdGZTNXNaV1owTFRFdGJYdHNaV1owT2pGeVpXMTlMbkpwWjJoMExURXRiWHR5YVdkb2REb3hjbVZ0ZlM1aWIzUjBiMjB0TVMxdGUySnZkSFJ2YlRveGNtVnRmUzUwYjNBdE1pMXRlM1J2Y0RveWNtVnRmUzVzWldaMExUSXRiWHRzWldaME9qSnlaVzE5TG5KcFoyaDBMVEl0Ylh0eWFXZG9kRG95Y21WdGZTNWliM1IwYjIwdE1pMXRlMkp2ZEhSdmJUb3ljbVZ0ZlM1MGIzQXRMVEV0Ylh0MGIzQTZMVEZ5WlcxOUxuSnBaMmgwTFMweExXMTdjbWxuYUhRNkxURnlaVzE5TG1KdmRIUnZiUzB0TVMxdGUySnZkSFJ2YlRvdE1YSmxiWDB1YkdWbWRDMHRNUzF0ZTJ4bFpuUTZMVEZ5WlcxOUxuUnZjQzB0TWkxdGUzUnZjRG90TW5KbGJYMHVjbWxuYUhRdExUSXRiWHR5YVdkb2REb3RNbkpsYlgwdVltOTBkRzl0TFMweUxXMTdZbTkwZEc5dE9pMHljbVZ0ZlM1c1pXWjBMUzB5TFcxN2JHVm1kRG90TW5KbGJYMHVZV0p6YjJ4MWRHVXRMV1pwYkd3dGJYdDBiM0E2TUR0eWFXZG9kRG93TzJKdmRIUnZiVG93TzJ4bFpuUTZNSDB1WTJ3dGJYdGpiR1ZoY2pwc1pXWjBmUzVqY2kxdGUyTnNaV0Z5T25KcFoyaDBmUzVqWWkxdGUyTnNaV0Z5T21KdmRHaDlMbU51TFcxN1kyeGxZWEk2Ym05dVpYMHVaRzR0Ylh0a2FYTndiR0Y1T201dmJtVjlMbVJwTFcxN1pHbHpjR3hoZVRwcGJteHBibVY5TG1SaUxXMTdaR2x6Y0d4aGVUcGliRzlqYTMwdVpHbGlMVzE3WkdsemNHeGhlVHBwYm14cGJtVXRZbXh2WTJ0OUxtUnBkQzF0ZTJScGMzQnNZWGs2YVc1c2FXNWxMWFJoWW14bGZTNWtkQzF0ZTJScGMzQnNZWGs2ZEdGaWJHVjlMbVIwWXkxdGUyUnBjM0JzWVhrNmRHRmliR1V0WTJWc2JIMHVaSFF0Y205M0xXMTdaR2x6Y0d4aGVUcDBZV0pzWlMxeWIzZDlMbVIwTFhKdmR5MW5jbTkxY0MxdGUyUnBjM0JzWVhrNmRHRmliR1V0Y205M0xXZHliM1Z3ZlM1a2RDMWpiMngxYlc0dGJYdGthWE53YkdGNU9uUmhZbXhsTFdOdmJIVnRibjB1WkhRdFkyOXNkVzF1TFdkeWIzVndMVzE3WkdsemNHeGhlVHAwWVdKc1pTMWpiMngxYlc0dFozSnZkWEI5TG1SMExTMW1hWGhsWkMxdGUzUmhZbXhsTFd4aGVXOTFkRHBtYVhobFpEdDNhV1IwYURveE1EQWxmUzVtYkdWNExXMTdaR2x6Y0d4aGVUcG1iR1Y0ZlM1cGJteHBibVV0Wm14bGVDMXRlMlJwYzNCc1lYazZhVzVzYVc1bExXWnNaWGg5TG1ac1pYZ3RZWFYwYnkxdGUyWnNaWGc2TVNBeElHRjFkRzg3YldsdUxYZHBaSFJvT2pBN2JXbHVMV2hsYVdkb2REb3dmUzVtYkdWNExXNXZibVV0Ylh0bWJHVjRPbTV2Ym1WOUxtWnNaWGd0WTI5c2RXMXVMVzE3Wm14bGVDMWthWEpsWTNScGIyNDZZMjlzZFcxdWZTNW1iR1Y0TFhKdmR5MXRlMlpzWlhndFpHbHlaV04wYVc5dU9uSnZkMzB1Wm14bGVDMTNjbUZ3TFcxN1pteGxlQzEzY21Gd09uZHlZWEI5TG1ac1pYZ3RibTkzY21Gd0xXMTdabXhsZUMxM2NtRndPbTV2ZDNKaGNIMHVabXhsZUMxM2NtRndMWEpsZG1WeWMyVXRiWHRtYkdWNExYZHlZWEE2ZDNKaGNDMXlaWFpsY25ObGZTNW1iR1Y0TFdOdmJIVnRiaTF5WlhabGNuTmxMVzE3Wm14bGVDMWthWEpsWTNScGIyNDZZMjlzZFcxdUxYSmxkbVZ5YzJWOUxtWnNaWGd0Y205M0xYSmxkbVZ5YzJVdGJYdG1iR1Y0TFdScGNtVmpkR2x2YmpweWIzY3RjbVYyWlhKelpYMHVhWFJsYlhNdGMzUmhjblF0Ylh0aGJHbG5iaTFwZEdWdGN6cG1iR1Y0TFhOMFlYSjBmUzVwZEdWdGN5MWxibVF0Ylh0aGJHbG5iaTFwZEdWdGN6cG1iR1Y0TFdWdVpIMHVhWFJsYlhNdFkyVnVkR1Z5TFcxN1lXeHBaMjR0YVhSbGJYTTZZMlZ1ZEdWeWZTNXBkR1Z0Y3kxaVlYTmxiR2x1WlMxdGUyRnNhV2R1TFdsMFpXMXpPbUpoYzJWc2FXNWxmUzVwZEdWdGN5MXpkSEpsZEdOb0xXMTdZV3hwWjI0dGFYUmxiWE02YzNSeVpYUmphSDB1YzJWc1ppMXpkR0Z5ZEMxdGUyRnNhV2R1TFhObGJHWTZabXhsZUMxemRHRnlkSDB1YzJWc1ppMWxibVF0Ylh0aGJHbG5iaTF6Wld4bU9tWnNaWGd0Wlc1a2ZTNXpaV3htTFdObGJuUmxjaTF0ZTJGc2FXZHVMWE5sYkdZNlkyVnVkR1Z5ZlM1elpXeG1MV0poYzJWc2FXNWxMVzE3WVd4cFoyNHRjMlZzWmpwaVlYTmxiR2x1WlgwdWMyVnNaaTF6ZEhKbGRHTm9MVzE3WVd4cFoyNHRjMlZzWmpwemRISmxkR05vZlM1cWRYTjBhV1o1TFhOMFlYSjBMVzE3YW5WemRHbG1lUzFqYjI1MFpXNTBPbVpzWlhndGMzUmhjblI5TG1wMWMzUnBabmt0Wlc1a0xXMTdhblZ6ZEdsbWVTMWpiMjUwWlc1ME9tWnNaWGd0Wlc1a2ZTNXFkWE4wYVdaNUxXTmxiblJsY2kxdGUycDFjM1JwWm5rdFkyOXVkR1Z1ZERwalpXNTBaWEo5TG1wMWMzUnBabmt0WW1WMGQyVmxiaTF0ZTJwMWMzUnBabmt0WTI5dWRHVnVkRHB6Y0dGalpTMWlaWFIzWldWdWZTNXFkWE4wYVdaNUxXRnliM1Z1WkMxdGUycDFjM1JwWm5rdFkyOXVkR1Z1ZERwemNHRmpaUzFoY205MWJtUjlMbU52Ym5SbGJuUXRjM1JoY25RdGJYdGhiR2xuYmkxamIyNTBaVzUwT21ac1pYZ3RjM1JoY25SOUxtTnZiblJsYm5RdFpXNWtMVzE3WVd4cFoyNHRZMjl1ZEdWdWREcG1iR1Y0TFdWdVpIMHVZMjl1ZEdWdWRDMWpaVzUwWlhJdGJYdGhiR2xuYmkxamIyNTBaVzUwT21ObGJuUmxjbjB1WTI5dWRHVnVkQzFpWlhSM1pXVnVMVzE3WVd4cFoyNHRZMjl1ZEdWdWREcHpjR0ZqWlMxaVpYUjNaV1Z1ZlM1amIyNTBaVzUwTFdGeWIzVnVaQzF0ZTJGc2FXZHVMV052Ym5SbGJuUTZjM0JoWTJVdFlYSnZkVzVrZlM1amIyNTBaVzUwTFhOMGNtVjBZMmd0Ylh0aGJHbG5iaTFqYjI1MFpXNTBPbk4wY21WMFkyaDlMbTl5WkdWeUxUQXRiWHR2Y21SbGNqb3dmUzV2Y21SbGNpMHhMVzE3YjNKa1pYSTZNWDB1YjNKa1pYSXRNaTF0ZTI5eVpHVnlPako5TG05eVpHVnlMVE10Ylh0dmNtUmxjam96ZlM1dmNtUmxjaTAwTFcxN2IzSmtaWEk2TkgwdWIzSmtaWEl0TlMxdGUyOXlaR1Z5T2pWOUxtOXlaR1Z5TFRZdGJYdHZjbVJsY2pvMmZTNXZjbVJsY2kwM0xXMTdiM0prWlhJNk4zMHViM0prWlhJdE9DMXRlMjl5WkdWeU9qaDlMbTl5WkdWeUxXeGhjM1F0Ylh0dmNtUmxjam81T1RrNU9YMHVabXhsZUMxbmNtOTNMVEF0Ylh0bWJHVjRMV2R5YjNjNk1IMHVabXhsZUMxbmNtOTNMVEV0Ylh0bWJHVjRMV2R5YjNjNk1YMHVabXhsZUMxemFISnBibXN0TUMxdGUyWnNaWGd0YzJoeWFXNXJPakI5TG1ac1pYZ3RjMmh5YVc1ckxURXRiWHRtYkdWNExYTm9jbWx1YXpveGZTNW1iQzF0ZTJac2IyRjBPbXhsWm5SOUxtWnNMVzBzTG1aeUxXMTdYMlJwYzNCc1lYazZhVzVzYVc1bGZTNW1jaTF0ZTJac2IyRjBPbkpwWjJoMGZTNW1iaTF0ZTJac2IyRjBPbTV2Ym1WOUxta3RiWHRtYjI1MExYTjBlV3hsT21sMFlXeHBZMzB1Wm5NdGJtOXliV0ZzTFcxN1ptOXVkQzF6ZEhsc1pUcHViM0p0WVd4OUxtNXZjbTFoYkMxdGUyWnZiblF0ZDJWcFoyaDBPalF3TUgwdVlpMXRlMlp2Ym5RdGQyVnBaMmgwT2pjd01IMHVabmN4TFcxN1ptOXVkQzEzWldsbmFIUTZNVEF3ZlM1bWR6SXRiWHRtYjI1MExYZGxhV2RvZERveU1EQjlMbVozTXkxdGUyWnZiblF0ZDJWcFoyaDBPak13TUgwdVpuYzBMVzE3Wm05dWRDMTNaV2xuYUhRNk5EQXdmUzVtZHpVdGJYdG1iMjUwTFhkbGFXZG9kRG8xTURCOUxtWjNOaTF0ZTJadmJuUXRkMlZwWjJoME9qWXdNSDB1Wm5jM0xXMTdabTl1ZEMxM1pXbG5hSFE2TnpBd2ZTNW1kemd0Ylh0bWIyNTBMWGRsYVdkb2REbzRNREI5TG1aM09TMXRlMlp2Ym5RdGQyVnBaMmgwT2prd01IMHVhREV0Ylh0b1pXbG5hSFE2TVhKbGJYMHVhREl0Ylh0b1pXbG5hSFE2TW5KbGJYMHVhRE10Ylh0b1pXbG5hSFE2TkhKbGJYMHVhRFF0Ylh0b1pXbG5hSFE2T0hKbGJYMHVhRFV0Ylh0b1pXbG5hSFE2TVRaeVpXMTlMbWd0TWpVdGJYdG9aV2xuYUhRNk1qVWxmUzVvTFRVd0xXMTdhR1ZwWjJoME9qVXdKWDB1YUMwM05TMXRlMmhsYVdkb2REbzNOU1Y5TG1ndE1UQXdMVzE3YUdWcFoyaDBPakV3TUNWOUxtMXBiaTFvTFRFd01DMXRlMjFwYmkxb1pXbG5hSFE2TVRBd0pYMHVkbWd0TWpVdGJYdG9aV2xuYUhRNk1qVjJhSDB1ZG1ndE5UQXRiWHRvWldsbmFIUTZOVEIyYUgwdWRtZ3ROelV0Ylh0b1pXbG5hSFE2TnpWMmFIMHVkbWd0TVRBd0xXMTdhR1ZwWjJoME9qRXdNSFpvZlM1dGFXNHRkbWd0TVRBd0xXMTdiV2x1TFdobGFXZG9kRG94TURCMmFIMHVhQzFoZFhSdkxXMTdhR1ZwWjJoME9tRjFkRzk5TG1ndGFXNW9aWEpwZEMxdGUyaGxhV2RvZERwcGJtaGxjbWwwZlM1MGNtRmphMlZrTFcxN2JHVjBkR1Z5TFhOd1lXTnBibWM2TGpGbGJYMHVkSEpoWTJ0bFpDMTBhV2RvZEMxdGUyeGxkSFJsY2kxemNHRmphVzVuT2kwdU1EVmxiWDB1ZEhKaFkydGxaQzF0WldkaExXMTdiR1YwZEdWeUxYTndZV05wYm1jNkxqSTFaVzE5TG14b0xYTnZiR2xrTFcxN2JHbHVaUzFvWldsbmFIUTZNWDB1YkdndGRHbDBiR1V0Ylh0c2FXNWxMV2hsYVdkb2REb3hMakkxZlM1c2FDMWpiM0I1TFcxN2JHbHVaUzFvWldsbmFIUTZNUzQxZlM1dGR5MHhNREF0Ylh0dFlYZ3RkMmxrZEdnNk1UQXdKWDB1YlhjeExXMTdiV0Y0TFhkcFpIUm9PakZ5WlcxOUxtMTNNaTF0ZTIxaGVDMTNhV1IwYURveWNtVnRmUzV0ZHpNdGJYdHRZWGd0ZDJsa2RHZzZOSEpsYlgwdWJYYzBMVzE3YldGNExYZHBaSFJvT2poeVpXMTlMbTEzTlMxdGUyMWhlQzEzYVdSMGFEb3hObkpsYlgwdWJYYzJMVzE3YldGNExYZHBaSFJvT2pNeWNtVnRmUzV0ZHpjdGJYdHRZWGd0ZDJsa2RHZzZORGh5WlcxOUxtMTNPQzF0ZTIxaGVDMTNhV1IwYURvMk5ISmxiWDB1YlhjNUxXMTdiV0Y0TFhkcFpIUm9PamsyY21WdGZTNXRkeTF1YjI1bExXMTdiV0Y0TFhkcFpIUm9PbTV2Ym1WOUxuY3hMVzE3ZDJsa2RHZzZNWEpsYlgwdWR6SXRiWHQzYVdSMGFEb3ljbVZ0ZlM1M015MXRlM2RwWkhSb09qUnlaVzE5TG5jMExXMTdkMmxrZEdnNk9ISmxiWDB1ZHpVdGJYdDNhV1IwYURveE5uSmxiWDB1ZHkweE1DMXRlM2RwWkhSb09qRXdKWDB1ZHkweU1DMXRlM2RwWkhSb09qSXdKWDB1ZHkweU5TMXRlM2RwWkhSb09qSTFKWDB1ZHkwek1DMXRlM2RwWkhSb09qTXdKWDB1ZHkwek15MXRlM2RwWkhSb09qTXpKWDB1ZHkwek5DMXRlM2RwWkhSb09qTTBKWDB1ZHkwME1DMXRlM2RwWkhSb09qUXdKWDB1ZHkwMU1DMXRlM2RwWkhSb09qVXdKWDB1ZHkwMk1DMXRlM2RwWkhSb09qWXdKWDB1ZHkwM01DMXRlM2RwWkhSb09qY3dKWDB1ZHkwM05TMXRlM2RwWkhSb09qYzFKWDB1ZHkwNE1DMXRlM2RwWkhSb09qZ3dKWDB1ZHkwNU1DMXRlM2RwWkhSb09qa3dKWDB1ZHkweE1EQXRiWHQzYVdSMGFEb3hNREFsZlM1M0xYUm9hWEprTFcxN2QybGtkR2c2TXpNdU16TXpNek1sZlM1M0xYUjNieTEwYUdseVpITXRiWHQzYVdSMGFEbzJOaTQyTmpZMk55VjlMbmN0WVhWMGJ5MXRlM2RwWkhSb09tRjFkRzk5TG05MlpYSm1iRzkzTFhacGMybGliR1V0Ylh0dmRtVnlabXh2ZHpwMmFYTnBZbXhsZlM1dmRtVnlabXh2ZHkxb2FXUmtaVzR0Ylh0dmRtVnlabXh2ZHpwb2FXUmtaVzU5TG05MlpYSm1iRzkzTFhOamNtOXNiQzF0ZTI5MlpYSm1iRzkzT25OamNtOXNiSDB1YjNabGNtWnNiM2N0WVhWMGJ5MXRlMjkyWlhKbWJHOTNPbUYxZEc5OUxtOTJaWEptYkc5M0xYZ3RkbWx6YVdKc1pTMXRlMjkyWlhKbWJHOTNMWGc2ZG1semFXSnNaWDB1YjNabGNtWnNiM2N0ZUMxb2FXUmtaVzR0Ylh0dmRtVnlabXh2ZHkxNE9taHBaR1JsYm4wdWIzWmxjbVpzYjNjdGVDMXpZM0p2Ykd3dGJYdHZkbVZ5Wm14dmR5MTRPbk5qY205c2JIMHViM1psY21ac2IzY3RlQzFoZFhSdkxXMTdiM1psY21ac2IzY3RlRHBoZFhSdmZTNXZkbVZ5Wm14dmR5MTVMWFpwYzJsaWJHVXRiWHR2ZG1WeVpteHZkeTE1T25acGMybGliR1Y5TG05MlpYSm1iRzkzTFhrdGFHbGtaR1Z1TFcxN2IzWmxjbVpzYjNjdGVUcG9hV1JrWlc1OUxtOTJaWEptYkc5M0xYa3RjMk55YjJ4c0xXMTdiM1psY21ac2IzY3RlVHB6WTNKdmJHeDlMbTkyWlhKbWJHOTNMWGt0WVhWMGJ5MXRlMjkyWlhKbWJHOTNMWGs2WVhWMGIzMHVjM1JoZEdsakxXMTdjRzl6YVhScGIyNDZjM1JoZEdsamZTNXlaV3hoZEdsMlpTMXRlM0J2YzJsMGFXOXVPbkpsYkdGMGFYWmxmUzVoWW5OdmJIVjBaUzF0ZTNCdmMybDBhVzl1T21GaWMyOXNkWFJsZlM1bWFYaGxaQzF0ZTNCdmMybDBhVzl1T21acGVHVmtmUzV5YjNSaGRHVXRORFV0YlhzdGQyVmlhMmwwTFhSeVlXNXpabTl5YlRweWIzUmhkR1VvTkRWa1pXY3BPM1J5WVc1elptOXliVHB5YjNSaGRHVW9ORFZrWldjcGZTNXliM1JoZEdVdE9UQXRiWHN0ZDJWaWEybDBMWFJ5WVc1elptOXliVHB5YjNSaGRHVW9PVEJrWldjcE8zUnlZVzV6Wm05eWJUcHliM1JoZEdVb09UQmtaV2NwZlM1eWIzUmhkR1V0TVRNMUxXMTdMWGRsWW10cGRDMTBjbUZ1YzJadmNtMDZjbTkwWVhSbEtERXpOV1JsWnlrN2RISmhibk5tYjNKdE9uSnZkR0YwWlNneE16VmtaV2NwZlM1eWIzUmhkR1V0TVRnd0xXMTdMWGRsWW10cGRDMTBjbUZ1YzJadmNtMDZjbTkwWVhSbEtERTRNR1JsWnlrN2RISmhibk5tYjNKdE9uSnZkR0YwWlNneE9EQmtaV2NwZlM1eWIzUmhkR1V0TWpJMUxXMTdMWGRsWW10cGRDMTBjbUZ1YzJadmNtMDZjbTkwWVhSbEtESXlOV1JsWnlrN2RISmhibk5tYjNKdE9uSnZkR0YwWlNneU1qVmtaV2NwZlM1eWIzUmhkR1V0TWpjd0xXMTdMWGRsWW10cGRDMTBjbUZ1YzJadmNtMDZjbTkwWVhSbEtESTNNR1JsWnlrN2RISmhibk5tYjNKdE9uSnZkR0YwWlNneU56QmtaV2NwZlM1eWIzUmhkR1V0TXpFMUxXMTdMWGRsWW10cGRDMTBjbUZ1YzJadmNtMDZjbTkwWVhSbEtETXhOV1JsWnlrN2RISmhibk5tYjNKdE9uSnZkR0YwWlNnek1UVmtaV2NwZlM1d1lUQXRiWHR3WVdSa2FXNW5PakI5TG5CaE1TMXRlM0JoWkdScGJtYzZMakkxY21WdGZTNXdZVEl0Ylh0d1lXUmthVzVuT2k0MWNtVnRmUzV3WVRNdGJYdHdZV1JrYVc1bk9qRnlaVzE5TG5CaE5DMXRlM0JoWkdScGJtYzZNbkpsYlgwdWNHRTFMVzE3Y0dGa1pHbHVaem8wY21WdGZTNXdZVFl0Ylh0d1lXUmthVzVuT2poeVpXMTlMbkJoTnkxdGUzQmhaR1JwYm1jNk1UWnlaVzE5TG5Cc01DMXRlM0JoWkdScGJtY3RiR1ZtZERvd2ZTNXdiREV0Ylh0d1lXUmthVzVuTFd4bFpuUTZMakkxY21WdGZTNXdiREl0Ylh0d1lXUmthVzVuTFd4bFpuUTZMalZ5WlcxOUxuQnNNeTF0ZTNCaFpHUnBibWN0YkdWbWREb3hjbVZ0ZlM1d2JEUXRiWHR3WVdSa2FXNW5MV3hsWm5RNk1uSmxiWDB1Y0d3MUxXMTdjR0ZrWkdsdVp5MXNaV1owT2pSeVpXMTlMbkJzTmkxdGUzQmhaR1JwYm1jdGJHVm1kRG80Y21WdGZTNXdiRGN0Ylh0d1lXUmthVzVuTFd4bFpuUTZNVFp5WlcxOUxuQnlNQzF0ZTNCaFpHUnBibWN0Y21sbmFIUTZNSDB1Y0hJeExXMTdjR0ZrWkdsdVp5MXlhV2RvZERvdU1qVnlaVzE5TG5CeU1pMXRlM0JoWkdScGJtY3RjbWxuYUhRNkxqVnlaVzE5TG5CeU15MXRlM0JoWkdScGJtY3RjbWxuYUhRNk1YSmxiWDB1Y0hJMExXMTdjR0ZrWkdsdVp5MXlhV2RvZERveWNtVnRmUzV3Y2pVdGJYdHdZV1JrYVc1bkxYSnBaMmgwT2pSeVpXMTlMbkJ5TmkxdGUzQmhaR1JwYm1jdGNtbG5hSFE2T0hKbGJYMHVjSEkzTFcxN2NHRmtaR2x1WnkxeWFXZG9kRG94Tm5KbGJYMHVjR0l3TFcxN2NHRmtaR2x1WnkxaWIzUjBiMjA2TUgwdWNHSXhMVzE3Y0dGa1pHbHVaeTFpYjNSMGIyMDZMakkxY21WdGZTNXdZakl0Ylh0d1lXUmthVzVuTFdKdmRIUnZiVG91TlhKbGJYMHVjR0l6TFcxN2NHRmtaR2x1WnkxaWIzUjBiMjA2TVhKbGJYMHVjR0kwTFcxN2NHRmtaR2x1WnkxaWIzUjBiMjA2TW5KbGJYMHVjR0kxTFcxN2NHRmtaR2x1WnkxaWIzUjBiMjA2TkhKbGJYMHVjR0kyTFcxN2NHRmtaR2x1WnkxaWIzUjBiMjA2T0hKbGJYMHVjR0kzTFcxN2NHRmtaR2x1WnkxaWIzUjBiMjA2TVRaeVpXMTlMbkIwTUMxdGUzQmhaR1JwYm1jdGRHOXdPakI5TG5CME1TMXRlM0JoWkdScGJtY3RkRzl3T2k0eU5YSmxiWDB1Y0hReUxXMTdjR0ZrWkdsdVp5MTBiM0E2TGpWeVpXMTlMbkIwTXkxdGUzQmhaR1JwYm1jdGRHOXdPakZ5WlcxOUxuQjBOQzF0ZTNCaFpHUnBibWN0ZEc5d09qSnlaVzE5TG5CME5TMXRlM0JoWkdScGJtY3RkRzl3T2pSeVpXMTlMbkIwTmkxdGUzQmhaR1JwYm1jdGRHOXdPamh5WlcxOUxuQjBOeTF0ZTNCaFpHUnBibWN0ZEc5d09qRTJjbVZ0ZlM1d2RqQXRiWHR3WVdSa2FXNW5MWFJ2Y0Rvd08zQmhaR1JwYm1jdFltOTBkRzl0T2pCOUxuQjJNUzF0ZTNCaFpHUnBibWN0ZEc5d09pNHlOWEpsYlR0d1lXUmthVzVuTFdKdmRIUnZiVG91TWpWeVpXMTlMbkIyTWkxdGUzQmhaR1JwYm1jdGRHOXdPaTQxY21WdE8zQmhaR1JwYm1jdFltOTBkRzl0T2k0MWNtVnRmUzV3ZGpNdGJYdHdZV1JrYVc1bkxYUnZjRG94Y21WdE8zQmhaR1JwYm1jdFltOTBkRzl0T2pGeVpXMTlMbkIyTkMxdGUzQmhaR1JwYm1jdGRHOXdPakp5WlcwN2NHRmtaR2x1WnkxaWIzUjBiMjA2TW5KbGJYMHVjSFkxTFcxN2NHRmtaR2x1WnkxMGIzQTZOSEpsYlR0d1lXUmthVzVuTFdKdmRIUnZiVG8wY21WdGZTNXdkall0Ylh0d1lXUmthVzVuTFhSdmNEbzRjbVZ0TzNCaFpHUnBibWN0WW05MGRHOXRPamh5WlcxOUxuQjJOeTF0ZTNCaFpHUnBibWN0ZEc5d09qRTJjbVZ0TzNCaFpHUnBibWN0WW05MGRHOXRPakUyY21WdGZTNXdhREF0Ylh0d1lXUmthVzVuTFd4bFpuUTZNRHR3WVdSa2FXNW5MWEpwWjJoME9qQjlMbkJvTVMxdGUzQmhaR1JwYm1jdGJHVm1kRG91TWpWeVpXMDdjR0ZrWkdsdVp5MXlhV2RvZERvdU1qVnlaVzE5TG5Cb01pMXRlM0JoWkdScGJtY3RiR1ZtZERvdU5YSmxiVHR3WVdSa2FXNW5MWEpwWjJoME9pNDFjbVZ0ZlM1d2FETXRiWHR3WVdSa2FXNW5MV3hsWm5RNk1YSmxiVHR3WVdSa2FXNW5MWEpwWjJoME9qRnlaVzE5TG5Cb05DMXRlM0JoWkdScGJtY3RiR1ZtZERveWNtVnRPM0JoWkdScGJtY3RjbWxuYUhRNk1uSmxiWDB1Y0dnMUxXMTdjR0ZrWkdsdVp5MXNaV1owT2pSeVpXMDdjR0ZrWkdsdVp5MXlhV2RvZERvMGNtVnRmUzV3YURZdGJYdHdZV1JrYVc1bkxXeGxablE2T0hKbGJUdHdZV1JrYVc1bkxYSnBaMmgwT2poeVpXMTlMbkJvTnkxdGUzQmhaR1JwYm1jdGJHVm1kRG94Tm5KbGJUdHdZV1JrYVc1bkxYSnBaMmgwT2pFMmNtVnRmUzV0WVRBdGJYdHRZWEpuYVc0Nk1IMHViV0V4TFcxN2JXRnlaMmx1T2k0eU5YSmxiWDB1YldFeUxXMTdiV0Z5WjJsdU9pNDFjbVZ0ZlM1dFlUTXRiWHR0WVhKbmFXNDZNWEpsYlgwdWJXRTBMVzE3YldGeVoybHVPakp5WlcxOUxtMWhOUzF0ZTIxaGNtZHBiam8wY21WdGZTNXRZVFl0Ylh0dFlYSm5hVzQ2T0hKbGJYMHViV0UzTFcxN2JXRnlaMmx1T2pFMmNtVnRmUzV0YkRBdGJYdHRZWEpuYVc0dGJHVm1kRG93ZlM1dGJERXRiWHR0WVhKbmFXNHRiR1ZtZERvdU1qVnlaVzE5TG0xc01pMXRlMjFoY21kcGJpMXNaV1owT2k0MWNtVnRmUzV0YkRNdGJYdHRZWEpuYVc0dGJHVm1kRG94Y21WdGZTNXRiRFF0Ylh0dFlYSm5hVzR0YkdWbWREb3ljbVZ0ZlM1dGJEVXRiWHR0WVhKbmFXNHRiR1ZtZERvMGNtVnRmUzV0YkRZdGJYdHRZWEpuYVc0dGJHVm1kRG80Y21WdGZTNXRiRGN0Ylh0dFlYSm5hVzR0YkdWbWREb3hObkpsYlgwdWJYSXdMVzE3YldGeVoybHVMWEpwWjJoME9qQjlMbTF5TVMxdGUyMWhjbWRwYmkxeWFXZG9kRG91TWpWeVpXMTlMbTF5TWkxdGUyMWhjbWRwYmkxeWFXZG9kRG91TlhKbGJYMHViWEl6TFcxN2JXRnlaMmx1TFhKcFoyaDBPakZ5WlcxOUxtMXlOQzF0ZTIxaGNtZHBiaTF5YVdkb2REb3ljbVZ0ZlM1dGNqVXRiWHR0WVhKbmFXNHRjbWxuYUhRNk5ISmxiWDB1YlhJMkxXMTdiV0Z5WjJsdUxYSnBaMmgwT2poeVpXMTlMbTF5TnkxdGUyMWhjbWRwYmkxeWFXZG9kRG94Tm5KbGJYMHViV0l3TFcxN2JXRnlaMmx1TFdKdmRIUnZiVG93ZlM1dFlqRXRiWHR0WVhKbmFXNHRZbTkwZEc5dE9pNHlOWEpsYlgwdWJXSXlMVzE3YldGeVoybHVMV0p2ZEhSdmJUb3VOWEpsYlgwdWJXSXpMVzE3YldGeVoybHVMV0p2ZEhSdmJUb3hjbVZ0ZlM1dFlqUXRiWHR0WVhKbmFXNHRZbTkwZEc5dE9qSnlaVzE5TG0xaU5TMXRlMjFoY21kcGJpMWliM1IwYjIwNk5ISmxiWDB1YldJMkxXMTdiV0Z5WjJsdUxXSnZkSFJ2YlRvNGNtVnRmUzV0WWpjdGJYdHRZWEpuYVc0dFltOTBkRzl0T2pFMmNtVnRmUzV0ZERBdGJYdHRZWEpuYVc0dGRHOXdPakI5TG0xME1TMXRlMjFoY21kcGJpMTBiM0E2TGpJMWNtVnRmUzV0ZERJdGJYdHRZWEpuYVc0dGRHOXdPaTQxY21WdGZTNXRkRE10Ylh0dFlYSm5hVzR0ZEc5d09qRnlaVzE5TG0xME5DMXRlMjFoY21kcGJpMTBiM0E2TW5KbGJYMHViWFExTFcxN2JXRnlaMmx1TFhSdmNEbzBjbVZ0ZlM1dGREWXRiWHR0WVhKbmFXNHRkRzl3T2poeVpXMTlMbTEwTnkxdGUyMWhjbWRwYmkxMGIzQTZNVFp5WlcxOUxtMTJNQzF0ZTIxaGNtZHBiaTEwYjNBNk1EdHRZWEpuYVc0dFltOTBkRzl0T2pCOUxtMTJNUzF0ZTIxaGNtZHBiaTEwYjNBNkxqSTFjbVZ0TzIxaGNtZHBiaTFpYjNSMGIyMDZMakkxY21WdGZTNXRkakl0Ylh0dFlYSm5hVzR0ZEc5d09pNDFjbVZ0TzIxaGNtZHBiaTFpYjNSMGIyMDZMalZ5WlcxOUxtMTJNeTF0ZTIxaGNtZHBiaTEwYjNBNk1YSmxiVHR0WVhKbmFXNHRZbTkwZEc5dE9qRnlaVzE5TG0xMk5DMXRlMjFoY21kcGJpMTBiM0E2TW5KbGJUdHRZWEpuYVc0dFltOTBkRzl0T2pKeVpXMTlMbTEyTlMxdGUyMWhjbWRwYmkxMGIzQTZOSEpsYlR0dFlYSm5hVzR0WW05MGRHOXRPalJ5WlcxOUxtMTJOaTF0ZTIxaGNtZHBiaTEwYjNBNk9ISmxiVHR0WVhKbmFXNHRZbTkwZEc5dE9qaHlaVzE5TG0xMk55MXRlMjFoY21kcGJpMTBiM0E2TVRaeVpXMDdiV0Z5WjJsdUxXSnZkSFJ2YlRveE5uSmxiWDB1Yldnd0xXMTdiV0Z5WjJsdUxXeGxablE2TUR0dFlYSm5hVzR0Y21sbmFIUTZNSDB1YldneExXMTdiV0Z5WjJsdUxXeGxablE2TGpJMWNtVnRPMjFoY21kcGJpMXlhV2RvZERvdU1qVnlaVzE5TG0xb01pMXRlMjFoY21kcGJpMXNaV1owT2k0MWNtVnRPMjFoY21kcGJpMXlhV2RvZERvdU5YSmxiWDB1YldnekxXMTdiV0Z5WjJsdUxXeGxablE2TVhKbGJUdHRZWEpuYVc0dGNtbG5hSFE2TVhKbGJYMHViV2cwTFcxN2JXRnlaMmx1TFd4bFpuUTZNbkpsYlR0dFlYSm5hVzR0Y21sbmFIUTZNbkpsYlgwdWJXZzFMVzE3YldGeVoybHVMV3hsWm5RNk5ISmxiVHR0WVhKbmFXNHRjbWxuYUhRNk5ISmxiWDB1YldnMkxXMTdiV0Z5WjJsdUxXeGxablE2T0hKbGJUdHRZWEpuYVc0dGNtbG5hSFE2T0hKbGJYMHViV2czTFcxN2JXRnlaMmx1TFd4bFpuUTZNVFp5WlcwN2JXRnlaMmx1TFhKcFoyaDBPakUyY21WdGZTNXVZVEV0Ylh0dFlYSm5hVzQ2TFM0eU5YSmxiWDB1Ym1FeUxXMTdiV0Z5WjJsdU9pMHVOWEpsYlgwdWJtRXpMVzE3YldGeVoybHVPaTB4Y21WdGZTNXVZVFF0Ylh0dFlYSm5hVzQ2TFRKeVpXMTlMbTVoTlMxdGUyMWhjbWRwYmpvdE5ISmxiWDB1Ym1FMkxXMTdiV0Z5WjJsdU9pMDRjbVZ0ZlM1dVlUY3RiWHR0WVhKbmFXNDZMVEUyY21WdGZTNXViREV0Ylh0dFlYSm5hVzR0YkdWbWREb3RMakkxY21WdGZTNXViREl0Ylh0dFlYSm5hVzR0YkdWbWREb3RMalZ5WlcxOUxtNXNNeTF0ZTIxaGNtZHBiaTFzWldaME9pMHhjbVZ0ZlM1dWJEUXRiWHR0WVhKbmFXNHRiR1ZtZERvdE1uSmxiWDB1Ym13MUxXMTdiV0Z5WjJsdUxXeGxablE2TFRSeVpXMTlMbTVzTmkxdGUyMWhjbWRwYmkxc1pXWjBPaTA0Y21WdGZTNXViRGN0Ylh0dFlYSm5hVzR0YkdWbWREb3RNVFp5WlcxOUxtNXlNUzF0ZTIxaGNtZHBiaTF5YVdkb2REb3RMakkxY21WdGZTNXVjakl0Ylh0dFlYSm5hVzR0Y21sbmFIUTZMUzQxY21WdGZTNXVjak10Ylh0dFlYSm5hVzR0Y21sbmFIUTZMVEZ5WlcxOUxtNXlOQzF0ZTIxaGNtZHBiaTF5YVdkb2REb3RNbkpsYlgwdWJuSTFMVzE3YldGeVoybHVMWEpwWjJoME9pMDBjbVZ0ZlM1dWNqWXRiWHR0WVhKbmFXNHRjbWxuYUhRNkxUaHlaVzE5TG01eU55MXRlMjFoY21kcGJpMXlhV2RvZERvdE1UWnlaVzE5TG01aU1TMXRlMjFoY21kcGJpMWliM1IwYjIwNkxTNHlOWEpsYlgwdWJtSXlMVzE3YldGeVoybHVMV0p2ZEhSdmJUb3RMalZ5WlcxOUxtNWlNeTF0ZTIxaGNtZHBiaTFpYjNSMGIyMDZMVEZ5WlcxOUxtNWlOQzF0ZTIxaGNtZHBiaTFpYjNSMGIyMDZMVEp5WlcxOUxtNWlOUzF0ZTIxaGNtZHBiaTFpYjNSMGIyMDZMVFJ5WlcxOUxtNWlOaTF0ZTIxaGNtZHBiaTFpYjNSMGIyMDZMVGh5WlcxOUxtNWlOeTF0ZTIxaGNtZHBiaTFpYjNSMGIyMDZMVEUyY21WdGZTNXVkREV0Ylh0dFlYSm5hVzR0ZEc5d09pMHVNalZ5WlcxOUxtNTBNaTF0ZTIxaGNtZHBiaTEwYjNBNkxTNDFjbVZ0ZlM1dWRETXRiWHR0WVhKbmFXNHRkRzl3T2kweGNtVnRmUzV1ZERRdGJYdHRZWEpuYVc0dGRHOXdPaTB5Y21WdGZTNXVkRFV0Ylh0dFlYSm5hVzR0ZEc5d09pMDBjbVZ0ZlM1dWREWXRiWHR0WVhKbmFXNHRkRzl3T2kwNGNtVnRmUzV1ZERjdGJYdHRZWEpuYVc0dGRHOXdPaTB4Tm5KbGJYMHVjM1J5YVd0bExXMTdkR1Y0ZEMxa1pXTnZjbUYwYVc5dU9teHBibVV0ZEdoeWIzVm5hSDB1ZFc1a1pYSnNhVzVsTFcxN2RHVjRkQzFrWldOdmNtRjBhVzl1T25WdVpHVnliR2x1WlgwdWJtOHRkVzVrWlhKc2FXNWxMVzE3ZEdWNGRDMWtaV052Y21GMGFXOXVPbTV2Ym1WOUxuUnNMVzE3ZEdWNGRDMWhiR2xuYmpwc1pXWjBmUzUwY2kxdGUzUmxlSFF0WVd4cFoyNDZjbWxuYUhSOUxuUmpMVzE3ZEdWNGRDMWhiR2xuYmpwalpXNTBaWEo5TG5ScUxXMTdkR1Y0ZEMxaGJHbG5ianBxZFhOMGFXWjVmUzUwZEdNdGJYdDBaWGgwTFhSeVlXNXpabTl5YlRwallYQnBkR0ZzYVhwbGZTNTBkR3d0Ylh0MFpYaDBMWFJ5WVc1elptOXliVHBzYjNkbGNtTmhjMlY5TG5SMGRTMXRlM1JsZUhRdGRISmhibk5tYjNKdE9uVndjR1Z5WTJGelpYMHVkSFJ1TFcxN2RHVjRkQzEwY21GdWMyWnZjbTA2Ym05dVpYMHVaaTAyTFcwc0xtWXRhR1ZoWkd4cGJtVXRiWHRtYjI1MExYTnBlbVU2Tm5KbGJYMHVaaTAxTFcwc0xtWXRjM1ZpYUdWaFpHeHBibVV0Ylh0bWIyNTBMWE5wZW1VNk5YSmxiWDB1WmpFdGJYdG1iMjUwTFhOcGVtVTZNM0psYlgwdVpqSXRiWHRtYjI1MExYTnBlbVU2TWk0eU5YSmxiWDB1WmpNdGJYdG1iMjUwTFhOcGVtVTZNUzQxY21WdGZTNW1OQzF0ZTJadmJuUXRjMmw2WlRveExqSTFjbVZ0ZlM1bU5TMXRlMlp2Ym5RdGMybDZaVG94Y21WdGZTNW1OaTF0ZTJadmJuUXRjMmw2WlRvdU9EYzFjbVZ0ZlM1bU55MXRlMlp2Ym5RdGMybDZaVG91TnpWeVpXMTlMbTFsWVhOMWNtVXRiWHR0WVhndGQybGtkR2c2TXpCbGJYMHViV1ZoYzNWeVpTMTNhV1JsTFcxN2JXRjRMWGRwWkhSb09qTTBaVzE5TG0xbFlYTjFjbVV0Ym1GeWNtOTNMVzE3YldGNExYZHBaSFJvT2pJd1pXMTlMbWx1WkdWdWRDMXRlM1JsZUhRdGFXNWtaVzUwT2pGbGJUdHRZWEpuYVc0dGRHOXdPakE3YldGeVoybHVMV0p2ZEhSdmJUb3dmUzV6YldGc2JDMWpZWEJ6TFcxN1ptOXVkQzEyWVhKcFlXNTBPbk50WVd4c0xXTmhjSE45TG5SeWRXNWpZWFJsTFcxN2QyaHBkR1V0YzNCaFkyVTZibTkzY21Gd08yOTJaWEptYkc5M09taHBaR1JsYmp0MFpYaDBMVzkyWlhKbWJHOTNPbVZzYkdsd2MybHpmUzVqWlc1MFpYSXRiWHR0WVhKbmFXNHRiR1ZtZERwaGRYUnZmUzVqWlc1MFpYSXRiU3d1YlhJdFlYVjBieTF0ZTIxaGNtZHBiaTF5YVdkb2REcGhkWFJ2ZlM1dGJDMWhkWFJ2TFcxN2JXRnlaMmx1TFd4bFpuUTZZWFYwYjMwdVkyeHBjQzF0ZTNCdmMybDBhVzl1T21acGVHVmtJV2x0Y0c5eWRHRnVkRHRmY0c5emFYUnBiMjQ2WVdKemIyeDFkR1VoYVcxd2IzSjBZVzUwTzJOc2FYQTZjbVZqZENneGNIZ2dNWEI0SURGd2VDQXhjSGdwTzJOc2FYQTZjbVZqZENneGNIZ3NNWEI0TERGd2VDd3hjSGdwZlM1M2N5MXViM0p0WVd3dGJYdDNhR2wwWlMxemNHRmpaVHB1YjNKdFlXeDlMbTV2ZDNKaGNDMXRlM2RvYVhSbExYTndZV05sT201dmQzSmhjSDB1Y0hKbExXMTdkMmhwZEdVdGMzQmhZMlU2Y0hKbGZTNTJMV0poYzJVdGJYdDJaWEowYVdOaGJDMWhiR2xuYmpwaVlYTmxiR2x1WlgwdWRpMXRhV1F0Ylh0MlpYSjBhV05oYkMxaGJHbG5ianB0YVdSa2JHVjlMbll0ZEc5d0xXMTdkbVZ5ZEdsallXd3RZV3hwWjI0NmRHOXdmUzUyTFdKMGJTMXRlM1psY25ScFkyRnNMV0ZzYVdkdU9tSnZkSFJ2YlgxOVFHMWxaR2xoSUhOamNtVmxiaUJoYm1RZ0tHMXBiaTEzYVdSMGFEbzJNR1Z0S1hzdVlYTndaV04wTFhKaGRHbHZMV3g3YUdWcFoyaDBPakE3Y0c5emFYUnBiMjQ2Y21Wc1lYUnBkbVY5TG1GemNHVmpkQzF5WVhScGJ5MHRNVFo0T1Mxc2UzQmhaR1JwYm1jdFltOTBkRzl0T2pVMkxqSTFKWDB1WVhOd1pXTjBMWEpoZEdsdkxTMDVlREUyTFd4N2NHRmtaR2x1WnkxaWIzUjBiMjA2TVRjM0xqYzNKWDB1WVhOd1pXTjBMWEpoZEdsdkxTMDBlRE10Ykh0d1lXUmthVzVuTFdKdmRIUnZiVG8zTlNWOUxtRnpjR1ZqZEMxeVlYUnBieTB0TTNnMExXeDdjR0ZrWkdsdVp5MWliM1IwYjIwNk1UTXpMak16SlgwdVlYTndaV04wTFhKaGRHbHZMUzAyZURRdGJIdHdZV1JrYVc1bkxXSnZkSFJ2YlRvMk5pNDJKWDB1WVhOd1pXTjBMWEpoZEdsdkxTMDBlRFl0Ykh0d1lXUmthVzVuTFdKdmRIUnZiVG94TlRBbGZTNWhjM0JsWTNRdGNtRjBhVzh0TFRoNE5TMXNlM0JoWkdScGJtY3RZbTkwZEc5dE9qWXlMalVsZlM1aGMzQmxZM1F0Y21GMGFXOHRMVFY0T0Mxc2UzQmhaR1JwYm1jdFltOTBkRzl0T2pFMk1DVjlMbUZ6Y0dWamRDMXlZWFJwYnkwdE4zZzFMV3g3Y0dGa1pHbHVaeTFpYjNSMGIyMDZOekV1TkRJbGZTNWhjM0JsWTNRdGNtRjBhVzh0TFRWNE55MXNlM0JoWkdScGJtY3RZbTkwZEc5dE9qRTBNQ1Y5TG1GemNHVmpkQzF5WVhScGJ5MHRNWGd4TFd4N2NHRmtaR2x1WnkxaWIzUjBiMjA2TVRBd0pYMHVZWE53WldOMExYSmhkR2x2TFMxdlltcGxZM1F0Ykh0d2IzTnBkR2x2YmpwaFluTnZiSFYwWlR0MGIzQTZNRHR5YVdkb2REb3dPMkp2ZEhSdmJUb3dPMnhsWm5RNk1EdDNhV1IwYURveE1EQWxPMmhsYVdkb2REb3hNREFsTzNvdGFXNWtaWGc2TVRBd2ZTNWpiM1psY2kxc2UySmhZMnRuY205MWJtUXRjMmw2WlRwamIzWmxjaUZwYlhCdmNuUmhiblI5TG1OdmJuUmhhVzR0Ykh0aVlXTnJaM0p2ZFc1a0xYTnBlbVU2WTI5dWRHRnBiaUZwYlhCdmNuUmhiblI5TG1KbkxXTmxiblJsY2kxc2UySmhZMnRuY205MWJtUXRjRzl6YVhScGIyNDZOVEFsZlM1aVp5MWpaVzUwWlhJdGJDd3VZbWN0ZEc5d0xXeDdZbUZqYTJkeWIzVnVaQzF5WlhCbFlYUTZibTh0Y21Wd1pXRjBmUzVpWnkxMGIzQXRiSHRpWVdOclozSnZkVzVrTFhCdmMybDBhVzl1T25SdmNIMHVZbWN0Y21sbmFIUXRiSHRpWVdOclozSnZkVzVrTFhCdmMybDBhVzl1T2pFd01DVjlMbUpuTFdKdmRIUnZiUzFzTEM1aVp5MXlhV2RvZEMxc2UySmhZMnRuY205MWJtUXRjbVZ3WldGME9tNXZMWEpsY0dWaGRIMHVZbWN0WW05MGRHOXRMV3g3WW1GamEyZHliM1Z1WkMxd2IzTnBkR2x2YmpwaWIzUjBiMjE5TG1KbkxXeGxablF0Ykh0aVlXTnJaM0p2ZFc1a0xYSmxjR1ZoZERwdWJ5MXlaWEJsWVhRN1ltRmphMmR5YjNWdVpDMXdiM05wZEdsdmJqb3dmUzV2ZFhSc2FXNWxMV3g3YjNWMGJHbHVaVG94Y0hnZ2MyOXNhV1I5TG05MWRHeHBibVV0ZEhKaGJuTndZWEpsYm5RdGJIdHZkWFJzYVc1bE9qRndlQ0J6YjJ4cFpDQjBjbUZ1YzNCaGNtVnVkSDB1YjNWMGJHbHVaUzB3TFd4N2IzVjBiR2x1WlRvd2ZTNWlZUzFzZTJKdmNtUmxjaTF6ZEhsc1pUcHpiMnhwWkR0aWIzSmtaWEl0ZDJsa2RHZzZNWEI0ZlM1aWRDMXNlMkp2Y21SbGNpMTBiM0F0YzNSNWJHVTZjMjlzYVdRN1ltOXlaR1Z5TFhSdmNDMTNhV1IwYURveGNIaDlMbUp5TFd4N1ltOXlaR1Z5TFhKcFoyaDBMWE4wZVd4bE9uTnZiR2xrTzJKdmNtUmxjaTF5YVdkb2RDMTNhV1IwYURveGNIaDlMbUppTFd4N1ltOXlaR1Z5TFdKdmRIUnZiUzF6ZEhsc1pUcHpiMnhwWkR0aWIzSmtaWEl0WW05MGRHOXRMWGRwWkhSb09qRndlSDB1WW13dGJIdGliM0prWlhJdGJHVm1kQzF6ZEhsc1pUcHpiMnhwWkR0aWIzSmtaWEl0YkdWbWRDMTNhV1IwYURveGNIaDlMbUp1TFd4N1ltOXlaR1Z5TFhOMGVXeGxPbTV2Ym1VN1ltOXlaR1Z5TFhkcFpIUm9PakI5TG1KeU1DMXNlMkp2Y21SbGNpMXlZV1JwZFhNNk1IMHVZbkl4TFd4N1ltOXlaR1Z5TFhKaFpHbDFjem91TVRJMWNtVnRmUzVpY2pJdGJIdGliM0prWlhJdGNtRmthWFZ6T2k0eU5YSmxiWDB1WW5JekxXeDdZbTl5WkdWeUxYSmhaR2wxY3pvdU5YSmxiWDB1WW5JMExXeDdZbTl5WkdWeUxYSmhaR2wxY3pveGNtVnRmUzVpY2kweE1EQXRiSHRpYjNKa1pYSXRjbUZrYVhWek9qRXdNQ1Y5TG1KeUxYQnBiR3d0Ykh0aWIzSmtaWEl0Y21Ga2FYVnpPams1T1Rsd2VIMHVZbkl0TFdKdmRIUnZiUzFzZTJKdmNtUmxjaTEwYjNBdGJHVm1kQzF5WVdScGRYTTZNRHRpYjNKa1pYSXRkRzl3TFhKcFoyaDBMWEpoWkdsMWN6b3dmUzVpY2kwdGRHOXdMV3g3WW05eVpHVnlMV0p2ZEhSdmJTMXlhV2RvZEMxeVlXUnBkWE02TUgwdVluSXRMWEpwWjJoMExXd3NMbUp5TFMxMGIzQXRiSHRpYjNKa1pYSXRZbTkwZEc5dExXeGxablF0Y21Ga2FYVnpPakI5TG1KeUxTMXlhV2RvZEMxc2UySnZjbVJsY2kxMGIzQXRiR1ZtZEMxeVlXUnBkWE02TUgwdVluSXRMV3hsWm5RdGJIdGliM0prWlhJdGRHOXdMWEpwWjJoMExYSmhaR2wxY3pvd08ySnZjbVJsY2kxaWIzUjBiMjB0Y21sbmFIUXRjbUZrYVhWek9qQjlMbUl0TFdSdmRIUmxaQzFzZTJKdmNtUmxjaTF6ZEhsc1pUcGtiM1IwWldSOUxtSXRMV1JoYzJobFpDMXNlMkp2Y21SbGNpMXpkSGxzWlRwa1lYTm9aV1I5TG1JdExYTnZiR2xrTFd4N1ltOXlaR1Z5TFhOMGVXeGxPbk52Ykdsa2ZTNWlMUzF1YjI1bExXeDdZbTl5WkdWeUxYTjBlV3hsT201dmJtVjlMbUozTUMxc2UySnZjbVJsY2kxM2FXUjBhRG93ZlM1aWR6RXRiSHRpYjNKa1pYSXRkMmxrZEdnNkxqRXlOWEpsYlgwdVluY3lMV3g3WW05eVpHVnlMWGRwWkhSb09pNHlOWEpsYlgwdVluY3pMV3g3WW05eVpHVnlMWGRwWkhSb09pNDFjbVZ0ZlM1aWR6UXRiSHRpYjNKa1pYSXRkMmxrZEdnNk1YSmxiWDB1WW5jMUxXeDdZbTl5WkdWeUxYZHBaSFJvT2pKeVpXMTlMbUowTFRBdGJIdGliM0prWlhJdGRHOXdMWGRwWkhSb09qQjlMbUp5TFRBdGJIdGliM0prWlhJdGNtbG5hSFF0ZDJsa2RHZzZNSDB1WW1JdE1DMXNlMkp2Y21SbGNpMWliM1IwYjIwdGQybGtkR2c2TUgwdVltd3RNQzFzZTJKdmNtUmxjaTFzWldaMExYZHBaSFJvT2pCOUxuTm9ZV1J2ZHkweExXeDdZbTk0TFhOb1lXUnZkem93SURBZ05IQjRJREp3ZUNCeVoySmhLREFzTUN3d0xDNHlLWDB1YzJoaFpHOTNMVEl0Ykh0aWIzZ3RjMmhoWkc5M09qQWdNQ0E0Y0hnZ01uQjRJSEpuWW1Fb01Dd3dMREFzTGpJcGZTNXphR0ZrYjNjdE15MXNlMkp2ZUMxemFHRmtiM2M2TW5CNElESndlQ0EwY0hnZ01uQjRJSEpuWW1Fb01Dd3dMREFzTGpJcGZTNXphR0ZrYjNjdE5DMXNlMkp2ZUMxemFHRmtiM2M2TW5CNElESndlQ0E0Y0hnZ01DQnlaMkpoS0RBc01Dd3dMQzR5S1gwdWMyaGhaRzkzTFRVdGJIdGliM2d0YzJoaFpHOTNPalJ3ZUNBMGNIZ2dPSEI0SURBZ2NtZGlZU2d3TERBc01Dd3VNaWw5TG5SdmNDMHdMV3g3ZEc5d09qQjlMbXhsWm5RdE1DMXNlMnhsWm5RNk1IMHVjbWxuYUhRdE1DMXNlM0pwWjJoME9qQjlMbUp2ZEhSdmJTMHdMV3g3WW05MGRHOXRPakI5TG5SdmNDMHhMV3g3ZEc5d09qRnlaVzE5TG14bFpuUXRNUzFzZTJ4bFpuUTZNWEpsYlgwdWNtbG5hSFF0TVMxc2UzSnBaMmgwT2pGeVpXMTlMbUp2ZEhSdmJTMHhMV3g3WW05MGRHOXRPakZ5WlcxOUxuUnZjQzB5TFd4N2RHOXdPakp5WlcxOUxteGxablF0TWkxc2UyeGxablE2TW5KbGJYMHVjbWxuYUhRdE1pMXNlM0pwWjJoME9qSnlaVzE5TG1KdmRIUnZiUzB5TFd4N1ltOTBkRzl0T2pKeVpXMTlMblJ2Y0MwdE1TMXNlM1J2Y0RvdE1YSmxiWDB1Y21sbmFIUXRMVEV0Ykh0eWFXZG9kRG90TVhKbGJYMHVZbTkwZEc5dExTMHhMV3g3WW05MGRHOXRPaTB4Y21WdGZTNXNaV1owTFMweExXeDdiR1ZtZERvdE1YSmxiWDB1ZEc5d0xTMHlMV3g3ZEc5d09pMHljbVZ0ZlM1eWFXZG9kQzB0TWkxc2UzSnBaMmgwT2kweWNtVnRmUzVpYjNSMGIyMHRMVEl0Ykh0aWIzUjBiMjA2TFRKeVpXMTlMbXhsWm5RdExUSXRiSHRzWldaME9pMHljbVZ0ZlM1aFluTnZiSFYwWlMwdFptbHNiQzFzZTNSdmNEb3dPM0pwWjJoME9qQTdZbTkwZEc5dE9qQTdiR1ZtZERvd2ZTNWpiQzFzZTJOc1pXRnlPbXhsWm5SOUxtTnlMV3g3WTJ4bFlYSTZjbWxuYUhSOUxtTmlMV3g3WTJ4bFlYSTZZbTkwYUgwdVkyNHRiSHRqYkdWaGNqcHViMjVsZlM1a2JpMXNlMlJwYzNCc1lYazZibTl1WlgwdVpHa3RiSHRrYVhOd2JHRjVPbWx1YkdsdVpYMHVaR0l0Ykh0a2FYTndiR0Y1T21Kc2IyTnJmUzVrYVdJdGJIdGthWE53YkdGNU9tbHViR2x1WlMxaWJHOWphMzB1WkdsMExXeDdaR2x6Y0d4aGVUcHBibXhwYm1VdGRHRmliR1Y5TG1SMExXeDdaR2x6Y0d4aGVUcDBZV0pzWlgwdVpIUmpMV3g3WkdsemNHeGhlVHAwWVdKc1pTMWpaV3hzZlM1a2RDMXliM2N0Ykh0a2FYTndiR0Y1T25SaFlteGxMWEp2ZDMwdVpIUXRjbTkzTFdkeWIzVndMV3g3WkdsemNHeGhlVHAwWVdKc1pTMXliM2N0WjNKdmRYQjlMbVIwTFdOdmJIVnRiaTFzZTJScGMzQnNZWGs2ZEdGaWJHVXRZMjlzZFcxdWZTNWtkQzFqYjJ4MWJXNHRaM0p2ZFhBdGJIdGthWE53YkdGNU9uUmhZbXhsTFdOdmJIVnRiaTFuY205MWNIMHVaSFF0TFdacGVHVmtMV3g3ZEdGaWJHVXRiR0Y1YjNWME9tWnBlR1ZrTzNkcFpIUm9PakV3TUNWOUxtWnNaWGd0Ykh0a2FYTndiR0Y1T21ac1pYaDlMbWx1YkdsdVpTMW1iR1Y0TFd4N1pHbHpjR3hoZVRwcGJteHBibVV0Wm14bGVIMHVabXhsZUMxaGRYUnZMV3g3Wm14bGVEb3hJREVnWVhWMGJ6dHRhVzR0ZDJsa2RHZzZNRHR0YVc0dGFHVnBaMmgwT2pCOUxtWnNaWGd0Ym05dVpTMXNlMlpzWlhnNmJtOXVaWDB1Wm14bGVDMWpiMngxYlc0dGJIdG1iR1Y0TFdScGNtVmpkR2x2YmpwamIyeDFiVzU5TG1ac1pYZ3RjbTkzTFd4N1pteGxlQzFrYVhKbFkzUnBiMjQ2Y205M2ZTNW1iR1Y0TFhkeVlYQXRiSHRtYkdWNExYZHlZWEE2ZDNKaGNIMHVabXhsZUMxdWIzZHlZWEF0Ykh0bWJHVjRMWGR5WVhBNmJtOTNjbUZ3ZlM1bWJHVjRMWGR5WVhBdGNtVjJaWEp6WlMxc2UyWnNaWGd0ZDNKaGNEcDNjbUZ3TFhKbGRtVnljMlY5TG1ac1pYZ3RZMjlzZFcxdUxYSmxkbVZ5YzJVdGJIdG1iR1Y0TFdScGNtVmpkR2x2YmpwamIyeDFiVzR0Y21WMlpYSnpaWDB1Wm14bGVDMXliM2N0Y21WMlpYSnpaUzFzZTJac1pYZ3RaR2x5WldOMGFXOXVPbkp2ZHkxeVpYWmxjbk5sZlM1cGRHVnRjeTF6ZEdGeWRDMXNlMkZzYVdkdUxXbDBaVzF6T21ac1pYZ3RjM1JoY25SOUxtbDBaVzF6TFdWdVpDMXNlMkZzYVdkdUxXbDBaVzF6T21ac1pYZ3RaVzVrZlM1cGRHVnRjeTFqWlc1MFpYSXRiSHRoYkdsbmJpMXBkR1Z0Y3pwalpXNTBaWEo5TG1sMFpXMXpMV0poYzJWc2FXNWxMV3g3WVd4cFoyNHRhWFJsYlhNNlltRnpaV3hwYm1WOUxtbDBaVzF6TFhOMGNtVjBZMmd0Ykh0aGJHbG5iaTFwZEdWdGN6cHpkSEpsZEdOb2ZTNXpaV3htTFhOMFlYSjBMV3g3WVd4cFoyNHRjMlZzWmpwbWJHVjRMWE4wWVhKMGZTNXpaV3htTFdWdVpDMXNlMkZzYVdkdUxYTmxiR1k2Wm14bGVDMWxibVI5TG5ObGJHWXRZMlZ1ZEdWeUxXeDdZV3hwWjI0dGMyVnNaanBqWlc1MFpYSjlMbk5sYkdZdFltRnpaV3hwYm1VdGJIdGhiR2xuYmkxelpXeG1PbUpoYzJWc2FXNWxmUzV6Wld4bUxYTjBjbVYwWTJndGJIdGhiR2xuYmkxelpXeG1Pbk4wY21WMFkyaDlMbXAxYzNScFpua3RjM1JoY25RdGJIdHFkWE4wYVdaNUxXTnZiblJsYm5RNlpteGxlQzF6ZEdGeWRIMHVhblZ6ZEdsbWVTMWxibVF0Ykh0cWRYTjBhV1o1TFdOdmJuUmxiblE2Wm14bGVDMWxibVI5TG1wMWMzUnBabmt0WTJWdWRHVnlMV3g3YW5WemRHbG1lUzFqYjI1MFpXNTBPbU5sYm5SbGNuMHVhblZ6ZEdsbWVTMWlaWFIzWldWdUxXeDdhblZ6ZEdsbWVTMWpiMjUwWlc1ME9uTndZV05sTFdKbGRIZGxaVzU5TG1wMWMzUnBabmt0WVhKdmRXNWtMV3g3YW5WemRHbG1lUzFqYjI1MFpXNTBPbk53WVdObExXRnliM1Z1WkgwdVkyOXVkR1Z1ZEMxemRHRnlkQzFzZTJGc2FXZHVMV052Ym5SbGJuUTZabXhsZUMxemRHRnlkSDB1WTI5dWRHVnVkQzFsYm1RdGJIdGhiR2xuYmkxamIyNTBaVzUwT21ac1pYZ3RaVzVrZlM1amIyNTBaVzUwTFdObGJuUmxjaTFzZTJGc2FXZHVMV052Ym5SbGJuUTZZMlZ1ZEdWeWZTNWpiMjUwWlc1MExXSmxkSGRsWlc0dGJIdGhiR2xuYmkxamIyNTBaVzUwT25Od1lXTmxMV0psZEhkbFpXNTlMbU52Ym5SbGJuUXRZWEp2ZFc1a0xXeDdZV3hwWjI0dFkyOXVkR1Z1ZERwemNHRmpaUzFoY205MWJtUjlMbU52Ym5SbGJuUXRjM1J5WlhSamFDMXNlMkZzYVdkdUxXTnZiblJsYm5RNmMzUnlaWFJqYUgwdWIzSmtaWEl0TUMxc2UyOXlaR1Z5T2pCOUxtOXlaR1Z5TFRFdGJIdHZjbVJsY2pveGZTNXZjbVJsY2kweUxXeDdiM0prWlhJNk1uMHViM0prWlhJdE15MXNlMjl5WkdWeU9qTjlMbTl5WkdWeUxUUXRiSHR2Y21SbGNqbzBmUzV2Y21SbGNpMDFMV3g3YjNKa1pYSTZOWDB1YjNKa1pYSXROaTFzZTI5eVpHVnlPalo5TG05eVpHVnlMVGN0Ykh0dmNtUmxjam8zZlM1dmNtUmxjaTA0TFd4N2IzSmtaWEk2T0gwdWIzSmtaWEl0YkdGemRDMXNlMjl5WkdWeU9qazVPVGs1ZlM1bWJHVjRMV2R5YjNjdE1DMXNlMlpzWlhndFozSnZkem93ZlM1bWJHVjRMV2R5YjNjdE1TMXNlMlpzWlhndFozSnZkem94ZlM1bWJHVjRMWE5vY21sdWF5MHdMV3g3Wm14bGVDMXphSEpwYm1zNk1IMHVabXhsZUMxemFISnBibXN0TVMxc2UyWnNaWGd0YzJoeWFXNXJPakY5TG1ac0xXeDdabXh2WVhRNmJHVm1kSDB1Wm13dGJDd3Vabkl0Ykh0ZlpHbHpjR3hoZVRwcGJteHBibVY5TG1aeUxXeDdabXh2WVhRNmNtbG5hSFI5TG1adUxXeDdabXh2WVhRNmJtOXVaWDB1YVMxc2UyWnZiblF0YzNSNWJHVTZhWFJoYkdsamZTNW1jeTF1YjNKdFlXd3RiSHRtYjI1MExYTjBlV3hsT201dmNtMWhiSDB1Ym05eWJXRnNMV3g3Wm05dWRDMTNaV2xuYUhRNk5EQXdmUzVpTFd4N1ptOXVkQzEzWldsbmFIUTZOekF3ZlM1bWR6RXRiSHRtYjI1MExYZGxhV2RvZERveE1EQjlMbVozTWkxc2UyWnZiblF0ZDJWcFoyaDBPakl3TUgwdVpuY3pMV3g3Wm05dWRDMTNaV2xuYUhRNk16QXdmUzVtZHpRdGJIdG1iMjUwTFhkbGFXZG9kRG8wTURCOUxtWjNOUzFzZTJadmJuUXRkMlZwWjJoME9qVXdNSDB1Wm5jMkxXeDdabTl1ZEMxM1pXbG5hSFE2TmpBd2ZTNW1kemN0Ykh0bWIyNTBMWGRsYVdkb2REbzNNREI5TG1aM09DMXNlMlp2Ym5RdGQyVnBaMmgwT2pnd01IMHVabmM1TFd4N1ptOXVkQzEzWldsbmFIUTZPVEF3ZlM1b01TMXNlMmhsYVdkb2REb3hjbVZ0ZlM1b01pMXNlMmhsYVdkb2REb3ljbVZ0ZlM1b015MXNlMmhsYVdkb2REbzBjbVZ0ZlM1b05DMXNlMmhsYVdkb2REbzRjbVZ0ZlM1b05TMXNlMmhsYVdkb2REb3hObkpsYlgwdWFDMHlOUzFzZTJobGFXZG9kRG95TlNWOUxtZ3ROVEF0Ykh0b1pXbG5hSFE2TlRBbGZTNW9MVGMxTFd4N2FHVnBaMmgwT2pjMUpYMHVhQzB4TURBdGJIdG9aV2xuYUhRNk1UQXdKWDB1YldsdUxXZ3RNVEF3TFd4N2JXbHVMV2hsYVdkb2REb3hNREFsZlM1MmFDMHlOUzFzZTJobGFXZG9kRG95Tlhab2ZTNTJhQzAxTUMxc2UyaGxhV2RvZERvMU1IWm9mUzUyYUMwM05TMXNlMmhsYVdkb2REbzNOWFpvZlM1MmFDMHhNREF0Ykh0b1pXbG5hSFE2TVRBd2RtaDlMbTFwYmkxMmFDMHhNREF0Ykh0dGFXNHRhR1ZwWjJoME9qRXdNSFpvZlM1b0xXRjFkRzh0Ykh0b1pXbG5hSFE2WVhWMGIzMHVhQzFwYm1obGNtbDBMV3g3YUdWcFoyaDBPbWx1YUdWeWFYUjlMblJ5WVdOclpXUXRiSHRzWlhSMFpYSXRjM0JoWTJsdVp6b3VNV1Z0ZlM1MGNtRmphMlZrTFhScFoyaDBMV3g3YkdWMGRHVnlMWE53WVdOcGJtYzZMUzR3TldWdGZTNTBjbUZqYTJWa0xXMWxaMkV0Ykh0c1pYUjBaWEl0YzNCaFkybHVaem91TWpWbGJYMHViR2d0YzI5c2FXUXRiSHRzYVc1bExXaGxhV2RvZERveGZTNXNhQzEwYVhSc1pTMXNlMnhwYm1VdGFHVnBaMmgwT2pFdU1qVjlMbXhvTFdOdmNIa3RiSHRzYVc1bExXaGxhV2RvZERveExqVjlMbTEzTFRFd01DMXNlMjFoZUMxM2FXUjBhRG94TURBbGZTNXRkekV0Ykh0dFlYZ3RkMmxrZEdnNk1YSmxiWDB1YlhjeUxXeDdiV0Y0TFhkcFpIUm9Pakp5WlcxOUxtMTNNeTFzZTIxaGVDMTNhV1IwYURvMGNtVnRmUzV0ZHpRdGJIdHRZWGd0ZDJsa2RHZzZPSEpsYlgwdWJYYzFMV3g3YldGNExYZHBaSFJvT2pFMmNtVnRmUzV0ZHpZdGJIdHRZWGd0ZDJsa2RHZzZNekp5WlcxOUxtMTNOeTFzZTIxaGVDMTNhV1IwYURvME9ISmxiWDB1YlhjNExXeDdiV0Y0TFhkcFpIUm9PalkwY21WdGZTNXRkemt0Ykh0dFlYZ3RkMmxrZEdnNk9UWnlaVzE5TG0xM0xXNXZibVV0Ykh0dFlYZ3RkMmxrZEdnNmJtOXVaWDB1ZHpFdGJIdDNhV1IwYURveGNtVnRmUzUzTWkxc2UzZHBaSFJvT2pKeVpXMTlMbmN6TFd4N2QybGtkR2c2TkhKbGJYMHVkelF0Ykh0M2FXUjBhRG80Y21WdGZTNTNOUzFzZTNkcFpIUm9PakUyY21WdGZTNTNMVEV3TFd4N2QybGtkR2c2TVRBbGZTNTNMVEl3TFd4N2QybGtkR2c2TWpBbGZTNTNMVEkxTFd4N2QybGtkR2c2TWpVbGZTNTNMVE13TFd4N2QybGtkR2c2TXpBbGZTNTNMVE16TFd4N2QybGtkR2c2TXpNbGZTNTNMVE0wTFd4N2QybGtkR2c2TXpRbGZTNTNMVFF3TFd4N2QybGtkR2c2TkRBbGZTNTNMVFV3TFd4N2QybGtkR2c2TlRBbGZTNTNMVFl3TFd4N2QybGtkR2c2TmpBbGZTNTNMVGN3TFd4N2QybGtkR2c2TnpBbGZTNTNMVGMxTFd4N2QybGtkR2c2TnpVbGZTNTNMVGd3TFd4N2QybGtkR2c2T0RBbGZTNTNMVGt3TFd4N2QybGtkR2c2T1RBbGZTNTNMVEV3TUMxc2UzZHBaSFJvT2pFd01DVjlMbmN0ZEdocGNtUXRiSHQzYVdSMGFEb3pNeTR6TXpNek15VjlMbmN0ZEhkdkxYUm9hWEprY3kxc2UzZHBaSFJvT2pZMkxqWTJOalkzSlgwdWR5MWhkWFJ2TFd4N2QybGtkR2c2WVhWMGIzMHViM1psY21ac2IzY3RkbWx6YVdKc1pTMXNlMjkyWlhKbWJHOTNPblpwYzJsaWJHVjlMbTkyWlhKbWJHOTNMV2hwWkdSbGJpMXNlMjkyWlhKbWJHOTNPbWhwWkdSbGJuMHViM1psY21ac2IzY3RjMk55YjJ4c0xXeDdiM1psY21ac2IzYzZjMk55YjJ4c2ZTNXZkbVZ5Wm14dmR5MWhkWFJ2TFd4N2IzWmxjbVpzYjNjNllYVjBiMzB1YjNabGNtWnNiM2N0ZUMxMmFYTnBZbXhsTFd4N2IzWmxjbVpzYjNjdGVEcDJhWE5wWW14bGZTNXZkbVZ5Wm14dmR5MTRMV2hwWkdSbGJpMXNlMjkyWlhKbWJHOTNMWGc2YUdsa1pHVnVmUzV2ZG1WeVpteHZkeTE0TFhOamNtOXNiQzFzZTI5MlpYSm1iRzkzTFhnNmMyTnliMnhzZlM1dmRtVnlabXh2ZHkxNExXRjFkRzh0Ykh0dmRtVnlabXh2ZHkxNE9tRjFkRzk5TG05MlpYSm1iRzkzTFhrdGRtbHphV0pzWlMxc2UyOTJaWEptYkc5M0xYazZkbWx6YVdKc1pYMHViM1psY21ac2IzY3RlUzFvYVdSa1pXNHRiSHR2ZG1WeVpteHZkeTE1T21ocFpHUmxibjB1YjNabGNtWnNiM2N0ZVMxelkzSnZiR3d0Ykh0dmRtVnlabXh2ZHkxNU9uTmpjbTlzYkgwdWIzWmxjbVpzYjNjdGVTMWhkWFJ2TFd4N2IzWmxjbVpzYjNjdGVUcGhkWFJ2ZlM1emRHRjBhV010Ykh0d2IzTnBkR2x2YmpwemRHRjBhV045TG5KbGJHRjBhWFpsTFd4N2NHOXphWFJwYjI0NmNtVnNZWFJwZG1WOUxtRmljMjlzZFhSbExXeDdjRzl6YVhScGIyNDZZV0p6YjJ4MWRHVjlMbVpwZUdWa0xXeDdjRzl6YVhScGIyNDZabWw0WldSOUxuSnZkR0YwWlMwME5TMXNleTEzWldKcmFYUXRkSEpoYm5ObWIzSnRPbkp2ZEdGMFpTZzBOV1JsWnlrN2RISmhibk5tYjNKdE9uSnZkR0YwWlNnME5XUmxaeWw5TG5KdmRHRjBaUzA1TUMxc2V5MTNaV0pyYVhRdGRISmhibk5tYjNKdE9uSnZkR0YwWlNnNU1HUmxaeWs3ZEhKaGJuTm1iM0p0T25KdmRHRjBaU2c1TUdSbFp5bDlMbkp2ZEdGMFpTMHhNelV0YkhzdGQyVmlhMmwwTFhSeVlXNXpabTl5YlRweWIzUmhkR1VvTVRNMVpHVm5LVHQwY21GdWMyWnZjbTA2Y205MFlYUmxLREV6TldSbFp5bDlMbkp2ZEdGMFpTMHhPREF0YkhzdGQyVmlhMmwwTFhSeVlXNXpabTl5YlRweWIzUmhkR1VvTVRnd1pHVm5LVHQwY21GdWMyWnZjbTA2Y205MFlYUmxLREU0TUdSbFp5bDlMbkp2ZEdGMFpTMHlNalV0YkhzdGQyVmlhMmwwTFhSeVlXNXpabTl5YlRweWIzUmhkR1VvTWpJMVpHVm5LVHQwY21GdWMyWnZjbTA2Y205MFlYUmxLREl5TldSbFp5bDlMbkp2ZEdGMFpTMHlOekF0YkhzdGQyVmlhMmwwTFhSeVlXNXpabTl5YlRweWIzUmhkR1VvTWpjd1pHVm5LVHQwY21GdWMyWnZjbTA2Y205MFlYUmxLREkzTUdSbFp5bDlMbkp2ZEdGMFpTMHpNVFV0YkhzdGQyVmlhMmwwTFhSeVlXNXpabTl5YlRweWIzUmhkR1VvTXpFMVpHVm5LVHQwY21GdWMyWnZjbTA2Y205MFlYUmxLRE14TldSbFp5bDlMbkJoTUMxc2UzQmhaR1JwYm1jNk1IMHVjR0V4TFd4N2NHRmtaR2x1WnpvdU1qVnlaVzE5TG5CaE1pMXNlM0JoWkdScGJtYzZMalZ5WlcxOUxuQmhNeTFzZTNCaFpHUnBibWM2TVhKbGJYMHVjR0UwTFd4N2NHRmtaR2x1WnpveWNtVnRmUzV3WVRVdGJIdHdZV1JrYVc1bk9qUnlaVzE5TG5CaE5pMXNlM0JoWkdScGJtYzZPSEpsYlgwdWNHRTNMV3g3Y0dGa1pHbHVaem94Tm5KbGJYMHVjR3d3TFd4N2NHRmtaR2x1Wnkxc1pXWjBPakI5TG5Cc01TMXNlM0JoWkdScGJtY3RiR1ZtZERvdU1qVnlaVzE5TG5Cc01pMXNlM0JoWkdScGJtY3RiR1ZtZERvdU5YSmxiWDB1Y0d3ekxXeDdjR0ZrWkdsdVp5MXNaV1owT2pGeVpXMTlMbkJzTkMxc2UzQmhaR1JwYm1jdGJHVm1kRG95Y21WdGZTNXdiRFV0Ykh0d1lXUmthVzVuTFd4bFpuUTZOSEpsYlgwdWNHdzJMV3g3Y0dGa1pHbHVaeTFzWldaME9qaHlaVzE5TG5Cc055MXNlM0JoWkdScGJtY3RiR1ZtZERveE5uSmxiWDB1Y0hJd0xXeDdjR0ZrWkdsdVp5MXlhV2RvZERvd2ZTNXdjakV0Ykh0d1lXUmthVzVuTFhKcFoyaDBPaTR5TlhKbGJYMHVjSEl5TFd4N2NHRmtaR2x1WnkxeWFXZG9kRG91TlhKbGJYMHVjSEl6TFd4N2NHRmtaR2x1WnkxeWFXZG9kRG94Y21WdGZTNXdjalF0Ykh0d1lXUmthVzVuTFhKcFoyaDBPakp5WlcxOUxuQnlOUzFzZTNCaFpHUnBibWN0Y21sbmFIUTZOSEpsYlgwdWNISTJMV3g3Y0dGa1pHbHVaeTF5YVdkb2REbzRjbVZ0ZlM1d2NqY3RiSHR3WVdSa2FXNW5MWEpwWjJoME9qRTJjbVZ0ZlM1d1lqQXRiSHR3WVdSa2FXNW5MV0p2ZEhSdmJUb3dmUzV3WWpFdGJIdHdZV1JrYVc1bkxXSnZkSFJ2YlRvdU1qVnlaVzE5TG5CaU1pMXNlM0JoWkdScGJtY3RZbTkwZEc5dE9pNDFjbVZ0ZlM1d1lqTXRiSHR3WVdSa2FXNW5MV0p2ZEhSdmJUb3hjbVZ0ZlM1d1lqUXRiSHR3WVdSa2FXNW5MV0p2ZEhSdmJUb3ljbVZ0ZlM1d1lqVXRiSHR3WVdSa2FXNW5MV0p2ZEhSdmJUbzBjbVZ0ZlM1d1lqWXRiSHR3WVdSa2FXNW5MV0p2ZEhSdmJUbzRjbVZ0ZlM1d1lqY3RiSHR3WVdSa2FXNW5MV0p2ZEhSdmJUb3hObkpsYlgwdWNIUXdMV3g3Y0dGa1pHbHVaeTEwYjNBNk1IMHVjSFF4TFd4N2NHRmtaR2x1WnkxMGIzQTZMakkxY21WdGZTNXdkREl0Ykh0d1lXUmthVzVuTFhSdmNEb3VOWEpsYlgwdWNIUXpMV3g3Y0dGa1pHbHVaeTEwYjNBNk1YSmxiWDB1Y0hRMExXeDdjR0ZrWkdsdVp5MTBiM0E2TW5KbGJYMHVjSFExTFd4N2NHRmtaR2x1WnkxMGIzQTZOSEpsYlgwdWNIUTJMV3g3Y0dGa1pHbHVaeTEwYjNBNk9ISmxiWDB1Y0hRM0xXeDdjR0ZrWkdsdVp5MTBiM0E2TVRaeVpXMTlMbkIyTUMxc2UzQmhaR1JwYm1jdGRHOXdPakE3Y0dGa1pHbHVaeTFpYjNSMGIyMDZNSDB1Y0hZeExXeDdjR0ZrWkdsdVp5MTBiM0E2TGpJMWNtVnRPM0JoWkdScGJtY3RZbTkwZEc5dE9pNHlOWEpsYlgwdWNIWXlMV3g3Y0dGa1pHbHVaeTEwYjNBNkxqVnlaVzA3Y0dGa1pHbHVaeTFpYjNSMGIyMDZMalZ5WlcxOUxuQjJNeTFzZTNCaFpHUnBibWN0ZEc5d09qRnlaVzA3Y0dGa1pHbHVaeTFpYjNSMGIyMDZNWEpsYlgwdWNIWTBMV3g3Y0dGa1pHbHVaeTEwYjNBNk1uSmxiVHR3WVdSa2FXNW5MV0p2ZEhSdmJUb3ljbVZ0ZlM1d2RqVXRiSHR3WVdSa2FXNW5MWFJ2Y0RvMGNtVnRPM0JoWkdScGJtY3RZbTkwZEc5dE9qUnlaVzE5TG5CMk5pMXNlM0JoWkdScGJtY3RkRzl3T2poeVpXMDdjR0ZrWkdsdVp5MWliM1IwYjIwNk9ISmxiWDB1Y0hZM0xXeDdjR0ZrWkdsdVp5MTBiM0E2TVRaeVpXMDdjR0ZrWkdsdVp5MWliM1IwYjIwNk1UWnlaVzE5TG5Cb01DMXNlM0JoWkdScGJtY3RiR1ZtZERvd08zQmhaR1JwYm1jdGNtbG5hSFE2TUgwdWNHZ3hMV3g3Y0dGa1pHbHVaeTFzWldaME9pNHlOWEpsYlR0d1lXUmthVzVuTFhKcFoyaDBPaTR5TlhKbGJYMHVjR2d5TFd4N2NHRmtaR2x1Wnkxc1pXWjBPaTQxY21WdE8zQmhaR1JwYm1jdGNtbG5hSFE2TGpWeVpXMTlMbkJvTXkxc2UzQmhaR1JwYm1jdGJHVm1kRG94Y21WdE8zQmhaR1JwYm1jdGNtbG5hSFE2TVhKbGJYMHVjR2cwTFd4N2NHRmtaR2x1Wnkxc1pXWjBPakp5WlcwN2NHRmtaR2x1WnkxeWFXZG9kRG95Y21WdGZTNXdhRFV0Ykh0d1lXUmthVzVuTFd4bFpuUTZOSEpsYlR0d1lXUmthVzVuTFhKcFoyaDBPalJ5WlcxOUxuQm9OaTFzZTNCaFpHUnBibWN0YkdWbWREbzRjbVZ0TzNCaFpHUnBibWN0Y21sbmFIUTZPSEpsYlgwdWNHZzNMV3g3Y0dGa1pHbHVaeTFzWldaME9qRTJjbVZ0TzNCaFpHUnBibWN0Y21sbmFIUTZNVFp5WlcxOUxtMWhNQzFzZTIxaGNtZHBiam93ZlM1dFlURXRiSHR0WVhKbmFXNDZMakkxY21WdGZTNXRZVEl0Ykh0dFlYSm5hVzQ2TGpWeVpXMTlMbTFoTXkxc2UyMWhjbWRwYmpveGNtVnRmUzV0WVRRdGJIdHRZWEpuYVc0Nk1uSmxiWDB1YldFMUxXeDdiV0Z5WjJsdU9qUnlaVzE5TG0xaE5pMXNlMjFoY21kcGJqbzRjbVZ0ZlM1dFlUY3RiSHR0WVhKbmFXNDZNVFp5WlcxOUxtMXNNQzFzZTIxaGNtZHBiaTFzWldaME9qQjlMbTFzTVMxc2UyMWhjbWRwYmkxc1pXWjBPaTR5TlhKbGJYMHViV3d5TFd4N2JXRnlaMmx1TFd4bFpuUTZMalZ5WlcxOUxtMXNNeTFzZTIxaGNtZHBiaTFzWldaME9qRnlaVzE5TG0xc05DMXNlMjFoY21kcGJpMXNaV1owT2pKeVpXMTlMbTFzTlMxc2UyMWhjbWRwYmkxc1pXWjBPalJ5WlcxOUxtMXNOaTFzZTIxaGNtZHBiaTFzWldaME9qaHlaVzE5TG0xc055MXNlMjFoY21kcGJpMXNaV1owT2pFMmNtVnRmUzV0Y2pBdGJIdHRZWEpuYVc0dGNtbG5hSFE2TUgwdWJYSXhMV3g3YldGeVoybHVMWEpwWjJoME9pNHlOWEpsYlgwdWJYSXlMV3g3YldGeVoybHVMWEpwWjJoME9pNDFjbVZ0ZlM1dGNqTXRiSHR0WVhKbmFXNHRjbWxuYUhRNk1YSmxiWDB1YlhJMExXeDdiV0Z5WjJsdUxYSnBaMmgwT2pKeVpXMTlMbTF5TlMxc2UyMWhjbWRwYmkxeWFXZG9kRG8wY21WdGZTNXRjall0Ykh0dFlYSm5hVzR0Y21sbmFIUTZPSEpsYlgwdWJYSTNMV3g3YldGeVoybHVMWEpwWjJoME9qRTJjbVZ0ZlM1dFlqQXRiSHR0WVhKbmFXNHRZbTkwZEc5dE9qQjlMbTFpTVMxc2UyMWhjbWRwYmkxaWIzUjBiMjA2TGpJMWNtVnRmUzV0WWpJdGJIdHRZWEpuYVc0dFltOTBkRzl0T2k0MWNtVnRmUzV0WWpNdGJIdHRZWEpuYVc0dFltOTBkRzl0T2pGeVpXMTlMbTFpTkMxc2UyMWhjbWRwYmkxaWIzUjBiMjA2TW5KbGJYMHViV0kxTFd4N2JXRnlaMmx1TFdKdmRIUnZiVG8wY21WdGZTNXRZall0Ykh0dFlYSm5hVzR0WW05MGRHOXRPamh5WlcxOUxtMWlOeTFzZTIxaGNtZHBiaTFpYjNSMGIyMDZNVFp5WlcxOUxtMTBNQzFzZTIxaGNtZHBiaTEwYjNBNk1IMHViWFF4TFd4N2JXRnlaMmx1TFhSdmNEb3VNalZ5WlcxOUxtMTBNaTFzZTIxaGNtZHBiaTEwYjNBNkxqVnlaVzE5TG0xME15MXNlMjFoY21kcGJpMTBiM0E2TVhKbGJYMHViWFEwTFd4N2JXRnlaMmx1TFhSdmNEb3ljbVZ0ZlM1dGREVXRiSHR0WVhKbmFXNHRkRzl3T2pSeVpXMTlMbTEwTmkxc2UyMWhjbWRwYmkxMGIzQTZPSEpsYlgwdWJYUTNMV3g3YldGeVoybHVMWFJ2Y0RveE5uSmxiWDB1YlhZd0xXeDdiV0Z5WjJsdUxYUnZjRG93TzIxaGNtZHBiaTFpYjNSMGIyMDZNSDB1YlhZeExXeDdiV0Z5WjJsdUxYUnZjRG91TWpWeVpXMDdiV0Z5WjJsdUxXSnZkSFJ2YlRvdU1qVnlaVzE5TG0xMk1pMXNlMjFoY21kcGJpMTBiM0E2TGpWeVpXMDdiV0Z5WjJsdUxXSnZkSFJ2YlRvdU5YSmxiWDB1YlhZekxXeDdiV0Z5WjJsdUxYUnZjRG94Y21WdE8yMWhjbWRwYmkxaWIzUjBiMjA2TVhKbGJYMHViWFkwTFd4N2JXRnlaMmx1TFhSdmNEb3ljbVZ0TzIxaGNtZHBiaTFpYjNSMGIyMDZNbkpsYlgwdWJYWTFMV3g3YldGeVoybHVMWFJ2Y0RvMGNtVnRPMjFoY21kcGJpMWliM1IwYjIwNk5ISmxiWDB1YlhZMkxXeDdiV0Z5WjJsdUxYUnZjRG80Y21WdE8yMWhjbWRwYmkxaWIzUjBiMjA2T0hKbGJYMHViWFkzTFd4N2JXRnlaMmx1TFhSdmNEb3hObkpsYlR0dFlYSm5hVzR0WW05MGRHOXRPakUyY21WdGZTNXRhREF0Ykh0dFlYSm5hVzR0YkdWbWREb3dPMjFoY21kcGJpMXlhV2RvZERvd2ZTNXRhREV0Ykh0dFlYSm5hVzR0YkdWbWREb3VNalZ5WlcwN2JXRnlaMmx1TFhKcFoyaDBPaTR5TlhKbGJYMHViV2d5TFd4N2JXRnlaMmx1TFd4bFpuUTZMalZ5WlcwN2JXRnlaMmx1TFhKcFoyaDBPaTQxY21WdGZTNXRhRE10Ykh0dFlYSm5hVzR0YkdWbWREb3hjbVZ0TzIxaGNtZHBiaTF5YVdkb2REb3hjbVZ0ZlM1dGFEUXRiSHR0WVhKbmFXNHRiR1ZtZERveWNtVnRPMjFoY21kcGJpMXlhV2RvZERveWNtVnRmUzV0YURVdGJIdHRZWEpuYVc0dGJHVm1kRG8wY21WdE8yMWhjbWRwYmkxeWFXZG9kRG8wY21WdGZTNXRhRFl0Ykh0dFlYSm5hVzR0YkdWbWREbzRjbVZ0TzIxaGNtZHBiaTF5YVdkb2REbzRjbVZ0ZlM1dGFEY3RiSHR0WVhKbmFXNHRiR1ZtZERveE5uSmxiVHR0WVhKbmFXNHRjbWxuYUhRNk1UWnlaVzE5TG01aE1TMXNlMjFoY21kcGJqb3RMakkxY21WdGZTNXVZVEl0Ykh0dFlYSm5hVzQ2TFM0MWNtVnRmUzV1WVRNdGJIdHRZWEpuYVc0NkxURnlaVzE5TG01aE5DMXNlMjFoY21kcGJqb3RNbkpsYlgwdWJtRTFMV3g3YldGeVoybHVPaTAwY21WdGZTNXVZVFl0Ykh0dFlYSm5hVzQ2TFRoeVpXMTlMbTVoTnkxc2UyMWhjbWRwYmpvdE1UWnlaVzE5TG01c01TMXNlMjFoY21kcGJpMXNaV1owT2kwdU1qVnlaVzE5TG01c01pMXNlMjFoY21kcGJpMXNaV1owT2kwdU5YSmxiWDB1Ym13ekxXeDdiV0Z5WjJsdUxXeGxablE2TFRGeVpXMTlMbTVzTkMxc2UyMWhjbWRwYmkxc1pXWjBPaTB5Y21WdGZTNXViRFV0Ykh0dFlYSm5hVzR0YkdWbWREb3ROSEpsYlgwdWJtdzJMV3g3YldGeVoybHVMV3hsWm5RNkxUaHlaVzE5TG01c055MXNlMjFoY21kcGJpMXNaV1owT2kweE5uSmxiWDB1Ym5JeExXeDdiV0Z5WjJsdUxYSnBaMmgwT2kwdU1qVnlaVzE5TG01eU1pMXNlMjFoY21kcGJpMXlhV2RvZERvdExqVnlaVzE5TG01eU15MXNlMjFoY21kcGJpMXlhV2RvZERvdE1YSmxiWDB1Ym5JMExXeDdiV0Z5WjJsdUxYSnBaMmgwT2kweWNtVnRmUzV1Y2pVdGJIdHRZWEpuYVc0dGNtbG5hSFE2TFRSeVpXMTlMbTV5Tmkxc2UyMWhjbWRwYmkxeWFXZG9kRG90T0hKbGJYMHVibkkzTFd4N2JXRnlaMmx1TFhKcFoyaDBPaTB4Tm5KbGJYMHVibUl4TFd4N2JXRnlaMmx1TFdKdmRIUnZiVG90TGpJMWNtVnRmUzV1WWpJdGJIdHRZWEpuYVc0dFltOTBkRzl0T2kwdU5YSmxiWDB1Ym1JekxXeDdiV0Z5WjJsdUxXSnZkSFJ2YlRvdE1YSmxiWDB1Ym1JMExXeDdiV0Z5WjJsdUxXSnZkSFJ2YlRvdE1uSmxiWDB1Ym1JMUxXeDdiV0Z5WjJsdUxXSnZkSFJ2YlRvdE5ISmxiWDB1Ym1JMkxXeDdiV0Z5WjJsdUxXSnZkSFJ2YlRvdE9ISmxiWDB1Ym1JM0xXeDdiV0Z5WjJsdUxXSnZkSFJ2YlRvdE1UWnlaVzE5TG01ME1TMXNlMjFoY21kcGJpMTBiM0E2TFM0eU5YSmxiWDB1Ym5ReUxXeDdiV0Z5WjJsdUxYUnZjRG90TGpWeVpXMTlMbTUwTXkxc2UyMWhjbWRwYmkxMGIzQTZMVEZ5WlcxOUxtNTBOQzFzZTIxaGNtZHBiaTEwYjNBNkxUSnlaVzE5TG01ME5TMXNlMjFoY21kcGJpMTBiM0E2TFRSeVpXMTlMbTUwTmkxc2UyMWhjbWRwYmkxMGIzQTZMVGh5WlcxOUxtNTBOeTFzZTIxaGNtZHBiaTEwYjNBNkxURTJjbVZ0ZlM1emRISnBhMlV0Ykh0MFpYaDBMV1JsWTI5eVlYUnBiMjQ2YkdsdVpTMTBhSEp2ZFdkb2ZTNTFibVJsY214cGJtVXRiSHQwWlhoMExXUmxZMjl5WVhScGIyNDZkVzVrWlhKc2FXNWxmUzV1YnkxMWJtUmxjbXhwYm1VdGJIdDBaWGgwTFdSbFkyOXlZWFJwYjI0NmJtOXVaWDB1ZEd3dGJIdDBaWGgwTFdGc2FXZHVPbXhsWm5SOUxuUnlMV3g3ZEdWNGRDMWhiR2xuYmpweWFXZG9kSDB1ZEdNdGJIdDBaWGgwTFdGc2FXZHVPbU5sYm5SbGNuMHVkR290Ykh0MFpYaDBMV0ZzYVdkdU9tcDFjM1JwWm5sOUxuUjBZeTFzZTNSbGVIUXRkSEpoYm5ObWIzSnRPbU5oY0dsMFlXeHBlbVY5TG5SMGJDMXNlM1JsZUhRdGRISmhibk5tYjNKdE9teHZkMlZ5WTJGelpYMHVkSFIxTFd4N2RHVjRkQzEwY21GdWMyWnZjbTA2ZFhCd1pYSmpZWE5sZlM1MGRHNHRiSHQwWlhoMExYUnlZVzV6Wm05eWJUcHViMjVsZlM1bUxUWXRiQ3d1Wmkxb1pXRmtiR2x1WlMxc2UyWnZiblF0YzJsNlpUbzJjbVZ0ZlM1bUxUVXRiQ3d1WmkxemRXSm9aV0ZrYkdsdVpTMXNlMlp2Ym5RdGMybDZaVG8xY21WdGZTNW1NUzFzZTJadmJuUXRjMmw2WlRvemNtVnRmUzVtTWkxc2UyWnZiblF0YzJsNlpUb3lMakkxY21WdGZTNW1NeTFzZTJadmJuUXRjMmw2WlRveExqVnlaVzE5TG1ZMExXeDdabTl1ZEMxemFYcGxPakV1TWpWeVpXMTlMbVkxTFd4N1ptOXVkQzF6YVhwbE9qRnlaVzE5TG1ZMkxXeDdabTl1ZEMxemFYcGxPaTQ0TnpWeVpXMTlMbVkzTFd4N1ptOXVkQzF6YVhwbE9pNDNOWEpsYlgwdWJXVmhjM1Z5WlMxc2UyMWhlQzEzYVdSMGFEb3pNR1Z0ZlM1dFpXRnpkWEpsTFhkcFpHVXRiSHR0WVhndGQybGtkR2c2TXpSbGJYMHViV1ZoYzNWeVpTMXVZWEp5YjNjdGJIdHRZWGd0ZDJsa2RHZzZNakJsYlgwdWFXNWtaVzUwTFd4N2RHVjRkQzFwYm1SbGJuUTZNV1Z0TzIxaGNtZHBiaTEwYjNBNk1EdHRZWEpuYVc0dFltOTBkRzl0T2pCOUxuTnRZV3hzTFdOaGNITXRiSHRtYjI1MExYWmhjbWxoYm5RNmMyMWhiR3d0WTJGd2MzMHVkSEoxYm1OaGRHVXRiSHQzYUdsMFpTMXpjR0ZqWlRwdWIzZHlZWEE3YjNabGNtWnNiM2M2YUdsa1pHVnVPM1JsZUhRdGIzWmxjbVpzYjNjNlpXeHNhWEJ6YVhOOUxtTmxiblJsY2kxc2UyMWhjbWRwYmkxc1pXWjBPbUYxZEc5OUxtTmxiblJsY2kxc0xDNXRjaTFoZFhSdkxXeDdiV0Z5WjJsdUxYSnBaMmgwT21GMWRHOTlMbTFzTFdGMWRHOHRiSHR0WVhKbmFXNHRiR1ZtZERwaGRYUnZmUzVqYkdsd0xXeDdjRzl6YVhScGIyNDZabWw0WldRaGFXMXdiM0owWVc1ME8xOXdiM05wZEdsdmJqcGhZbk52YkhWMFpTRnBiWEJ2Y25SaGJuUTdZMnhwY0RweVpXTjBLREZ3ZUNBeGNIZ2dNWEI0SURGd2VDazdZMnhwY0RweVpXTjBLREZ3ZUN3eGNIZ3NNWEI0TERGd2VDbDlMbmR6TFc1dmNtMWhiQzFzZTNkb2FYUmxMWE53WVdObE9tNXZjbTFoYkgwdWJtOTNjbUZ3TFd4N2QyaHBkR1V0YzNCaFkyVTZibTkzY21Gd2ZTNXdjbVV0Ykh0M2FHbDBaUzF6Y0dGalpUcHdjbVY5TG5ZdFltRnpaUzFzZTNabGNuUnBZMkZzTFdGc2FXZHVPbUpoYzJWc2FXNWxmUzUyTFcxcFpDMXNlM1psY25ScFkyRnNMV0ZzYVdkdU9tMXBaR1JzWlgwdWRpMTBiM0F0Ykh0MlpYSjBhV05oYkMxaGJHbG5ianAwYjNCOUxuWXRZblJ0TFd4N2RtVnlkR2xqWVd3dFlXeHBaMjQ2WW05MGRHOXRmWDFjYmx4dUlsMTkgKi88L3N0eWxlPlxuXG48VG9wTWVudSAvPlxuPG1haW4gY2xhc3M9XCJ3LTEwMCBidCBiLS1ibGFjay0xMCBiZy13aGl0ZVwiPlxuICA8IS0tIDxNZW51IC8+IC0tPlxuICA8Um91dGVyXG4gICAge3JvdXRlc31cbiAgICBvbjpjb25kaXRpb25zRmFpbGVkPXtjb25kaXRpb25zRmFpbGVkfVxuICAgIG9uOnJvdXRlTG9hZGVkPXtyb3V0ZUxvYWRlZH0gLz5cbjwvbWFpbj5cblxuPEZvb3RlciAvPlxuPG1haW4gY2xhc3M9XCJ3LTEwMCBidCBiLS1ibGFjay0xMCBiZy13aGl0ZVwiPlxuXG4gIDwhLS0gVXNlZCBmb3IgdGVzdGluZ1xuPHAgaWQ9XCJsb2dib3hcIj57bG9nYm94fTwvcD5cbiAtLT5cbjwvbWFpbj5cbiJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFnRFEsRUFBRSxBQUFDLENBQUMsVUFBVSxHQUFHLENBQUMsT0FBTyxLQUFLLENBQUMsQ0FBQyxDQUFDLEFBQ2pDLENBQUMsQUFBQyxDQUFDLGlCQUFpQixXQUFXLENBQUMsQUFDaEMsQ0FBQyxBQUFDLENBQUMsWUFBWSxNQUFNLENBQUMsQUFDdEIsS0FBSyxBQUFDLENBQUMsVUFBVSxHQUFHLENBQUMsQUFDckIsTUFBTSxBQUFDLENBQUMsQUFBUSxLQUFLLEFBQUMsQ0FBQyxBQUFRLE1BQU0sQUFBQyxDQUFDLEFBQVEsUUFBUSxBQUFDLENBQUMsWUFBWSxPQUFPLENBQUMsVUFBVSxJQUFJLENBQUMsWUFBWSxJQUFJLENBQUMsT0FBTyxDQUFDLENBQUMsQUFDdEgsTUFBTSxBQUFDLENBQUMsQUFBUSxLQUFLLEFBQUMsQ0FBQyxTQUFTLE9BQU8sQ0FBQyxBQUN4QyxNQUFNLEFBQUMsQ0FBQyxBQUFRLE1BQU0sQUFBQyxDQUFDLGVBQWUsSUFBSSxDQUFDLEFBQzVDLE1BQU0sQUFBQyxDQUFDLG1CQUFtQixNQUFNLENBQUMsQUFDbEMsd0JBQXdCLEFBQUMsQ0FBQyxhQUFhLElBQUksQ0FBQyxRQUFRLENBQUMsQ0FBQyxBQUN0RCxxQkFBcUIsQUFBQyxDQUFDLFFBQVEsR0FBRyxDQUFDLE1BQU0sQ0FBQyxVQUFVLENBQUMsQUFDckQsUUFBUSxBQUFDLENBQUMsUUFBUSxLQUFLLENBQUMsS0FBSyxDQUFDLE1BQU0sQ0FBQyxBQUNyQyxRQUFRLEFBQUMsQ0FBQyxTQUFTLElBQUksQ0FBQyxBQUN4QixDQUFDLEFBQUMsQ0FBQyxBQUFRLE9BQU8sQUFBQyxDQUFDLEFBQVEsVUFBVSxBQUFDLENBQUMsQUFBUSxHQUFHLEFBQUMsQ0FBQyxBQUFRLEVBQUUsQUFBQyxDQUFDLEFBQVEsUUFBUSxBQUFDLENBQUMsQUFBUSxJQUFJLEFBQUMsQ0FBQyxBQUFRLEVBQUUsQUFBQyxDQUFDLEFBQVEsRUFBRSxBQUFDLENBQUMsQUFBUSxFQUFFLEFBQUMsQ0FBQyxBQUFRLE1BQU0sQUFBQyxDQUFDLEFBQVEsRUFBRSxBQUFDLENBQUMsQUFBUSxJQUFJLEFBQUMsQ0FBQyxBQUFRLENBQUMsQUFBQyxDQUFDLEFBQVEsT0FBTyxBQUFDLENBQUMsQUFBUSxRQUFRLEFBQUMsQ0FBQyxBQUFRLEVBQUUsQUFBQyxDQUFDLFdBQVcsVUFBVSxDQUFDLEFBQy9QLEdBQUcsQUFBQyxDQUFDLGFBQWEsS0FBSyxDQUFDLGFBQWEsR0FBRyxDQUFDLEFBQ3pDLEdBQUcsQUFBQyxDQUFDLGlCQUFpQixLQUFLLENBQUMsaUJBQWlCLEdBQUcsQ0FBQyxBQUNqRCxHQUFHLEFBQUMsQ0FBQyxtQkFBbUIsS0FBSyxDQUFDLG1CQUFtQixHQUFHLENBQUMsQUFDckQsR0FBRyxBQUFDLENBQUMsb0JBQW9CLEtBQUssQ0FBQyxvQkFBb0IsR0FBRyxDQUFDLEFBQ3ZELEdBQUcsQUFBQyxDQUFDLGtCQUFrQixLQUFLLENBQUMsa0JBQWtCLEdBQUcsQ0FBQyxBQUNuRCxHQUFHLEFBQUMsQ0FBQyxhQUFhLElBQUksQ0FBQyxhQUFhLENBQUMsQ0FBQyxBQUN0QyxTQUFTLEFBQUMsQ0FBQyxhQUFhLElBQUksQ0FBQyxBQUM3QixhQUFhLEFBQUMsQ0FBQyxhQUFhLElBQUksQ0FBQyxBQUNqQyxZQUFZLEFBQUMsQ0FBQyxhQUFhLEtBQUssQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsRUFBRSxDQUFDLENBQUMsQUFDMUMsWUFBWSxBQUFDLENBQUMsYUFBYSxLQUFLLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxDQUFDLEFBQzFDLFlBQVksQUFBQyxDQUFDLGFBQWEsS0FBSyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxFQUFFLENBQUMsQ0FBQyxBQUMxQyxZQUFZLEFBQUMsQ0FBQyxhQUFhLEtBQUssQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsQUFDM0MsT0FBTyxBQUFDLENBQUMsYUFBYSxPQUFPLENBQUMsQUFDOUIsVUFBVSxBQUFDLENBQUMsYUFBYSxJQUFJLENBQUMsQUFDOUIsaUJBQWlCLEFBQUMsQ0FBQyxhQUFhLE9BQU8sQ0FBQyxBQUN4QyxlQUFlLEFBQUMsQ0FBQyxhQUFhLFdBQVcsQ0FBQyxBQUMxQyxJQUFJLEFBQUMsQ0FBQyxjQUFjLE9BQU8sQ0FBQyxBQUM1QixJQUFJLEFBQUMsQ0FBQyxjQUFjLE1BQU0sQ0FBQyxBQUMzQixJQUFJLEFBQUMsQ0FBQyxjQUFjLEtBQUssQ0FBQyxBQUMxQixVQUFVLEFBQUMsQ0FBQyxhQUFhLE1BQU0sQ0FBQyxBQUNoQyxJQUFJLEFBQUMsQ0FBQyxhQUFhLE9BQU8sQ0FBQyxBQUMzQixJQUFJLEFBQUMsQ0FBQyxhQUFhLE1BQU0sQ0FBQyxBQUMxQixJQUFJLEFBQUMsQ0FBQyxhQUFhLEtBQUssQ0FBQyxBQUN6QixLQUFLLEFBQUMsQ0FBQyxpQkFBaUIsQ0FBQyxDQUFDLEFBQzFCLEtBQUssQUFBQyxDQUFDLG1CQUFtQixDQUFDLENBQUMsQUFDNUIsS0FBSyxBQUFDLENBQUMsa0JBQWtCLENBQUMsQ0FBQyxBQUMzQixHQUFHLEFBQUMsQ0FBQyxRQUFRLE1BQU0sQ0FBQyxBQUNwQixHQUFHLEFBQUMsQ0FBQyxRQUFRLEtBQUssQ0FBQyxBQUNuQixJQUFJLEFBQUMsQ0FBQyxRQUFRLFlBQVksQ0FBQyxBQUMzQixHQUFHLEFBQUMsQ0FBQyxRQUFRLEtBQUssQ0FBQyxBQUNuQixJQUFJLEFBQUMsQ0FBQyxRQUFRLFVBQVUsQ0FBQyxBQUN6QixLQUFLLEFBQUMsQ0FBQyxRQUFRLElBQUksQ0FBQyxBQUNwQixZQUFZLEFBQUMsQ0FBQyxlQUFlLE1BQU0sQ0FBQyxBQUNwQyxVQUFVLEFBQUMsQ0FBQyxZQUFZLFFBQVEsQ0FBQyxBQUNqQyxhQUFhLEFBQUMsQ0FBQyxZQUFZLE1BQU0sQ0FBQyxBQUNsQyxlQUFlLEFBQUMsQ0FBQyxnQkFBZ0IsTUFBTSxDQUFDLEFBQ3hDLEdBQUcsQUFBQyxDQUFDLE1BQU0sSUFBSSxDQUFDLEFBQ2hCLEdBQUcsQUFBQyxDQUFDLEFBQVEsR0FBRyxBQUFDLENBQUMsU0FBUyxNQUFNLENBQUMsQUFDbEMsR0FBRyxBQUFDLENBQUMsTUFBTSxLQUFLLENBQUMsQUFDakIsUUFBUSxBQUFDLENBQUMsWUFBWSxPQUFPLENBQUMsT0FBTyxDQUFDLEtBQUssQ0FBQyxBQUM1QyxVQUFVLEFBQUMsQ0FBQyxXQUFXLE1BQU0sQ0FBQyxBQUM5QixFQUFFLEFBQUMsQ0FBQyxZQUFZLEdBQUcsQ0FBQyxBQUNwQixJQUFJLEFBQUMsQ0FBQyxZQUFZLEdBQUcsQ0FBQyxBQUN0QixJQUFJLEFBQUMsQ0FBQyxZQUFZLEdBQUcsQ0FBQyxBQUN0QixJQUFJLEFBQUMsQ0FBQyxZQUFZLEdBQUcsQ0FBQyxBQUN0QixJQUFJLEFBQUMsQ0FBQyxZQUFZLEdBQUcsQ0FBQyxBQUN0QixJQUFJLEFBQUMsQ0FBQyxZQUFZLEdBQUcsQ0FBQyxBQUN0QixZQUFZLEFBQUMsQ0FBQyxtQkFBbUIsSUFBSSxDQUFDLGdCQUFnQixJQUFJLENBQUMsQUFDM0QsOEJBQThCLEFBQUMsQ0FBQyxPQUFPLENBQUMsQ0FBQyxRQUFRLENBQUMsQ0FBQyxBQUNuRCxHQUFHLEFBQUMsQ0FBQyxPQUFPLElBQUksQ0FBQyxBQUNqQixHQUFHLEFBQUMsQ0FBQyxPQUFPLElBQUksQ0FBQyxBQUNqQixHQUFHLEFBQUMsQ0FBQyxPQUFPLElBQUksQ0FBQyxBQUNqQixRQUFRLEFBQUMsQ0FBQyxlQUFlLElBQUksQ0FBQyxBQUM5QixTQUFTLEFBQUMsQ0FBQyxZQUFZLElBQUksQ0FBQyxBQUM1QixRQUFRLEFBQUMsQ0FBQyxZQUFZLEdBQUcsQ0FBQyxBQUMxQixLQUFLLEFBQUMsQ0FBQyxnQkFBZ0IsSUFBSSxDQUFDLEFBQzVCLEtBQUssQUFBQyxDQUFDLEFBQVEsWUFBWSxBQUFDLENBQUMsQUFBUSxXQUFXLEFBQUMsQ0FBQyxBQUFRLFdBQVcsQUFBQyxDQUFDLEFBQVEsVUFBVSxBQUFDLENBQUMsQUFBUSxhQUFhLEFBQUMsQ0FBQyxXQUFXLEtBQUssQ0FBQyxJQUFJLENBQUMsT0FBTyxDQUFDLEFBQ2hKLFdBQVcsQUFBQyxDQUFDLFFBQVEsR0FBRyxDQUFDLE1BQU0sQ0FBQyxZQUFZLENBQUMsQUFDN0MsS0FBSyxBQUFDLENBQUMsZ0JBQWdCLElBQUksQ0FBQyxBQUM1QixPQUFPLEFBQUMsQ0FBQyxVQUFVLElBQUksQ0FBQyxBQUN4QixHQUFHLEFBQUMsQ0FBQyxNQUFNLElBQUksQ0FBQyxBQUNoQixLQUFLLEFBQUMsQ0FBQyxNQUFNLEdBQUcsQ0FBQyxBQUNqQixLQUFLLEFBQUMsQ0FBQyxNQUFNLEdBQUcsQ0FBQyxBQUNqQixNQUFNLEFBQUMsQ0FBQyxNQUFNLElBQUksQ0FBQyxBQUNuQixTQUFTLEFBQUMsQ0FBQyxNQUFNLEtBQUssQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsRUFBRSxDQUFDLENBQUMsQUFDaEMsU0FBUyxBQUFDLENBQUMsTUFBTSxLQUFLLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxDQUFDLEFBQ2hDLFNBQVMsQUFBQyxDQUFDLE1BQU0sS0FBSyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxFQUFFLENBQUMsQ0FBQyxBQUNoQyxTQUFTLEFBQUMsQ0FBQyxNQUFNLEtBQUssQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsRUFBRSxDQUFDLENBQUMsQUFDaEMsU0FBUyxBQUFDLENBQUMsTUFBTSxLQUFLLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxDQUFDLEFBQ2hDLFNBQVMsQUFBQyxDQUFDLE1BQU0sS0FBSyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxFQUFFLENBQUMsQ0FBQyxBQUNoQyxNQUFNLEFBQUMsQ0FBQyxNQUFNLElBQUksQ0FBQyxBQUNuQixVQUFVLEFBQUMsQ0FBQyxNQUFNLElBQUksQ0FBQyxBQUN2QixNQUFNLEFBQUMsQ0FBQyxNQUFNLElBQUksQ0FBQyxBQUNuQixLQUFLLEFBQUMsQ0FBQyxNQUFNLE9BQU8sQ0FBQyxBQUNyQixTQUFTLEFBQUMsQ0FBQyxpQkFBaUIsSUFBSSxDQUFDLEFBQ2pDLGVBQWUsQUFBQyxDQUFDLGlCQUFpQixXQUFXLENBQUMsQUFDOUMsYUFBYSxBQUFDLENBQUMsaUJBQWlCLE9BQU8sQ0FBQyxBQUN4QyxpQkFBaUIsQUFBQyxDQUFDLGlCQUFpQixPQUFPLENBQUMsQUFDNUMsY0FBYyxBQUFDLENBQUMsaUJBQWlCLE9BQU8sQ0FBQyxBQUN6QyxnQkFBZ0IsQUFBQyxDQUFDLEFBQVEsZ0JBQWdCLEFBQUMsQ0FBQyxNQUFNLE9BQU8sQ0FBQyxBQUMxRCxpQkFBaUIsQUFBQyxDQUFDLEFBQVEsaUJBQWlCLEFBQUMsQ0FBQyxNQUFNLE9BQU8sQ0FBQyxBQUM1RCxJQUFJLEFBQUMsQ0FBQyxRQUFRLENBQUMsQ0FBQyxBQUNoQixJQUFJLEFBQUMsQ0FBQyxRQUFRLE1BQU0sQ0FBQyxBQUNyQixJQUFJLEFBQUMsQ0FBQyxRQUFRLEtBQUssQ0FBQyxBQUNwQixJQUFJLEFBQUMsQ0FBQyxRQUFRLElBQUksQ0FBQyxBQUNuQixJQUFJLEFBQUMsQ0FBQyxRQUFRLElBQUksQ0FBQyxBQUNuQixJQUFJLEFBQUMsQ0FBQyxhQUFhLENBQUMsQ0FBQyxBQUNyQixJQUFJLEFBQUMsQ0FBQyxhQUFhLElBQUksQ0FBQyxBQUN4QixJQUFJLEFBQUMsQ0FBQyxjQUFjLEtBQUssQ0FBQyxBQUMxQixJQUFJLEFBQUMsQ0FBQyxlQUFlLEtBQUssQ0FBQyxBQUMzQixJQUFJLEFBQUMsQ0FBQyxZQUFZLEtBQUssQ0FBQyxBQUN4QixJQUFJLEFBQUMsQ0FBQyxZQUFZLElBQUksQ0FBQyxBQUN2QixJQUFJLEFBQUMsQ0FBQyxZQUFZLENBQUMsQ0FBQyxlQUFlLENBQUMsQ0FBQyxBQUNyQyxJQUFJLEFBQUMsQ0FBQyxZQUFZLE1BQU0sQ0FBQyxlQUFlLE1BQU0sQ0FBQyxBQUMvQyxJQUFJLEFBQUMsQ0FBQyxZQUFZLEtBQUssQ0FBQyxlQUFlLEtBQUssQ0FBQyxBQUM3QyxJQUFJLEFBQUMsQ0FBQyxZQUFZLElBQUksQ0FBQyxlQUFlLElBQUksQ0FBQyxBQUMzQyxJQUFJLEFBQUMsQ0FBQyxZQUFZLElBQUksQ0FBQyxlQUFlLElBQUksQ0FBQyxBQUMzQyxJQUFJLEFBQUMsQ0FBQyxhQUFhLENBQUMsQ0FBQyxjQUFjLENBQUMsQ0FBQyxBQUNyQyxJQUFJLEFBQUMsQ0FBQyxhQUFhLE1BQU0sQ0FBQyxjQUFjLE1BQU0sQ0FBQyxBQUMvQyxJQUFJLEFBQUMsQ0FBQyxhQUFhLElBQUksQ0FBQyxjQUFjLElBQUksQ0FBQyxBQUMzQyxJQUFJLEFBQUMsQ0FBQyxhQUFhLElBQUksQ0FBQyxjQUFjLElBQUksQ0FBQyxBQUMzQyxJQUFJLEFBQUMsQ0FBQyxZQUFZLENBQUMsQ0FBQyxBQUNwQixJQUFJLEFBQUMsQ0FBQyxZQUFZLElBQUksQ0FBQyxBQUN2QixJQUFJLEFBQUMsQ0FBQyxhQUFhLE1BQU0sQ0FBQyxBQUMxQixJQUFJLEFBQUMsQ0FBQyxhQUFhLEtBQUssQ0FBQyxBQUN6QixJQUFJLEFBQUMsQ0FBQyxjQUFjLENBQUMsQ0FBQyxBQUN0QixJQUFJLEFBQUMsQ0FBQyxjQUFjLE1BQU0sQ0FBQyxBQUMzQixJQUFJLEFBQUMsQ0FBQyxjQUFjLEtBQUssQ0FBQyxBQUMxQixJQUFJLEFBQUMsQ0FBQyxjQUFjLElBQUksQ0FBQyxBQUN6QixJQUFJLEFBQUMsQ0FBQyxjQUFjLElBQUksQ0FBQyxBQUN6QixJQUFJLEFBQUMsQ0FBQyxXQUFXLENBQUMsQ0FBQyxBQUNuQixJQUFJLEFBQUMsQ0FBQyxXQUFXLEtBQUssQ0FBQyxBQUN2QixJQUFJLEFBQUMsQ0FBQyxXQUFXLElBQUksQ0FBQyxBQUN0QixJQUFJLEFBQUMsQ0FBQyxXQUFXLElBQUksQ0FBQyxBQUN0QixJQUFJLEFBQUMsQ0FBQyxXQUFXLENBQUMsQ0FBQyxjQUFjLENBQUMsQ0FBQyxBQUNuQyxJQUFJLEFBQUMsQ0FBQyxXQUFXLEtBQUssQ0FBQyxjQUFjLEtBQUssQ0FBQyxBQUMzQyxJQUFJLEFBQUMsQ0FBQyxXQUFXLElBQUksQ0FBQyxjQUFjLElBQUksQ0FBQyxBQUN6QyxJQUFJLEFBQUMsQ0FBQyxZQUFZLENBQUMsQ0FBQyxhQUFhLENBQUMsQ0FBQyxBQUNuQyxJQUFJLEFBQUMsQ0FBQyxZQUFZLE1BQU0sQ0FBQyxhQUFhLE1BQU0sQ0FBQyxBQUM3QyxVQUFVLEFBQUMsQ0FBQyxnQkFBZ0IsU0FBUyxDQUFDLEFBQ3RDLEdBQUcsQUFBQyxDQUFDLFdBQVcsTUFBTSxDQUFDLEFBQ3ZCLElBQUksQUFBQyxDQUFDLGVBQWUsU0FBUyxDQUFDLEFBQy9CLEdBQUcsQUFBQyxDQUFDLFVBQVUsT0FBTyxDQUFDLEFBQ3ZCLEdBQUcsQUFBQyxDQUFDLFVBQVUsTUFBTSxDQUFDLEFBQ3RCLEdBQUcsQUFBQyxDQUFDLFVBQVUsT0FBTyxDQUFDLEFBQ3ZCLEdBQUcsQUFBQyxDQUFDLFVBQVUsSUFBSSxDQUFDLEFBQ3BCLEdBQUcsQUFBQyxDQUFDLFVBQVUsT0FBTyxDQUFDLEFBQ3ZCLFFBQVEsQUFBQyxDQUFDLFVBQVUsSUFBSSxDQUFDLEFBQ3pCLE9BQU8sQUFBQyxDQUFDLFlBQVksSUFBSSxDQUFDLEFBQzFCLE9BQU8sQUFBQyxDQUFDLGFBQWEsSUFBSSxDQUFDLEFBQzNCLE9BQU8sQUFBQyxDQUFDLFlBQVksTUFBTSxDQUFDLEFBQzVCLE1BQU0sQUFBQyxDQUFDLGVBQWUsTUFBTSxDQUFDLEFBQzlCLElBQUksQUFBQyxDQUFDLFFBQVEsQ0FBQyxDQUFDLEFBQ2hCLElBQUksQUFBQyxDQUFDLEFBQVEsVUFBVSxBQUFDLENBQUMsQUFBUSxVQUFVLEFBQUMsQ0FBQyxXQUFXLE9BQU8sQ0FBQyxJQUFJLENBQUMsT0FBTyxDQUFDLEFBQzlFLFVBQVUsQUFBQyxDQUFDLEFBQVEsVUFBVSxBQUFDLENBQUMsUUFBUSxFQUFFLENBQUMsQUFDM0MsV0FBVyxBQUFDLENBQUMsUUFBUSxFQUFFLENBQUMsV0FBVyxPQUFPLENBQUMsSUFBSSxDQUFDLFFBQVEsQ0FBQyxBQUNqRSxPQUFPLE1BQU0sQ0FBQyxHQUFHLENBQUMsV0FBVyxJQUFJLENBQUMsQ0FBQyxBQUFRLE9BQU8sQUFBQyxDQUFDLFFBQVEsS0FBSyxDQUFDLEFBQVEsT0FBTyxBQUFDLENBQUMsUUFBUSxJQUFJLENBQUMsQUFBUSxPQUFPLEFBQUMsQ0FBQyxhQUFhLE1BQU0sQ0FBQyxjQUFjLE1BQU0sQ0FBQyxBQUFRLE9BQU8sQUFBQyxDQUFDLGFBQWEsSUFBSSxDQUFDLGNBQWMsSUFBSSxDQUFDLEFBQVEsT0FBTyxBQUFDLENBQUMsYUFBYSxJQUFJLENBQUMsQUFBUSxPQUFPLEFBQUMsQ0FBQyxXQUFXLElBQUksQ0FBQyxBQUFRLE1BQU0sQUFBQyxDQUFDLFVBQVUsT0FBTyxDQUFDLEFBQVEsTUFBTSxBQUFDLENBQUMsVUFBVSxPQUFPLENBQUMsQUFBUSxNQUFNLEFBQUMsQ0FBQyxVQUFVLElBQUksQ0FBQyxDQUFDLEFBQ3hYLE9BQU8sTUFBTSxDQUFDLEdBQUcsQ0FBQyxXQUFXLElBQUksQ0FBQyxDQUFDLEdBQUcsQ0FBQyxXQUFXLElBQUksQ0FBQyxDQUFDLEFBQVEsTUFBTSxBQUFDLENBQUMsYUFBYSxJQUFJLENBQUMsQUFBUSxLQUFLLEFBQUMsQ0FBQyxVQUFVLE9BQU8sQ0FBQyxDQUFDLEFBQzVILE9BQU8sTUFBTSxDQUFDLEdBQUcsQ0FBQyxXQUFXLElBQUksQ0FBQyxDQUFDLEFBQVEsTUFBTSxBQUFDLENBQUMsYUFBYSxJQUFJLENBQUMsY0FBYyxJQUFJLENBQUMsQUFBUSxNQUFNLEFBQUMsQ0FBQyxhQUFhLEtBQUssQ0FBQyxBQUFRLE1BQU0sQUFBQyxDQUFDLGFBQWEsSUFBSSxDQUFDLEFBQVEsTUFBTSxBQUFDLENBQUMsYUFBYSxJQUFJLENBQUMsQUFBUSxLQUFLLEFBQUMsQ0FBQyxVQUFVLE1BQU0sQ0FBQyxDQUFDIn0= */";
	append_dev(document.head, style);
}

function create_fragment$z(ctx) {
	let t0;
	let main0;
	let t1;
	let t2;
	let main1;
	let current;
	const topmenu = new Menu_top({ $$inline: true });
	const router = new Router({ props: { routes }, $$inline: true });
	router.$on("conditionsFailed", conditionsFailed);
	router.$on("routeLoaded", routeLoaded);
	const footer = new Footer({ $$inline: true });

	const block = {
		c: function create() {
			create_component(topmenu.$$.fragment);
			t0 = space();
			main0 = element("main");
			create_component(router.$$.fragment);
			t1 = space();
			create_component(footer.$$.fragment);
			t2 = space();
			main1 = element("main");
			attr_dev(main0, "class", "w-100 bt b--black-10 bg-white");
			add_location(main0, file$y, 203, 0, 111662);
			attr_dev(main1, "class", "w-100 bt b--black-10 bg-white");
			add_location(main1, file$y, 212, 0, 111849);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			mount_component(topmenu, target, anchor);
			insert_dev(target, t0, anchor);
			insert_dev(target, main0, anchor);
			mount_component(router, main0, null);
			insert_dev(target, t1, anchor);
			mount_component(footer, target, anchor);
			insert_dev(target, t2, anchor);
			insert_dev(target, main1, anchor);
			current = true;
		},
		p: noop,
		i: function intro(local) {
			if (current) return;
			transition_in(topmenu.$$.fragment, local);
			transition_in(router.$$.fragment, local);
			transition_in(footer.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(topmenu.$$.fragment, local);
			transition_out(router.$$.fragment, local);
			transition_out(footer.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(topmenu, detaching);
			if (detaching) detach_dev(t0);
			if (detaching) detach_dev(main0);
			destroy_component(router);
			if (detaching) detach_dev(t1);
			destroy_component(footer, detaching);
			if (detaching) detach_dev(t2);
			if (detaching) detach_dev(main1);
		}
	};

	dispatch_dev("SvelteRegisterBlock", {
		block,
		id: create_fragment$z.name,
		type: "component",
		source: "",
		ctx
	});

	return block;
}

function conditionsFailed(event) {
	// eslint-disable-next-line no-console
	// console.error('Caught event conditionsFailed', event.detail)
	// logbox += 'conditionsFailed - ' + JSON.stringify(event.detail) + '\n'
	// Replace the route
	replace("/login");
}

// Handles the "routeLoaded" event dispatched by the router after a route has been successfully loaded
function routeLoaded(event) {
	// eslint-disable-next-line no-console
	// console.info('Caught event routeLoaded', event.detail)
	// logbox += 'routeLoaded - ' + JSON.stringify(event.detail) + '\n'
	cache.save(cache.keys["last.screen"], "#" + event.detail.location);
}

function instance$z($$self, $$props, $$invalidate) {
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<App> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("App", $$slots, []);

	$$self.$capture_state = () => ({
		cache,
		Router,
		TopMenu: Menu_top,
		Footer,
		Menu,
		Box,
		link,
		push,
		pop,
		replace,
		location: location$1,
		querystring,
		active,
		routes,
		conditionsFailed,
		routeLoaded
	});

	return [];
}

class App extends SvelteComponentDev {
	constructor(options) {
		super(options);
		if (!document.getElementById("svelte-64n3a6-style")) add_css$9();
		init(this, options, instance$z, create_fragment$z, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "App",
			options,
			id: create_fragment$z.name
		});
	}
}

if (!localStorage.hasOwnProperty(cache.keys['settings.install.defaults'])) {
    cache.clear();
    window.location = location.origin + location.pathname;
}

// Specific for the chrome extension
if (window.location.protocol === 'chrome-extension:') {
    let last = cache.get(cache.keys['last.screen']);
    console.log(last);
    if (last) {
        if (last !== window.location.hash) {
            history.replaceState(undefined, undefined, last);
            window.dispatchEvent(new Event('hashchange'));
        }
    }
}

if (window.location.protocol !== 'chrome-extension:') {
    let last = cache.get(cache.keys['last.screen']);
    if (last) {
        if (last !== window.location.hash) {
            history.replaceState(undefined, undefined, (location.origin + location.pathname + last));
            window.dispatchEvent(new Event('hashchange'));
        }
    }
}


var app = new App({
    target: document.body
});

export default app;
//# sourceMappingURL=editor.1585943759773.js.map
