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

const globals = (typeof window !== 'undefined'
    ? window
    : typeof globalThis !== 'undefined'
        ? globalThis
        : global);

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
    document.dispatchEvent(custom_event(type, Object.assign({ version: '3.21.0' }, detail)));
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
  'authentication.bearer': 'settings.authentication',
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
  const apiServer = document.querySelector('meta[name="api.server"]');
  if (apiServer) {
    save(keys['settings.server'], apiServer.content);
  } else {
    save(keys['settings.server'], 'https://learnalist.net');
  }
  // TODO why is this not showing up?
  save(keys['my.edited.lists'], []);
  save(keys['lists.by.me'], []);
}

var cache$1 = {
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

/* node_modules/svelte-spa-router/Router.svelte generated by Svelte v3.21.0 */

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

const KeySettingsServer = "settings.server";
const KeyUserAuthentication = "app.user.authentication";

function get$1(key, defaultResult) {
  if (!localStorage.hasOwnProperty(key)) {
    return defaultResult;
  }

  return JSON.parse(localStorage.getItem(key))
}

function save$1(key, data) {
  localStorage.setItem(key, JSON.stringify(data));
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
			const auth = get$1(KeyUserAuthentication);
			return auth ? true : false;
		})()
	};

	const { subscribe, set, update } = writable(obj);

	return {
		subscribe,

		login: ((session) => {
			save$1(KeyUserAuthentication, session.token);
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

/* src/editor/components/menu_top.svelte generated by Svelte v3.21.0 */
const file = "src/editor/components/menu_top.svelte";

function create_fragment$1(ctx) {
	let header;
	let div0;
	let t0;
	let div2;
	let div1;
	let a0;
	let link_action;
	let t2;
	let a1;
	let link_action_1;
	let dispose;

	const block = {
		c: function create() {
			header = element("header");
			div0 = element("div");
			t0 = space();
			div2 = element("div");
			div1 = element("div");
			a0 = element("a");
			a0.textContent = "Create";
			t2 = space();
			a1 = element("a");
			a1.textContent = "Find";
			attr_dev(div0, "class", "w-25 pa3 mr2");
			add_location(div0, file, 6, 2, 148);
			attr_dev(a0, "title", "Documentation");
			attr_dev(a0, "href", "/create");
			attr_dev(a0, "class", "f6 fw6 hover-blue link black-70 mr2 mr3-m mr2-l di");
			add_location(a0, file, 10, 6, 244);
			attr_dev(a1, "title", "Components");
			attr_dev(a1, "href", "/lists/by/me");
			attr_dev(a1, "class", "f6 fw6 hover-blue link black-70 mr2 mr3-m mr5-l di");
			add_location(a1, file, 17, 6, 417);
			attr_dev(div1, "class", "fr mt0");
			add_location(div1, file, 9, 4, 217);
			attr_dev(div2, "class", "w-75 pa3 items-end");
			add_location(div2, file, 8, 2, 180);
			attr_dev(header, "class", "flex");
			add_location(header, file, 5, 0, 124);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor, remount) {
			insert_dev(target, header, anchor);
			append_dev(header, div0);
			append_dev(header, t0);
			append_dev(header, div2);
			append_dev(div2, div1);
			append_dev(div1, a0);
			append_dev(div1, t2);
			append_dev(div1, a1);
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
			if (detaching) detach_dev(header);
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

function instance$1($$self, $$props, $$invalidate) {
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Menu_top> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("Menu_top", $$slots, []);
	$$self.$capture_state = () => ({ link, location: location$1, loginHelper });
	return [];
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

/* src/editor/components/footer.svelte generated by Svelte v3.21.0 */
const file$1 = "src/editor/components/footer.svelte";

// (8:4) {#if $loginHelper.loggedIn}
function create_if_block$1(ctx) {
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
		id: create_if_block$1.name,
		type: "if",
		source: "(8:4) {#if $loginHelper.loggedIn}",
		ctx
	});

	return block;
}

function create_fragment$2(ctx) {
	let header;
	let div;
	let if_block = /*$loginHelper*/ ctx[0].loggedIn && create_if_block$1(ctx);

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
				if (if_block) ; else {
					if_block = create_if_block$1(ctx);
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

/* src/editor/components/menu.svelte generated by Svelte v3.21.0 */
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

/* src/editor/components/Box.svelte generated by Svelte v3.21.0 */

const file$3 = "src/editor/components/Box.svelte";

function create_fragment$4(ctx) {
	let div;
	let current;
	const default_slot_template = /*$$slots*/ ctx[1].default;
	const default_slot = create_slot(default_slot_template, ctx, /*$$scope*/ ctx[0], null);

	const block = {
		c: function create() {
			div = element("div");
			if (default_slot) default_slot.c();
			attr_dev(div, "class", "box svelte-wrrycl");
			add_location(div, file$3, 11, 0, 708);
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

/* src/editor/routes/home.svelte generated by Svelte v3.21.0 */
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
	cache$1.clear();
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
	$$self.$capture_state = () => ({ link, replace, loginHelper, cache: cache$1, reset });
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
  const token = get$1(KeyUserAuthentication, null);
  if (token === null) {
    throw new Error('login.required');
  }
  return `Bearer ${token}`;
}

function getServer() {
  const server = get$1(KeySettingsServer, null);
  if (server === null) {
    throw new Error('settings.server.missing');
  }
  return server;
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
  switch (res.status) {
    case 200:
    case 403:
    case 400:
      response.status = res.status;
      response.data = data;
      return response;
  }
  throw new Error('Unexpected response from the server');
}


// Look at https://github.com/freshteapot/learnalist-api/blob/master/docs/api.user.login.md
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
  switch (res.status) {
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

/* src/editor/components/error_box.svelte generated by Svelte v3.21.0 */
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

/* src/editor/components/login.svelte generated by Svelte v3.21.0 */
const file$6 = "src/editor/components/login.svelte";

// (35:0) {#if message}
function create_if_block_1(ctx) {
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
		id: create_if_block_1.name,
		type: "if",
		source: "(35:0) {#if message}",
		ctx
	});

	return block;
}

// (99:2) {:else}
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
			add_location(br, file$6, 101, 6, 2927);
			attr_dev(a, "href", "/welcome.html");
			add_location(a, file$6, 103, 6, 2955);
			attr_dev(p, "class", "measure center");
			add_location(p, file$6, 99, 4, 2861);
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
		source: "(99:2) {:else}",
		ctx
	});

	return block;
}

// (39:2) {#if !isLoggedIn}
function create_if_block$2(ctx) {
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
			add_location(label0, file$6, 43, 10, 1080);
			attr_dev(input0, "class", "pa2 input-reset ba bg-transparent b--black-20 w-100 br2");
			attr_dev(input0, "type", "text");
			attr_dev(input0, "name", "username");
			attr_dev(input0, "id", "username");
			attr_dev(input0, "autocapitalize", "none");
			add_location(input0, file$6, 44, 10, 1155);
			attr_dev(div0, "class", "mt3");
			add_location(div0, file$6, 42, 8, 1052);
			attr_dev(label1, "class", "db fw6 lh-copy f6");
			attr_dev(label1, "for", "password");
			add_location(label1, file$6, 53, 10, 1438);
			attr_dev(input1, "class", "b pa2 input-reset ba bg-transparent b--black-20 w-100 br2");
			attr_dev(input1, "type", "password");
			attr_dev(input1, "name", "password");
			attr_dev(input1, "autocomplete", "off");
			attr_dev(input1, "id", "password");
			add_location(input1, file$6, 54, 10, 1513);
			attr_dev(div1, "class", "mv3");
			add_location(div1, file$6, 52, 8, 1410);
			attr_dev(fieldset, "id", "sign_up");
			attr_dev(fieldset, "class", "ba b--transparent ph0 mh0");
			add_location(fieldset, file$6, 41, 6, 986);
			attr_dev(button, "class", "db w-100");
			attr_dev(button, "type", "submit");
			add_location(button, file$6, 68, 14, 1942);
			attr_dev(div2, "class", "flex items-center mb2");
			add_location(div2, file$6, 67, 12, 1892);
			attr_dev(a0, "target", "_blank");
			attr_dev(a0, "href", "https://learnalist.net/api/v1/oauth/google/redirect");
			attr_dev(a0, "class", "f6 link underline dib black");
			add_location(a0, file$6, 74, 16, 2151);
			attr_dev(span0, "class", "f6 link dib black");
			add_location(span0, file$6, 72, 14, 2078);
			attr_dev(div3, "class", "flex items-center mb2");
			add_location(div3, file$6, 71, 12, 2028);
			attr_dev(a1, "target", "_blank");
			attr_dev(a1, "href", "https://learnalist.net/login.html");
			attr_dev(a1, "class", "f6 link underline dib black");
			add_location(a1, file$6, 86, 16, 2542);
			attr_dev(span1, "class", "f6 link dib black");
			add_location(span1, file$6, 84, 14, 2470);
			attr_dev(div4, "class", "flex items-center mb2");
			add_location(div4, file$6, 83, 12, 2420);
			attr_dev(div5, "class", "fr");
			add_location(div5, file$6, 66, 10, 1863);
			attr_dev(div6, "class", "w-100 items-end");
			add_location(div6, file$6, 65, 8, 1823);
			attr_dev(div7, "class", "measure flex");
			add_location(div7, file$6, 64, 6, 1788);
			attr_dev(form, "class", "measure center");
			add_location(form, file$6, 39, 4, 909);
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
		id: create_if_block$2.name,
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
	let if_block0 = /*message*/ ctx[2] && create_if_block_1(ctx);

	function select_block_type(ctx, dirty) {
		if (!/*isLoggedIn*/ ctx[3]) return create_if_block$2;
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
			add_location(main, file$6, 37, 0, 857);
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

					if (dirty & /*message*/ 4) {
						transition_in(if_block0, 1);
					}
				} else {
					if_block0 = create_if_block_1(ctx);
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

/* src/editor/routes/login.svelte generated by Svelte v3.21.0 */
const file$7 = "src/editor/routes/login.svelte";

// (7:0) {#if !$loginHelper.loggedIn}
function create_if_block_1$1(ctx) {
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
		id: create_if_block_1$1.name,
		type: "if",
		source: "(7:0) {#if !$loginHelper.loggedIn}",
		ctx
	});

	return block;
}

// (11:0) {#if $loginHelper.loggedIn}
function create_if_block$3(ctx) {
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
		id: create_if_block$3.name,
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
	let if_block0 = !/*$loginHelper*/ ctx[0].loggedIn && create_if_block_1$1(ctx);
	let if_block1 = /*$loginHelper*/ ctx[0].loggedIn && create_if_block$3(ctx);

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
				if (if_block0) {
					if (dirty & /*$loginHelper*/ 1) {
						transition_in(if_block0, 1);
					}
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

			if (/*$loginHelper*/ ctx[0].loggedIn) {
				if (if_block1) ; else {
					if_block1 = create_if_block$3(ctx);
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

/* src/editor/routes/logout.svelte generated by Svelte v3.21.0 */
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

/* src/editor/routes/not_found.svelte generated by Svelte v3.21.0 */

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

const current = cache$1.get(cache$1.keys["lists.by.me"]);
const { subscribe: subscribe$1, set, update: update$1 } = writable(current);
const loading = writable(false);
const error = writable('');

const ListsByMeStore = () => ({
  subscribe: subscribe$1,
  set,
  loading,
  error,
  async get() {
    let key = cache$1.keys['lists.by.me'];
    let data = [];
    try {
      data = cache$1.get(key, data);
      set(data);
      error.set('');
      if (data.length === 0) {
        loading.set(true);
      }

      const response = await getListsByMe();
      loading.set(false);
      cache$1.save(key, response);
      set(response);
      return response;
    } catch (e) {
      loading.set(false);
      data = cache$1.get(key, data);
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
      cache$1.save(cache$1.keys["lists.by.me"], myLists);
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
      cache$1.save(cache$1.keys["lists.by.me"], updated);
      return updated;
    });
  },

  remove(uuid) {
    update$1(myLists => {
      const found = myLists.filter(aList => aList.uuid !== uuid);
      cache$1.save(cache$1.keys["lists.by.me"], found);
      return found;
    });
  }
});

var myLists = ListsByMeStore();

const current$1 = cache$1.get(cache$1.keys["my.edited.lists"]);
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
        cache$1.save(cache$1.keys["my.edited.lists"], updated);
        return updated;
      });
    },

    remove(uuid) {
      update$2(edits => {
        const found = edits.filter(aList => aList.uuid !== uuid);
        cache$1.save(cache$1.keys["my.edited.lists"], found);
        return found;
      });
    }
});

var listsEdits = ListsEditsStore();

/* src/editor/routes/create_list.svelte generated by Svelte v3.21.0 */
const file$a = "src/editor/routes/create_list.svelte";

function get_each_context(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[9] = list[i];
	child_ctx[11] = i;
	return child_ctx;
}

// (83:2) {#if message}
function create_if_block$4(ctx) {
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
		id: create_if_block$4.name,
		type: "if",
		source: "(83:2) {#if message}",
		ctx
	});

	return block;
}

// (103:10) {#each listTypes as listType, pos}
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
			add_location(input, file$a, 104, 14, 2551);
			attr_dev(label, "for", label_for_value = "list-type-" + /*pos*/ ctx[11]);
			attr_dev(label, "class", "lh-copy");
			add_location(label, file$a, 111, 14, 2772);
			attr_dev(div, "class", "flex items-center mb2");
			add_location(div, file$a, 103, 12, 2501);
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
		source: "(103:10) {#each listTypes as listType, pos}",
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
	let if_block = /*message*/ ctx[2] && create_if_block$4(ctx);
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
			add_location(title_1, file$a, 72, 6, 1507);
			attr_dev(path, "d", "M11 15h2v2h-2v-2zm0-8h2v6h-2V7zm.99-5C6.47 2 2 6.48 2 12s4.47 10 9.99\n        10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42\n        0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z");
			add_location(path, file$a, 73, 6, 1538);
			attr_dev(svg, "class", "w1");
			attr_dev(svg, "data-icon", "info");
			attr_dev(svg, "viewBox", "0 0 24 24");
			set_style(svg, "fill", "currentcolor");
			add_location(svg, file$a, 67, 4, 1397);
			attr_dev(span, "class", "lh-title ml3");
			add_location(span, file$a, 78, 4, 1763);
			attr_dev(div0, "class", "flex items-center justify-center pa1 bg-light-red pv3");
			add_location(div0, file$a, 66, 2, 1325);
			attr_dev(h1, "class", "f4 br3 b--yellow black-70 mv0 pv2 ph4");
			add_location(h1, file$a, 87, 4, 1982);
			attr_dev(input, "class", "input-reset ba b--black-20 pa2 mb2 db w-100");
			attr_dev(input, "type", "text");
			attr_dev(input, "aria-describedby", "title-desc");
			attr_dev(input, "placeholder", "Title");
			add_location(input, file$a, 91, 8, 2160);
			attr_dev(div1, "class", "measure");
			add_location(div1, file$a, 90, 6, 2130);
			attr_dev(fieldset, "class", "bn");
			add_location(fieldset, file$a, 101, 8, 2422);
			attr_dev(div2, "class", "measure");
			add_location(div2, file$a, 100, 6, 2392);
			attr_dev(button, "type", "submit");
			add_location(button, file$a, 119, 8, 2986);
			attr_dev(div3, "class", "measure");
			add_location(div3, file$a, 118, 6, 2956);
			attr_dev(form, "class", "pa4 black-80");
			add_location(form, file$a, 89, 4, 2056);
			attr_dev(section, "class", "center pa3 ph1-ns");
			add_location(section, file$a, 86, 2, 1942);
			attr_dev(div4, "class", "pv0 mw100");
			add_location(div4, file$a, 65, 0, 1299);
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

					if (dirty & /*message*/ 4) {
						transition_in(if_block, 1);
					}
				} else {
					if_block = create_if_block$4(ctx);
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
		cache: cache$1,
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

/* src/editor/routes/create_label.svelte generated by Svelte v3.21.0 */
const file$b = "src/editor/routes/create_label.svelte";

// (46:4) {#if message}
function create_if_block$5(ctx) {
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
		id: create_if_block$5.name,
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
	let if_block = /*message*/ ctx[1] && create_if_block$5(ctx);

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

					if (dirty & /*message*/ 2) {
						transition_in(if_block, 1);
					}
				} else {
					if_block = create_if_block$5(ctx);
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

/* src/editor/routes/create.svelte generated by Svelte v3.21.0 */
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

/* src/editor/components/list_find_item.svelte generated by Svelte v3.21.0 */

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

/* src/editor/routes/list_find.svelte generated by Svelte v3.21.0 */
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
			add_location(button, file$e, 136, 12, 3271);
			attr_dev(div2, "class", "flex items-center mb2");
			add_location(div2, file$e, 135, 10, 3223);
			attr_dev(fieldset, "class", "bn");
			add_location(fieldset, file$e, 112, 8, 2506);
			add_location(div3, file$e, 111, 6, 2492);
			attr_dev(ul, "class", "list pl0 measure center");
			add_location(ul, file$e, 140, 6, 3367);
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
function create_if_block_1$2(ctx) {
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
		id: create_if_block_1$2.name,
		type: "if",
		source: "(109:23) ",
		ctx
	});

	return block;
}

