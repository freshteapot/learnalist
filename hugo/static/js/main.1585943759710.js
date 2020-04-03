(function (superstore) {
  'use strict';

  const KeySettingsServer = "settings.server";
  const KeySettingsInstallDefaults = "settings.install.defaults";
  const KeyAuthentication = "settings.authentication";
  const KeyUserUuid = "app.user.uuid";
  const KeyNotifications = "app.notifications";


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
    save(KeySettingsInstallDefaults, true);
    save(KeySettingsServer, "https://learnalist.net");
  }

  var cache = {
    KeyAuthentication,
    KeySettingsServer,
    KeySettingsInstallDefaults,
    KeyUserUuid,
    KeyNotifications,
    get,
    save,

    rm,
    clear
  };

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

  /* src/components/banner/banner.svelte generated by Svelte v3.20.1 */

  const { console: console_1 } = globals;
  const file = "src/components/banner/banner.svelte";

  // (193:0) {#if show}
  function create_if_block(ctx) {
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
  			add_location(title, file, 203, 6, 109572);
  			attr_dev(path, "d", path_d_value = /*getIcon*/ ctx[4](/*$notifications*/ ctx[1].level));
  			add_location(path, file, 204, 6, 109603);
  			attr_dev(svg, "class", "w1");
  			attr_dev(svg, "data-icon", "info");
  			attr_dev(svg, "viewBox", "0 0 24 24");
  			set_style(svg, "fill", "currentcolor");
  			set_style(svg, "width", "2em");
  			set_style(svg, "height", "2em");
  			add_location(svg, file, 198, 4, 109441);
  			attr_dev(span, "class", "lh-title ml3");
  			add_location(span, file, 206, 4, 109661);
  			attr_dev(div, "class", "flex items-center justify-center pa3 navy");
  			toggle_class(div, "info", /*level*/ ctx[0] === "info");
  			toggle_class(div, "error", /*level*/ ctx[0] === "error");
  			add_location(div, file, 193, 2, 109284);
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
  		id: create_if_block.name,
  		type: "if",
  		source: "(193:0) {#if show}",
  		ctx
  	});

  	return block;
  }

  function create_fragment(ctx) {
  	let if_block_anchor;
  	let if_block = /*show*/ ctx[3] && create_if_block(ctx);

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
  					if_block = create_if_block(ctx);
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
  		id: create_fragment.name,
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

  function instance($$self, $$props, $$invalidate) {
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
  		init(this, { target: this.shadowRoot }, instance, create_fragment, safe_not_equal, {});

  		if (options) {
  			if (options.target) {
  				insert_dev(options.target, this, options.anchor);
  			}
  		}
  	}
  }

  /* src/components/login_header.svelte generated by Svelte v3.20.1 */
  const file$1 = "src/components/login_header.svelte";

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
  			add_location(a, file$1, 185, 4, 108775);
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
  function create_if_block$1(ctx) {
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
  			add_location(a0, file$1, 166, 4, 108312);
  			attr_dev(a1, "title", "Lists created by you");
  			attr_dev(a1, "href", "/lists-by-me.html");
  			attr_dev(a1, "class", "f6 fw6 hover-blue link black-70 di");
  			add_location(a1, file$1, 172, 4, 108448);
  			attr_dev(a2, "title", "Logout");
  			attr_dev(a2, "href", "/logout.html");
  			attr_dev(a2, "class", "f6 fw6 hover-blue link black-70 di ml3");
  			add_location(a2, file$1, 178, 4, 108595);
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
  		id: create_if_block$1.name,
  		type: "if",
  		source: "(166:2) {#if loggedIn()}",
  		ctx
  	});

  	return block;
  }

  function create_fragment$1(ctx) {
  	let div;
  	let show_if;

  	function select_block_type(ctx, dirty) {
  		if (show_if == null) show_if = !!superstore.loggedIn();
  		if (show_if) return create_if_block$1;
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
  			add_location(div, file$1, 164, 0, 108268);
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
  		id: create_fragment$1.name,
  		type: "component",
  		source: "",
  		ctx
  	});

  	return block;
  }

  function instance$1($$self, $$props, $$invalidate) {
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
  		init(this, { target: this.shadowRoot }, instance$1, create_fragment$1, safe_not_equal, { loginurl: 0 });

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

  function getServer() {
    const server = cache.get(cache.KeySettingsServer, null);
    if (server === null) {
      throw new Error('settings.server.missing');
    }
    return server;
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

  /* src/components/user_login.svelte generated by Svelte v3.20.1 */

  const { console: console_1$1 } = globals;
  const file$2 = "src/components/user_login.svelte";

  // (237:0) {:else}
  function create_else_block(ctx) {
  	let p;
  	let t0;
  	let br;
  	let t1;
  	let a;

  	const block = {
  		c: function create() {
  			p = element("p");
  			t0 = text("You are already logged in.\n    ");
  			br = element("br");
  			t1 = text("\n    Goto the\n    ");
  			a = element("a");
  			a.textContent = "welcome page";
  			add_location(br, file$2, 239, 4, 110563);
  			attr_dev(a, "href", "/welcome.html");
  			add_location(a, file$2, 241, 4, 110587);
  			attr_dev(p, "class", "measure center");
  			add_location(p, file$2, 237, 2, 110501);
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
  		id: create_else_block.name,
  		type: "else",
  		source: "(237:0) {:else}",
  		ctx
  	});

  	return block;
  }

  // (192:0) {#if !isLoggedIn}
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
  	let div6;
  	let div5;
  	let div4;
  	let div2;
  	let button;
  	let t7;
  	let div3;
  	let span;
  	let t8;
  	let a;
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
  			div6 = element("div");
  			div5 = element("div");
  			div4 = element("div");
  			div2 = element("div");
  			button = element("button");
  			button.textContent = "Login";
  			t7 = space();
  			div3 = element("div");
  			span = element("span");
  			t8 = text("or with\n              ");
  			a = element("a");
  			a.textContent = "google";
  			attr_dev(label0, "class", "db fw6 lh-copy f6");
  			attr_dev(label0, "for", "username");
  			add_location(label0, file$2, 195, 8, 109189);
  			attr_dev(input0, "class", "pa2 input-reset ba bg-transparent b--black-20 w-100 br2");
  			attr_dev(input0, "type", "text");
  			attr_dev(input0, "name", "username");
  			attr_dev(input0, "id", "username");
  			attr_dev(input0, "autocapitalize", "none");
  			add_location(input0, file$2, 196, 8, 109262);
  			attr_dev(div0, "class", "mt3");
  			add_location(div0, file$2, 194, 6, 109163);
  			attr_dev(label1, "class", "db fw6 lh-copy f6");
  			attr_dev(label1, "for", "password");
  			add_location(label1, file$2, 205, 8, 109527);
  			attr_dev(input1, "class", "b pa2 input-reset ba bg-transparent b--black-20 w-100 br2");
  			attr_dev(input1, "type", "password");
  			attr_dev(input1, "name", "password");
  			attr_dev(input1, "autocomplete", "off");
  			attr_dev(input1, "id", "password");
  			add_location(input1, file$2, 206, 8, 109600);
  			attr_dev(div1, "class", "mv3");
  			add_location(div1, file$2, 204, 6, 109501);
  			attr_dev(fieldset, "id", "sign_up");
  			attr_dev(fieldset, "class", "ba b--transparent ph0 mh0");
  			add_location(fieldset, file$2, 193, 4, 109099);
  			attr_dev(button, "class", "db w-100");
  			attr_dev(button, "type", "submit");
  			add_location(button, file$2, 219, 12, 110002);
  			attr_dev(div2, "class", "flex items-center mb2");
  			add_location(div2, file$2, 218, 10, 109954);
  			attr_dev(a, "target", "_blank");
  			attr_dev(a, "href", "https://learnalist.net/api/v1/oauth/google/redirect");
  			attr_dev(a, "class", "f6 link underline dib black");
  			add_location(a, file$2, 224, 14, 110200);
  			attr_dev(span, "class", "f6 link dib black");
  			add_location(span, file$2, 222, 12, 110131);
  			attr_dev(div3, "class", "flex items-center mb2");
  			add_location(div3, file$2, 221, 10, 110083);
  			attr_dev(div4, "class", "fr");
  			add_location(div4, file$2, 217, 8, 109927);
  			attr_dev(div5, "class", "w-100 items-end");
  			add_location(div5, file$2, 216, 6, 109889);
  			attr_dev(div6, "class", "measure flex");
  			add_location(div6, file$2, 215, 4, 109856);
  			attr_dev(form, "class", "measure center");
  			add_location(form, file$2, 192, 2, 109025);
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
  			append_dev(form, div6);
  			append_dev(div6, div5);
  			append_dev(div5, div4);
  			append_dev(div4, div2);
  			append_dev(div2, button);
  			append_dev(div4, t7);
  			append_dev(div4, div3);
  			append_dev(div3, span);
  			append_dev(span, t8);
  			append_dev(span, a);
  			if (remount) run_all(dispose);

  			dispose = [
  				listen_dev(input0, "input", /*input0_input_handler*/ ctx[5]),
  				listen_dev(input1, "input", /*input1_input_handler*/ ctx[6]),
  				listen_dev(form, "submit", prevent_default(/*handleSubmit*/ ctx[3]), false, true, false)
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
  		source: "(192:0) {#if !isLoggedIn}",
  		ctx
  	});

  	return block;
  }

  function create_fragment$2(ctx) {
  	let if_block_anchor;

  	function select_block_type(ctx, dirty) {
  		if (!/*isLoggedIn*/ ctx[2]) return create_if_block$2;
  		return create_else_block;
  	}

  	let current_block_type = select_block_type(ctx);
  	let if_block = current_block_type(ctx);

  	const block = {
  		c: function create() {
  			if_block.c();
  			if_block_anchor = empty();
  			this.c = noop;
  		},
  		l: function claim(nodes) {
  			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
  		},
  		m: function mount(target, anchor) {
  			if_block.m(target, anchor);
  			insert_dev(target, if_block_anchor, anchor);
  		},
  		p: function update(ctx, [dirty]) {
  			if_block.p(ctx, dirty);
  		},
  		i: noop,
  		o: noop,
  		d: function destroy(detaching) {
  			if_block.d(detaching);
  			if (detaching) detach_dev(if_block_anchor);
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
  	let isLoggedIn = false;
  	let username = "";
  	let password = "";
  	let message;

  	async function handleSubmit() {
  		if (username === "" || password === "") {
  			message = "Please enter in a username and password";
  			superstore.notify("error", message);
  			return;
  		}

  		let response = await postLogin(username, password);

  		if (response.status != 200) {
  			superstore.notify("error", "Please try again");
  			console.log(response);
  			return;
  		}

  		console.log("TODO, log them in");
  		cache.save(cache.KeyUserUuid, response.body.user_uuid);
  		cache.save(cache.KeyAuthentication, response.body.token);
  		console.log(response);
  		superstore.login(response.body.token, "/");
  		return;
  	}

  	const writable_props = [];

  	Object.keys($$props).forEach(key => {
  		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console_1$1.warn(`<undefined> was created with unknown prop '${key}'`);
  	});

  	let { $$slots = {}, $$scope } = $$props;
  	validate_slots("undefined", $$slots, []);

  	function input0_input_handler() {
  		username = this.value;
  		$$invalidate(0, username);
  	}

  	function input1_input_handler() {
  		password = this.value;
  		$$invalidate(1, password);
  	}

  	$$self.$capture_state = () => ({
  		login: superstore.login,
  		notify: superstore.notify,
  		cache,
  		postLogin,
  		isLoggedIn,
  		username,
  		password,
  		message,
  		handleSubmit
  	});

  	$$self.$inject_state = $$props => {
  		if ("isLoggedIn" in $$props) $$invalidate(2, isLoggedIn = $$props.isLoggedIn);
  		if ("username" in $$props) $$invalidate(0, username = $$props.username);
  		if ("password" in $$props) $$invalidate(1, password = $$props.password);
  		if ("message" in $$props) message = $$props.message;
  	};

  	if ($$props && "$$inject" in $$props) {
  		$$self.$inject_state($$props.$$inject);
  	}

  	return [
  		username,
  		password,
  		isLoggedIn,
  		handleSubmit,
  		message,
  		input0_input_handler,
  		input1_input_handler
  	];
  }

  class User_login extends SvelteElement {
  	constructor(options) {
  		super();
  		this.shadowRoot.innerHTML = `<style>a{background-color:transparent}button,input{font-family:inherit;font-size:100%;line-height:1.15;margin:0}button,input{overflow:visible}button{text-transform:none}button{-webkit-appearance:button}button::-moz-focus-inner{border-style:none;padding:0}button:-moz-focusring{outline:1px dotted ButtonText}fieldset{padding:.35em .75em .625em}a,div,fieldset,form,p{box-sizing:border-box}.ba{border-style:solid;border-width:1px}.b--black-20{border-color:rgba(0,0,0,.2)}.b--transparent{border-color:transparent}.br2{border-radius:.25rem}.db{display:block}.dib{display:inline-block}.flex{display:flex}.items-end{align-items:flex-end}.items-center{align-items:center}.fr{_display:inline}.fr{float:right}.b{font-weight:700}.fw6{font-weight:600}.input-reset{-webkit-appearance:none;-moz-appearance:none}.input-reset::-moz-focus-inner{border:0;padding:0}.lh-copy{line-height:1.5}.link{text-decoration:none}.link,.link:active,.link:focus,.link:hover,.link:link,.link:visited{transition:color .15s ease-in}.link:focus{outline:1px dotted currentColor}.w-100{width:100%}.black{color:#000}.bg-transparent{background-color:transparent}.pa2{padding:.5rem}.ph0{padding-left:0;padding-right:0}.mb2{margin-bottom:.5rem}.mt3{margin-top:1rem}.mv3{margin-top:1rem;margin-bottom:1rem}.mh0{margin-left:0;margin-right:0}.underline{text-decoration:underline}.f6{font-size:.875rem}.measure{max-width:30em}.center{margin-left:auto}.center{margin-right:auto}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){}@media screen and (min-width:60em){}</style>`;
  		init(this, { target: this.shadowRoot }, instance$2, create_fragment$2, safe_not_equal, {});

  		if (options) {
  			if (options.target) {
  				insert_dev(options.target, this, options.anchor);
  			}
  		}
  	}
  }

  const installed = cache.get(cache.KeySettingsInstallDefaults, null);
  if (installed === null) {
      cache.clear();
  }

  // TODO setup
  customElements.define('login-header', Login_header);
  customElements.define('user-login', User_login);
  customElements.define('notification-center', Banner);

}(superstore));
//# sourceMappingURL=main.1585943759710.js.map