// (107:4) {#if $error}
function create_if_block$6(ctx) {
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
		id: create_if_block$6.name,
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
			add_location(option, file$e, 120, 16, 2769);
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
			add_location(select, file$e, 127, 14, 2992);
			attr_dev(div, "class", "flex items-center mb2");
			add_location(div, file$e, 126, 12, 2942);
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
			add_location(option, file$e, 129, 18, 3090);
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
	const if_block_creators = [create_if_block$6, create_if_block_1$2, create_else_block$2];
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

/* src/editor/components/list.view.data.item.v1.svelte generated by Svelte v3.21.0 */

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

/* src/editor/components/list.view.data.item.v2.svelte generated by Svelte v3.21.0 */

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

/* src/editor/components/list.view.data.item.v3.svelte generated by Svelte v3.21.0 */

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

/* src/editor/components/list.view.data.item.v4.svelte generated by Svelte v3.21.0 */

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

/* src/editor/components/list.view.svelte generated by Svelte v3.21.0 */
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

// (59:4) {#if labels.length > 0}
function create_if_block_1$3(ctx) {
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
			add_location(ul, file$j, 60, 8, 1438);
			attr_dev(div, "class", "nicebox");
			add_location(div, file$j, 59, 6, 1408);
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
		id: create_if_block_1$3.name,
		type: "if",
		source: "(59:4) {#if labels.length > 0}",
		ctx
	});

	return block;
}

// (63:10) {#each labels as item}
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
			add_location(span, file$j, 64, 14, 1549);
			attr_dev(li, "class", "dib mr1 mb2 pl0");
			add_location(li, file$j, 63, 12, 1506);
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
		source: "(63:10) {#each labels as item}",
		ctx
	});

	return block;
}

// (76:4) {#if data.length > 0}
function create_if_block$7(ctx) {
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
			add_location(div, file$j, 76, 6, 1804);
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
		id: create_if_block$7.name,
		type: "if",
		source: "(76:4) {#if data.length > 0}",
		ctx
	});

	return block;
}

// (78:8) {#each data as item}
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
		source: "(78:8) {#each data as item}",
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
	let if_block0 = /*labels*/ ctx[4].length > 0 && create_if_block_1$3(ctx);
	let if_block1 = /*data*/ ctx[0].length > 0 && create_if_block$7(ctx);

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
			add_location(button0, file$j, 48, 6, 1182);
			add_location(button1, file$j, 49, 6, 1231);
			add_location(div0, file$j, 47, 4, 1170);
			add_location(h1, file$j, 53, 6, 1324);
			add_location(p, file$j, 54, 6, 1347);
			add_location(div1, file$j, 52, 4, 1312);
			attr_dev(div2, "class", "pl0 measure center");
			add_location(div2, file$j, 45, 2, 1132);
			attr_dev(div3, "class", "pa3 pa5-ns");
			add_location(div3, file$j, 44, 0, 1105);
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

					if (dirty & /*data*/ 1) {
						transition_in(if_block1, 1);
					}
				} else {
					if_block1 = create_if_block$7(ctx);
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
		const server = getServer();
		window.open(`${server}/alist/${aList.uuid}.html`, "_blank");
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
		getServer,
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

/* src/editor/routes/list_view.svelte generated by Svelte v3.21.0 */

const { console: console_1$2 } = globals;
const file$k = "src/editor/routes/list_view.svelte";

// (15:0) {:else}
function create_else_block$3(ctx) {
	let p;

	const block = {
		c: function create() {
			p = element("p");
			p.textContent = "Not found";
			add_location(p, file$k, 15, 2, 357);
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
		source: "(15:0) {:else}",
		ctx
	});

	return block;
}

// (13:0) {#if show}
function create_if_block$8(ctx) {
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
		id: create_if_block$8.name,
		type: "if",
		source: "(13:0) {#if show}",
		ctx
	});

	return block;
}

function create_fragment$l(ctx) {
	let current_block_type_index;
	let if_block;
	let if_block_anchor;
	let current;
	const if_block_creators = [create_if_block$8, create_else_block$3];
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
		cache: cache$1,
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

/* src/editor/components/list_edit_title.svelte generated by Svelte v3.21.0 */

const file$l = "src/editor/components/list_edit_title.svelte";

function create_fragment$m(ctx) {
	let div;
	let input;
	let dispose;

	const block = {
		c: function create() {
			div = element("div");
			input = element("input");
			attr_dev(input, "placeholder", "Title");
			attr_dev(input, "class", "svelte-1qimety");
			add_location(input, file$l, 17, 2, 759);
			attr_dev(div, "class", "container svelte-1qimety");
			add_location(div, file$l, 16, 0, 733);
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

/* src/editor/components/list_edit_data_v1.svelte generated by Svelte v3.21.0 */
const file$m = "src/editor/components/list_edit_data_v1.svelte";

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
function create_if_block_1$4(ctx) {
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
			add_location(button0, file$m, 164, 4, 4731);
			add_location(button1, file$m, 166, 4, 4774);
			add_location(button2, file$m, 168, 4, 4830);
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
		id: create_if_block_1$4.name,
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
			attr_dev(input, "class", "item item-left svelte-1lkh1k0");
			add_location(input, file$m, 159, 8, 4568);
			attr_dev(button, "class", "item svelte-1lkh1k0");
			add_location(button, file$m, 160, 8, 4633);
			attr_dev(div, "class", "item-container svelte-1lkh1k0");
			add_location(div, file$m, 158, 6, 4531);
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
function create_if_block$9(ctx) {
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

			add_location(button, file$m, 187, 4, 5307);
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
		id: create_if_block$9.name,
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
			attr_dev(input, "class", "item item-left svelte-1lkh1k0");
			input.value = input_value_value = /*listItem*/ ctx[19];
			input.disabled = true;
			add_location(input, file$m, 183, 8, 5216);
			attr_dev(div, "draggable", "true");
			attr_dev(div, "class", "dropzone item-container svelte-1lkh1k0");
			attr_dev(div, "data-index", div_data_index_value = /*pos*/ ctx[21]);
			add_location(div, file$m, 173, 6, 4964);
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
	let if_block0 = !/*enableSortable*/ ctx[2] && create_if_block_1$4(ctx);
	let if_block1 = /*enableSortable*/ ctx[2] && create_if_block$9(ctx);

	const block = {
		c: function create() {
			h1 = element("h1");
			h1.textContent = "Items";
			t1 = space();
			div = element("div");
			if (if_block0) if_block0.c();
			t2 = space();
			if (if_block1) if_block1.c();
			add_location(h1, file$m, 153, 0, 4417);
			add_location(div, file$m, 155, 0, 4433);
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
					if_block0 = create_if_block_1$4(ctx);
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
					if_block1 = create_if_block$9(ctx);
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

/* src/editor/components/list_edit_data_v2.svelte generated by Svelte v3.21.0 */
const file$n = "src/editor/components/list_edit_data_v2.svelte";

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
			add_location(button0, file$n, 180, 4, 5103);
			add_location(button1, file$n, 182, 4, 5146);
			add_location(button2, file$n, 184, 4, 5202);
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
		id: create_if_block_1$5.name,
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
			attr_dev(input0, "class", "item item-left svelte-1lu9t4y");
			add_location(input0, file$n, 163, 10, 4666);
			attr_dev(input1, "placeholder", "to");
			attr_dev(input1, "class", "item item-left svelte-1lu9t4y");
			add_location(input1, file$n, 168, 10, 4803);
			attr_dev(div0, "class", "flex flex-column item-left svelte-1lu9t4y");
			add_location(div0, file$n, 162, 8, 4615);
			attr_dev(button, "class", "item svelte-1lu9t4y");
			add_location(button, file$n, 175, 10, 4990);
			attr_dev(div1, "class", "flex flex-column");
			add_location(div1, file$n, 174, 8, 4949);
			attr_dev(div2, "class", "item-container pv2 bb b--black-05 svelte-1lu9t4y");
			add_location(div2, file$n, 161, 6, 4559);
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
function create_if_block$a(ctx) {
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

			add_location(button, file$n, 214, 4, 5888);
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
		id: create_if_block$a.name,
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
			attr_dev(input0, "class", "item item-left svelte-1lu9t4y");
			input0.value = input0_value_value = /*listItem*/ ctx[21].from;
			input0.disabled = true;
			add_location(input0, file$n, 199, 8, 5588);
			attr_dev(input1, "placeholder", "to");
			attr_dev(input1, "class", "item item-left svelte-1lu9t4y");
			input1.value = input1_value_value = /*listItem*/ ctx[21].to;
			input1.disabled = true;
			add_location(input1, file$n, 205, 8, 5729);
			attr_dev(div, "draggable", "true");
			attr_dev(div, "class", "dropzone item-container svelte-1lu9t4y");
			attr_dev(div, "data-index", div_data_index_value = /*pos*/ ctx[23]);
			add_location(div, file$n, 189, 6, 5336);
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
			add_location(h1, file$n, 156, 0, 4445);
			add_location(div, file$n, 158, 0, 4461);
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

/* src/editor/components/list_edit_data_v3_split.svelte generated by Svelte v3.21.0 */
const file$o = "src/editor/components/list_edit_data_v3_split.svelte";

// (74:4) {:else}
function create_else_block$4(ctx) {
	let span;

	const block = {
		c: function create() {
			span = element("span");
			span.textContent = " ";
			add_location(span, file$o, 74, 6, 1781);
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
function create_if_block$b(ctx) {
	let button;
	let dispose;

	const block = {
		c: function create() {
			button = element("button");
			button.textContent = "x";
			attr_dev(button, "class", "item");
			add_location(button, file$o, 72, 6, 1712);
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
		id: create_if_block$b.name,
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
	let t1;
	let div2;
	let input2;
	let t2;
	let div3;
	let input3;
	let t3;
	let div4;
	let dispose;

	function select_block_type(ctx, dirty) {
		if (/*disabled*/ ctx[1] === undefined) return create_if_block$b;
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
			attr_dev(input0, "class", "w-100 svelte-8nxndc");
			add_location(input0, file$o, 37, 4, 1039);
			attr_dev(div0, "class", "w-25 pa0 mr2");
			add_location(div0, file$o, 36, 2, 1008);
			attr_dev(input1, "placeholder", "distance");
			attr_dev(input1, "type", "number");
			input1.disabled = /*disabled*/ ctx[1];
			attr_dev(input1, "class", "w-100 svelte-8nxndc");
			add_location(input1, file$o, 45, 4, 1189);
			attr_dev(div1, "class", "w-25 pa0 mr2");
			add_location(div1, file$o, 44, 2, 1158);
			attr_dev(input2, "placeholder", "/500m");
			input2.disabled = /*disabled*/ ctx[1];
			attr_dev(input2, "class", "w-100 svelte-8nxndc");
			add_location(input2, file$o, 54, 4, 1367);
			attr_dev(div2, "class", "w-25 pa0 mr2");
			add_location(div2, file$o, 53, 2, 1336);
			attr_dev(input3, "placeholder", "spm");
			attr_dev(input3, "type", "number");
			input3.disabled = /*disabled*/ ctx[1];
			attr_dev(input3, "class", "w-100 svelte-8nxndc");
			add_location(input3, file$o, 62, 4, 1518);
			attr_dev(div3, "class", "w-25 pa0 mr2");
			add_location(div3, file$o, 61, 2, 1487);
			attr_dev(div4, "class", "pa0");
			add_location(div4, file$o, 70, 2, 1655);
			attr_dev(div5, "class", "flex pv0");
			add_location(div5, file$o, 35, 0, 983);
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
				listen_dev(input1, "input", /*input1_input_handler*/ ctx[7]),
				listen_dev(input2, "input", /*input2_input_handler*/ ctx[8]),
				listen_dev(input3, "input", /*input3_input_handler*/ ctx[9])
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

			if (dirty & /*split*/ 1 && to_number(input1.value) !== /*split*/ ctx[0].distance) {
				set_input_value(input1, /*split*/ ctx[0].distance);
			}

			if (dirty & /*disabled*/ 2) {
				prop_dev(input2, "disabled", /*disabled*/ ctx[1]);
			}

			if (dirty & /*split*/ 1 && input2.value !== /*split*/ ctx[0].p500) {
				set_input_value(input2, /*split*/ ctx[0].p500);
			}

			if (dirty & /*disabled*/ 2) {
				prop_dev(input3, "disabled", /*disabled*/ ctx[1]);
			}

			if (dirty & /*split*/ 1 && to_number(input3.value) !== /*split*/ ctx[0].spm) {
				set_input_value(input3, /*split*/ ctx[0].spm);
			}

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

/* src/editor/components/list_edit_data_v3_record.svelte generated by Svelte v3.21.0 */
const file$p = "src/editor/components/list_edit_data_v3_record.svelte";

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
			add_location(span, file$p, 68, 8, 2263);
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
function create_if_block_1$6(ctx) {
	let button;
	let dispose;

	const block = {
		c: function create() {
			button = element("button");
			button.textContent = "x";
			attr_dev(button, "class", "item");
			add_location(button, file$p, 66, 8, 2189);
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
		id: create_if_block_1$6.name,
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
			add_location(div0, file$p, 106, 4, 3135);
			attr_dev(div1, "class", "item-container pv1 svelte-w8u11v");
			add_location(div1, file$p, 105, 2, 3098);
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
function create_if_block$c(ctx) {
	let div;
	let button;
	let dispose;

	const block = {
		c: function create() {
			div = element("div");
			button = element("button");
			button.textContent = "Add Split";
			attr_dev(button, "class", "mr1 ph1");
			add_location(button, file$p, 114, 4, 3338);
			attr_dev(div, "class", "flex pv1");
			add_location(div, file$p, 113, 2, 3311);
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
		id: create_if_block$c.name,
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
		if (/*disabled*/ ctx[1] === undefined) return create_if_block_1$6;
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

	let if_block1 = /*disabled*/ ctx[1] === undefined && create_if_block$c(ctx);

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
			attr_dev(input, "class", "svelte-w8u11v");
			add_location(input, file$p, 61, 6, 2045);
			attr_dev(div0, "class", "pa0 w-100 mr2");
			add_location(div0, file$p, 60, 4, 2011);
			attr_dev(div1, "class", "pa0");
			add_location(div1, file$p, 64, 4, 2128);
			attr_dev(div2, "class", "flex fl w-100");
			add_location(div2, file$p, 59, 2, 1979);
			attr_dev(div3, "class", "item-container pv2 svelte-w8u11v");
			add_location(div3, file$p, 58, 0, 1944);
			add_location(span0, file$p, 78, 8, 2466);
			attr_dev(div4, "class", "w-25 pa1 mr2");
			add_location(div4, file$p, 77, 6, 2431);
			add_location(span1, file$p, 81, 8, 2538);
			attr_dev(div5, "class", "w-25 pa1 mr2");
			add_location(div5, file$p, 80, 6, 2503);
			add_location(span2, file$p, 84, 8, 2612);
			attr_dev(div6, "class", "w-25 pa1 mr2");
			add_location(div6, file$p, 83, 6, 2577);
			add_location(span3, file$p, 88, 8, 2686);
			attr_dev(div7, "class", "w-25 pa1 mr2");
			add_location(div7, file$p, 87, 6, 2651);
			attr_dev(span4, "class", "item pa1");
			add_location(span4, file$p, 91, 8, 2748);
			attr_dev(div8, "class", "pa0");
			add_location(div8, file$p, 90, 6, 2722);
			attr_dev(div9, "class", "flex pv0");
			add_location(div9, file$p, 76, 4, 2402);
			attr_dev(div10, "class", "flex flex-column fl w-100");
			add_location(div10, file$p, 75, 2, 2358);
			attr_dev(div11, "class", "item-container pv2 svelte-w8u11v");
			add_location(div11, file$p, 74, 0, 2323);
			attr_dev(div12, "class", "flex flex-column pv2 fl w-100 bw1 bb bt b--moon-gray");
			add_location(div12, file$p, 98, 2, 2890);
			attr_dev(div13, "class", "item-container pv1 nodrag svelte-w8u11v");
			add_location(div13, file$p, 97, 0, 2848);
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
					if_block1 = create_if_block$c(ctx);
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

/* src/editor/components/list_edit_data_v3.svelte generated by Svelte v3.21.0 */
const file$q = "src/editor/components/list_edit_data_v3.svelte";

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
function create_if_block_1$7(ctx) {
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
			add_location(button0, file$q, 167, 6, 4649);
			attr_dev(button1, "class", "mh1 ph1");
			add_location(button1, file$q, 168, 6, 4709);
			attr_dev(button2, "class", "mh1 ph1");
			add_location(button2, file$q, 169, 6, 4782);
			attr_dev(div, "class", "flex pv1");
			add_location(div, file$q, 166, 4, 4620);
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
		id: create_if_block_1$7.name,
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
function create_if_block$d(ctx) {
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

			add_location(button, file$q, 190, 4, 5296);
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
		id: create_if_block$d.name,
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
			add_location(div, file$q, 176, 6, 4944);
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
	let if_block0 = !/*enableSortable*/ ctx[2] && create_if_block_1$7(ctx);
	let if_block1 = /*enableSortable*/ ctx[2] && create_if_block$d(ctx);

	const block = {
		c: function create() {
			h1 = element("h1");
			h1.textContent = "Items";
			t1 = space();
			div = element("div");
			if (if_block0) if_block0.c();
			t2 = space();
			if (if_block1) if_block1.c();
			add_location(h1, file$q, 158, 0, 4411);
			add_location(div, file$q, 160, 0, 4427);
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

					if (dirty & /*enableSortable*/ 4) {
						transition_in(if_block0, 1);
					}
				} else {
					if_block0 = create_if_block_1$7(ctx);
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

					if (dirty & /*enableSortable*/ 4) {
						transition_in(if_block1, 1);
					}
				} else {
					if_block1 = create_if_block$d(ctx);
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

/* src/editor/components/list_edit_data_v4.svelte generated by Svelte v3.21.0 */

const { console: console_1$3 } = globals;
const file$r = "src/editor/components/list_edit_data_v4.svelte";

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
function create_if_block_1$8(ctx) {
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
			add_location(button0, file$r, 185, 6, 5384);
			attr_dev(button1, "class", "mh1 ph1");
			add_location(button1, file$r, 187, 6, 5445);
			attr_dev(button2, "class", "mh1 ph1");
			add_location(button2, file$r, 189, 6, 5519);
			attr_dev(div, "class", "flex pv1");
			add_location(div, file$r, 184, 4, 5355);
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
		id: create_if_block_1$8.name,
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
			attr_dev(textarea, "class", "item item-left svelte-1d8525f");
			add_location(textarea, file$r, 167, 10, 4893);
			attr_dev(input, "placeholder", "url");
			attr_dev(input, "class", "item item-left mv2 svelte-1d8525f");
			add_location(input, file$r, 172, 10, 5049);
			attr_dev(div0, "class", "flex flex-column item-left svelte-1d8525f");
			add_location(div0, file$r, 166, 8, 4842);
			attr_dev(button, "class", "item svelte-1d8525f");
			add_location(button, file$r, 179, 10, 5242);
			attr_dev(div1, "class", "flex flex-column");
			add_location(div1, file$r, 178, 8, 5201);
			attr_dev(div2, "class", "item-container pv2 bb b--black-05 svelte-1d8525f");
			add_location(div2, file$r, 165, 6, 4786);
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
function create_if_block$e(ctx) {
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

			add_location(button, file$r, 222, 4, 6374);
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
		id: create_if_block$e.name,
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
			attr_dev(textarea, "class", "item item-left svelte-1d8525f");
			textarea.disabled = true;
			add_location(textarea, file$r, 206, 10, 6002);
			attr_dev(input, "placeholder", "url");
			attr_dev(input, "class", "item item-left mv2 svelte-1d8525f");
			input.disabled = true;
			add_location(input, file$r, 212, 10, 6179);
			attr_dev(div0, "class", "flex flex-column item-left svelte-1d8525f");
			add_location(div0, file$r, 205, 8, 5951);
			attr_dev(div1, "draggable", "true");
			attr_dev(div1, "class", "dropzone item-container pv2 bb b--black-05 svelte-1d8525f");
			attr_dev(div1, "data-index", div1_data_index_value = /*pos*/ ctx[25]);
			add_location(div1, file$r, 195, 6, 5680);
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
			add_location(h1, file$r, 160, 0, 4672);
			add_location(div, file$r, 162, 0, 4688);
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
					if_block0 = create_if_block_1$8(ctx);
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
					if_block1 = create_if_block$e(ctx);
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

/* src/editor/components/list_edit_data_todo.svelte generated by Svelte v3.21.0 */

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

/* src/editor/components/list_edit_labels.svelte generated by Svelte v3.21.0 */
const file$t = "src/editor/components/list_edit_labels.svelte";

function get_each_context$8(ctx, list, i) {
	const child_ctx = ctx.slice();
	child_ctx[9] = list[i];
	return child_ctx;
}

// (57:0) {#if !enableSortable}
function create_if_block$f(ctx) {
	let div0;
	let input;
	let t0;
	let button;
	let t2;
	let t3;
	let div1;
	let dispose;
	let if_block = /*labelElement*/ ctx[1] && /*labelElement*/ ctx[1].value !== "" && create_if_block_1$9(ctx);
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
			attr_dev(input, "class", "svelte-1mogb42");
			add_location(input, file$t, 58, 4, 1690);
			add_location(button, file$t, 60, 4, 1752);
			add_location(div0, file$t, 57, 2, 1680);
			attr_dev(div1, "class", "container svelte-1mogb42");
			add_location(div1, file$t, 66, 2, 1909);
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
					if_block = create_if_block_1$9(ctx);
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
		id: create_if_block$f.name,
		type: "if",
		source: "(57:0) {#if !enableSortable}",
		ctx
	});

	return block;
}

// (62:4) {#if labelElement && labelElement.value !== ''}
function create_if_block_1$9(ctx) {
	let button;
	let dispose;

	const block = {
		c: function create() {
			button = element("button");
			button.textContent = "x";
			add_location(button, file$t, 62, 6, 1848);
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
		id: create_if_block_1$9.name,
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
			attr_dev(span, "class", "item svelte-1mogb42");
			add_location(span, file$t, 68, 6, 1967);
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
	let if_block = !/*enableSortable*/ ctx[2] && create_if_block$f(ctx);

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
					if_block = create_if_block$f(ctx);
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

/* src/editor/components/list_edit.svelte generated by Svelte v3.21.0 */

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
			add_location(button0, file$u, 92, 2, 2413);
			add_location(button1, file$u, 93, 2, 2453);
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
			add_location(h2, file$u, 99, 4, 2541);
			attr_dev(input0, "type", "radio");
			input0.__value = "private";
			input0.value = input0.__value;
			/*$$binding_groups*/ ctx[11][1].push(input0);
			add_location(input0, file$u, 101, 6, 2574);
			add_location(label0, file$u, 100, 4, 2560);
			attr_dev(input1, "type", "radio");
			input1.__value = "public";
			input1.value = input1.__value;
			/*$$binding_groups*/ ctx[11][1].push(input1);
			add_location(input1, file$u, 105, 6, 2694);
			add_location(label1, file$u, 104, 4, 2680);
			attr_dev(input2, "type", "radio");
			input2.__value = "friends";
			input2.value = input2.__value;
			/*$$binding_groups*/ ctx[11][1].push(input2);
			add_location(input2, file$u, 109, 6, 2812);
			add_location(label2, file$u, 108, 4, 2798);
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

// (115:2) {#if canInteract}
function create_if_block$g(ctx) {
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
		id: create_if_block$g.name,
		type: "if",
		source: "(115:2) {#if canInteract}",
		ctx
	});

	return block;
}

// (118:6) <Box>
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
			add_location(h3, file$u, 118, 8, 2998);
			attr_dev(input0, "type", "radio");
			input0.__value = "0";
			input0.value = input0.__value;
			/*$$binding_groups*/ ctx[11][0].push(input0);
			add_location(input0, file$u, 120, 10, 3043);
			add_location(label0, file$u, 119, 8, 3025);
			attr_dev(input1, "type", "radio");
			input1.__value = "1";
			input1.value = input1.__value;
			/*$$binding_groups*/ ctx[11][0].push(input1);
			add_location(input1, file$u, 128, 10, 3217);
			add_location(label1, file$u, 127, 8, 3199);
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
		source: "(118:6) <Box>",
		ctx
	});

	return block;
}

// (116:4) <Box>
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
			add_location(h2, file$u, 116, 6, 2960);
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
		source: "(116:4) <Box>",
		ctx
	});

	return block;
}

// (141:4) <Box>
function create_default_slot_2(ctx) {
	let button;
	let dispose;

	const block = {
		c: function create() {
			button = element("button");
			button.textContent = "Delete this list forever";
			add_location(button, file$u, 141, 6, 3440);
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
		source: "(141:4) <Box>",
		ctx
	});

	return block;
}

// (139:2) <Box>
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
			add_location(h1, file$u, 139, 4, 3408);
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
		source: "(139:2) <Box>",
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

	let if_block = /*canInteract*/ ctx[1] && create_if_block$g(ctx);

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
			add_location(h1, file$u, 97, 2, 2511);
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

					if (dirty & /*canInteract*/ 2) {
						transition_in(if_block, 1);
					}
				} else {
					if_block = create_if_block$g(ctx);
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
		cache: cache$1,
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

/* src/editor/routes/list_edit.svelte generated by Svelte v3.21.0 */
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
function create_if_block$h(ctx) {
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
		id: create_if_block$h.name,
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
	const if_block_creators = [create_if_block$h, create_else_block$6];
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
		cache: cache$1,
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

/* src/editor/routes/list_deleted.svelte generated by Svelte v3.21.0 */

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

const { subscribe: subscribe$3, set: set$2, update: update$3 } = writable({
  "gitHash": "na",
  "gitDate": "na",
  "version": "na",
  "url": "https://github.com/freshteapot/learnalist-api"
}
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
    } catch (e) {
      loading$1.set(false);
      set$2({
        "gitHash": "na",
        "gitDate": "na",
        "version": "na",
        "url": "https://github.com/freshteapot/learnalist-api"
      });
      error$1.set(`Error has been occurred. Details: ${e.message}`);
    }
  }
});

var version = VersionStore();

/* src/editor/routes/settings_server_information.svelte generated by Svelte v3.21.0 */
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
function create_if_block_1$a(ctx) {
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
		id: create_if_block_1$a.name,
		type: "if",
		source: "(11:23) ",
		ctx
	});

	return block;
}

// (9:4) {#if $error}
function create_if_block$i(ctx) {
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
		id: create_if_block$i.name,
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
		if (/*$error*/ ctx[0]) return create_if_block$i;
		if (/*$loading*/ ctx[1]) return create_if_block_1$a;
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

/* src/editor/App.svelte generated by Svelte v3.21.0 */

function create_fragment$z(ctx) {
	let t;
	let current;
	const topmenu = new Menu_top({ $$inline: true });
	const router = new Router({ props: { routes }, $$inline: true });
	router.$on("conditionsFailed", conditionsFailed);
	router.$on("routeLoaded", routeLoaded);

	const block = {
		c: function create() {
			create_component(topmenu.$$.fragment);
			t = space();
			create_component(router.$$.fragment);
		},
		l: function claim(nodes) {
			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
		},
		m: function mount(target, anchor) {
			mount_component(topmenu, target, anchor);
			insert_dev(target, t, anchor);
			mount_component(router, target, anchor);
			current = true;
		},
		p: noop,
		i: function intro(local) {
			if (current) return;
			transition_in(topmenu.$$.fragment, local);
			transition_in(router.$$.fragment, local);
			current = true;
		},
		o: function outro(local) {
			transition_out(topmenu.$$.fragment, local);
			transition_out(router.$$.fragment, local);
			current = false;
		},
		d: function destroy(detaching) {
			destroy_component(topmenu, detaching);
			if (detaching) detach_dev(t);
			destroy_component(router, detaching);
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
	cache$1.save(cache$1.keys["last.screen"], "#" + event.detail.location);
}

function instance$z($$self, $$props, $$invalidate) {
	const writable_props = [];

	Object.keys($$props).forEach(key => {
		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<App> was created with unknown prop '${key}'`);
	});

	let { $$slots = {}, $$scope } = $$props;
	validate_slots("App", $$slots, []);

	$$self.$capture_state = () => ({
		cache: cache$1,
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
		init(this, options, instance$z, create_fragment$z, safe_not_equal, {});

		dispatch_dev("SvelteRegisterComponent", {
			component: this,
			tagName: "App",
			options,
			id: create_fragment$z.name
		});
	}
}

if (!localStorage.hasOwnProperty(cache$1.keys['settings.install.defaults'])) {
    cache$1.clear();
    window.location = location.origin + location.pathname;
}

// Specific for the chrome extension
if (window.location.protocol === 'chrome-extension:') {
    let last = cache$1.get(cache$1.keys['last.screen']);
    console.log(last);
    if (last) {
        if (last !== window.location.hash) {
            history.replaceState(undefined, undefined, last);
            window.dispatchEvent(new Event('hashchange'));
        }
    }
}

if (window.location.protocol !== 'chrome-extension:') {
    let last = cache$1.get(cache$1.keys['last.screen']);
    if (last) {
        if (last !== window.location.hash) {
            history.replaceState(undefined, undefined, (location.origin + location.pathname + last));
            window.dispatchEvent(new Event('hashchange'));
        }
    }
}


var app = new App({
    target: document.querySelector("#list-info")
});

export default app;
//# sourceMappingURL=editor.1588413726683.js.map
