(function (global, factory) {
    typeof exports === 'object' && typeof module !== 'undefined' ? module.exports = factory(require('shared')) :
    typeof define === 'function' && define.amd ? define(['shared'], factory) :
    (global = typeof globalThis !== 'undefined' ? globalThis : global || self, global['toolbox-plank-stats'] = factory(global.shared));
}(this, (function (shared) { 'use strict';

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
    function validate_store(store, name) {
        if (store != null && typeof store.subscribe !== 'function') {
            throw new Error(`'${name}' is not a store with a 'subscribe' method`);
        }
    }
    function subscribe$1(store, ...callbacks) {
        if (store == null) {
            return noop;
        }
        const unsub = store.subscribe(...callbacks);
        return unsub.unsubscribe ? () => unsub.unsubscribe() : unsub;
    }
    function component_subscribe(component, store, callback) {
        component.$$.on_destroy.push(subscribe$1(store, callback));
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
                update$1(component.$$);
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
    function update$1($$) {
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

    const subscriber_queue = [];
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

    async function deleteEntry(uuid) {
        try {
            return await shared.api.deletePlankEntry(uuid);
        } catch (error) {
            console.error("history", error);
            throw (error);
        }
    }

    // if I kept all,
    // then it would be easier to delete
    async function history() {
        try {
            return await shared.api.getPlankHistoryByUser();
        } catch (error) {
            console.error("history", error);
            throw (error);
        }
    }

    // TODO one by one
    // VS all at once and return
    async function saveEntry(entry) {
        try {
            return await shared.api.addPlankEntry(entry);
        } catch (error) {
            console.error("yo2", error);
            throw (error);
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

    function copyObject(item) {
        return JSON.parse(JSON.stringify(item))
    }

    var commonjsGlobal = typeof globalThis !== 'undefined' ? globalThis : typeof window !== 'undefined' ? window : typeof global !== 'undefined' ? global : typeof self !== 'undefined' ? self : {};

    function createCommonjsModule(fn) {
      var module = { exports: {} };
    	return fn(module, module.exports), module.exports;
    }

    var dayjs_min = createCommonjsModule(function (module, exports) {
    !function(t,e){module.exports=e();}(commonjsGlobal,function(){var t="millisecond",e="second",n="minute",r="hour",i="day",s="week",u="month",a="quarter",o="year",f="date",h=/^(\d{4})[-/]?(\d{1,2})?[-/]?(\d{0,2})[^0-9]*(\d{1,2})?:?(\d{1,2})?:?(\d{1,2})?[.:]?(\d+)?$/,c=/\[([^\]]+)]|Y{1,4}|M{1,4}|D{1,2}|d{1,4}|H{1,2}|h{1,2}|a|A|m{1,2}|s{1,2}|Z{1,2}|SSS/g,d={name:"en",weekdays:"Sunday_Monday_Tuesday_Wednesday_Thursday_Friday_Saturday".split("_"),months:"January_February_March_April_May_June_July_August_September_October_November_December".split("_")},$=function(t,e,n){var r=String(t);return !r||r.length>=e?t:""+Array(e+1-r.length).join(n)+t},l={s:$,z:function(t){var e=-t.utcOffset(),n=Math.abs(e),r=Math.floor(n/60),i=n%60;return (e<=0?"+":"-")+$(r,2,"0")+":"+$(i,2,"0")},m:function t(e,n){if(e.date()<n.date())return -t(n,e);var r=12*(n.year()-e.year())+(n.month()-e.month()),i=e.clone().add(r,u),s=n-i<0,a=e.clone().add(r+(s?-1:1),u);return +(-(r+(n-i)/(s?i-a:a-i))||0)},a:function(t){return t<0?Math.ceil(t)||0:Math.floor(t)},p:function(h){return {M:u,y:o,w:s,d:i,D:f,h:r,m:n,s:e,ms:t,Q:a}[h]||String(h||"").toLowerCase().replace(/s$/,"")},u:function(t){return void 0===t}},y="en",M={};M[y]=d;var m=function(t){return t instanceof S},D=function(t,e,n){var r;if(!t)return y;if("string"==typeof t)M[t]&&(r=t),e&&(M[t]=e,r=t);else {var i=t.name;M[i]=t,r=i;}return !n&&r&&(y=r),r||!n&&y},v=function(t,e){if(m(t))return t.clone();var n="object"==typeof e?e:{};return n.date=t,n.args=arguments,new S(n)},g=l;g.l=D,g.i=m,g.w=function(t,e){return v(t,{locale:e.$L,utc:e.$u,x:e.$x,$offset:e.$offset})};var S=function(){function d(t){this.$L=D(t.locale,null,!0),this.parse(t);}var $=d.prototype;return $.parse=function(t){this.$d=function(t){var e=t.date,n=t.utc;if(null===e)return new Date(NaN);if(g.u(e))return new Date;if(e instanceof Date)return new Date(e);if("string"==typeof e&&!/Z$/i.test(e)){var r=e.match(h);if(r){var i=r[2]-1||0,s=(r[7]||"0").substring(0,3);return n?new Date(Date.UTC(r[1],i,r[3]||1,r[4]||0,r[5]||0,r[6]||0,s)):new Date(r[1],i,r[3]||1,r[4]||0,r[5]||0,r[6]||0,s)}}return new Date(e)}(t),this.$x=t.x||{},this.init();},$.init=function(){var t=this.$d;this.$y=t.getFullYear(),this.$M=t.getMonth(),this.$D=t.getDate(),this.$W=t.getDay(),this.$H=t.getHours(),this.$m=t.getMinutes(),this.$s=t.getSeconds(),this.$ms=t.getMilliseconds();},$.$utils=function(){return g},$.isValid=function(){return !("Invalid Date"===this.$d.toString())},$.isSame=function(t,e){var n=v(t);return this.startOf(e)<=n&&n<=this.endOf(e)},$.isAfter=function(t,e){return v(t)<this.startOf(e)},$.isBefore=function(t,e){return this.endOf(e)<v(t)},$.$g=function(t,e,n){return g.u(t)?this[e]:this.set(n,t)},$.unix=function(){return Math.floor(this.valueOf()/1e3)},$.valueOf=function(){return this.$d.getTime()},$.startOf=function(t,a){var h=this,c=!!g.u(a)||a,d=g.p(t),$=function(t,e){var n=g.w(h.$u?Date.UTC(h.$y,e,t):new Date(h.$y,e,t),h);return c?n:n.endOf(i)},l=function(t,e){return g.w(h.toDate()[t].apply(h.toDate("s"),(c?[0,0,0,0]:[23,59,59,999]).slice(e)),h)},y=this.$W,M=this.$M,m=this.$D,D="set"+(this.$u?"UTC":"");switch(d){case o:return c?$(1,0):$(31,11);case u:return c?$(1,M):$(0,M+1);case s:var v=this.$locale().weekStart||0,S=(y<v?y+7:y)-v;return $(c?m-S:m+(6-S),M);case i:case f:return l(D+"Hours",0);case r:return l(D+"Minutes",1);case n:return l(D+"Seconds",2);case e:return l(D+"Milliseconds",3);default:return this.clone()}},$.endOf=function(t){return this.startOf(t,!1)},$.$set=function(s,a){var h,c=g.p(s),d="set"+(this.$u?"UTC":""),$=(h={},h[i]=d+"Date",h[f]=d+"Date",h[u]=d+"Month",h[o]=d+"FullYear",h[r]=d+"Hours",h[n]=d+"Minutes",h[e]=d+"Seconds",h[t]=d+"Milliseconds",h)[c],l=c===i?this.$D+(a-this.$W):a;if(c===u||c===o){var y=this.clone().set(f,1);y.$d[$](l),y.init(),this.$d=y.set(f,Math.min(this.$D,y.daysInMonth())).$d;}else $&&this.$d[$](l);return this.init(),this},$.set=function(t,e){return this.clone().$set(t,e)},$.get=function(t){return this[g.p(t)]()},$.add=function(t,a){var f,h=this;t=Number(t);var c=g.p(a),d=function(e){var n=v(h);return g.w(n.date(n.date()+Math.round(e*t)),h)};if(c===u)return this.set(u,this.$M+t);if(c===o)return this.set(o,this.$y+t);if(c===i)return d(1);if(c===s)return d(7);var $=(f={},f[n]=6e4,f[r]=36e5,f[e]=1e3,f)[c]||1,l=this.$d.getTime()+t*$;return g.w(l,this)},$.subtract=function(t,e){return this.add(-1*t,e)},$.format=function(t){var e=this;if(!this.isValid())return "Invalid Date";var n=t||"YYYY-MM-DDTHH:mm:ssZ",r=g.z(this),i=this.$locale(),s=this.$H,u=this.$m,a=this.$M,o=i.weekdays,f=i.months,h=function(t,r,i,s){return t&&(t[r]||t(e,n))||i[r].substr(0,s)},d=function(t){return g.s(s%12||12,t,"0")},$=i.meridiem||function(t,e,n){var r=t<12?"AM":"PM";return n?r.toLowerCase():r},l={YY:String(this.$y).slice(-2),YYYY:this.$y,M:a+1,MM:g.s(a+1,2,"0"),MMM:h(i.monthsShort,a,f,3),MMMM:h(f,a),D:this.$D,DD:g.s(this.$D,2,"0"),d:String(this.$W),dd:h(i.weekdaysMin,this.$W,o,2),ddd:h(i.weekdaysShort,this.$W,o,3),dddd:o[this.$W],H:String(s),HH:g.s(s,2,"0"),h:d(1),hh:d(2),a:$(s,u,!0),A:$(s,u,!1),m:String(u),mm:g.s(u,2,"0"),s:String(this.$s),ss:g.s(this.$s,2,"0"),SSS:g.s(this.$ms,3,"0"),Z:r};return n.replace(c,function(t,e){return e||l[t]||r.replace(":","")})},$.utcOffset=function(){return 15*-Math.round(this.$d.getTimezoneOffset()/15)},$.diff=function(t,f,h){var c,d=g.p(f),$=v(t),l=6e4*($.utcOffset()-this.utcOffset()),y=this-$,M=g.m(this,$);return M=(c={},c[o]=M/12,c[u]=M,c[a]=M/3,c[s]=(y-l)/6048e5,c[i]=(y-l)/864e5,c[r]=y/36e5,c[n]=y/6e4,c[e]=y/1e3,c)[d]||y,h?M:g.a(M)},$.daysInMonth=function(){return this.endOf(u).$D},$.$locale=function(){return M[this.$L]},$.locale=function(t,e){if(!t)return this.$L;var n=this.clone(),r=D(t,e,!0);return r&&(n.$L=r),n},$.clone=function(){return g.w(this.$d,this)},$.toDate=function(){return new Date(this.valueOf())},$.toJSON=function(){return this.isValid()?this.toISOString():null},$.toISOString=function(){return this.$d.toISOString()},$.toString=function(){return this.$d.toUTCString()},d}(),p=S.prototype;return v.prototype=p,[["$ms",t],["$s",e],["$m",n],["$H",r],["$W",i],["$M",u],["$y",o],["$D",f]].forEach(function(t){p[t[1]]=function(e){return this.$g(e,t[0],t[1])};}),v.extend=function(t,e){return t.$i||(t(e,S,v),t.$i=!0),v},v.locale=D,v.isDayjs=m,v.unix=function(t){return v(1e3*t)},v.en=M[y],v.Ls=M,v.p={},v});
    });

    var isToday = createCommonjsModule(function (module, exports) {
    !function(t,e){module.exports=e();}(commonjsGlobal,function(){return function(t,e,o){e.prototype.isToday=function(){var t=o();return this.format("YYYY-MM-DD")===t.format("YYYY-MM-DD")};}});
    });

    dayjs_min.extend(isToday);

    const StorageKeyPlankSettings = "plank.settings";
    const StorageKeyPlankSavedItems = "plank.saved.items";

    const emptyData = { history: [] };

    let data = copyObject(emptyData);
    const { subscribe, set, update } = writable(data);
    const loading = writable(false);
    const error = writable('');

    const loadHistory = async () => {
      if (!shared.loggedIn()) {
        const tempHistory = get(StorageKeyPlankSavedItems, []);
        data.history = tempHistory.reverse();
        set(data);
        return
      }

      try {
        error.set('');
        loading.set(true);
        const response = await history();

        data.history = response;
        set(data);
        loading.set(false);
      } catch (e) {
        console.log(e);
        data = copyObject(emptyData);
        set(data);
        loading.set(false);
        error.set(`Error has been occurred. Details: ${e.message}`);
      }
    };

    // Remove record
    // Find which day the record is on and remove it
    const deleteRecord = async (entry) => {
      try {
        error.set('');
        loading.set(true);
        await deleteEntry(entry.uuid);
        await loadHistory();
      } catch (e) {
        console.log(e);
        loading.set(false);
        data = copyObject(emptyData);
        set(data);
        error.set(`Error has been occurred. Details: ${e.message}`);
      }
    };

    // If entry is not set we try
    const record = async (entry) => {
      // TODO this will be greatly simplified
      if (entry) {
        // First we save to the temporary area.
        let items = get(StorageKeyPlankSavedItems, []);
        items.push(entry);
        save(StorageKeyPlankSavedItems, items);
      }

      if (!shared.loggedIn()) {
        // Even when not logged in we are building the fake data structures
        await loadHistory();
        return
      }

      const items = get(StorageKeyPlankSavedItems, []);
      if (items.length == 0) {
        return;
      }

      try {
        error.set('');
        loading.set(true);

        await Promise.all(items.map(item => saveEntry(item)));
        save(StorageKeyPlankSavedItems, []);

        await loadHistory();
      } catch (e) {
        console.log(e);
        loading.set(false);
        data = copyObject(emptyData);
        set(data);
        error.set(`Error has been occurred. Details: ${e.message}`);
      }
    };

    const PlankStore = () => ({
      subscribe,
      loading,
      error,

      history() {
        return copyObject(data.history);
      },

      record,
      deleteRecord,
      history: loadHistory,

      settings(input) {
        if (!input) {
          return get(StorageKeyPlankSettings, { showIntervals: false, intervalTime: 15 });
        }
        save(StorageKeyPlankSettings, { showIntervals: input.showIntervals, intervalTime: input.intervalTime });
      }
    });

    var store = PlankStore();

    var humanizeDuration = createCommonjsModule(function (module) {
    // HumanizeDuration.js - https://git.io/j0HgmQ

    /* global define, module */

    (function () {
      // This has to be defined separately because of a bug: we want to alias
      // `gr` and `el` for backwards-compatiblity. In a breaking change, we can
      // remove `gr` entirely.
      // See https://github.com/EvanHahn/HumanizeDuration.js/issues/143 for more.
      var greek = {
        y: function (c) {
          return c === 1 ? "χρόνος" : "χρόνια";
        },
        mo: function (c) {
          return c === 1 ? "μήνας" : "μήνες";
        },
        w: function (c) {
          return c === 1 ? "εβδομάδα" : "εβδομάδες";
        },
        d: function (c) {
          return c === 1 ? "μέρα" : "μέρες";
        },
        h: function (c) {
          return c === 1 ? "ώρα" : "ώρες";
        },
        m: function (c) {
          return c === 1 ? "λεπτό" : "λεπτά";
        },
        s: function (c) {
          return c === 1 ? "δευτερόλεπτο" : "δευτερόλεπτα";
        },
        ms: function (c) {
          return c === 1
            ? "χιλιοστό του δευτερολέπτου"
            : "χιλιοστά του δευτερολέπτου";
        },
        decimal: ",",
      };

      var LANGUAGES = {
        ar: {
          y: function (c) {
            return c === 1 ? "سنة" : "سنوات";
          },
          mo: function (c) {
            return c === 1 ? "شهر" : "أشهر";
          },
          w: function (c) {
            return c === 1 ? "أسبوع" : "أسابيع";
          },
          d: function (c) {
            return c === 1 ? "يوم" : "أيام";
          },
          h: function (c) {
            return c === 1 ? "ساعة" : "ساعات";
          },
          m: function (c) {
            return c > 2 && c < 11 ? "دقائق" : "دقيقة";
          },
          s: function (c) {
            return c === 1 ? "ثانية" : "ثواني";
          },
          ms: function (c) {
            return c === 1 ? "جزء من الثانية" : "أجزاء من الثانية";
          },
          decimal: ",",
        },
        bg: {
          y: function (c) {
            return ["години", "година", "години"][getSlavicForm(c)];
          },
          mo: function (c) {
            return ["месеца", "месец", "месеца"][getSlavicForm(c)];
          },
          w: function (c) {
            return ["седмици", "седмица", "седмици"][getSlavicForm(c)];
          },
          d: function (c) {
            return ["дни", "ден", "дни"][getSlavicForm(c)];
          },
          h: function (c) {
            return ["часа", "час", "часа"][getSlavicForm(c)];
          },
          m: function (c) {
            return ["минути", "минута", "минути"][getSlavicForm(c)];
          },
          s: function (c) {
            return ["секунди", "секунда", "секунди"][getSlavicForm(c)];
          },
          ms: function (c) {
            return ["милисекунди", "милисекунда", "милисекунди"][getSlavicForm(c)];
          },
          decimal: ",",
        },
        ca: {
          y: function (c) {
            return "any" + (c === 1 ? "" : "s");
          },
          mo: function (c) {
            return "mes" + (c === 1 ? "" : "os");
          },
          w: function (c) {
            return "setman" + (c === 1 ? "a" : "es");
          },
          d: function (c) {
            return "di" + (c === 1 ? "a" : "es");
          },
          h: function (c) {
            return "hor" + (c === 1 ? "a" : "es");
          },
          m: function (c) {
            return "minut" + (c === 1 ? "" : "s");
          },
          s: function (c) {
            return "segon" + (c === 1 ? "" : "s");
          },
          ms: function (c) {
            return "milisegon" + (c === 1 ? "" : "s");
          },
          decimal: ",",
        },
        cs: {
          y: function (c) {
            return ["rok", "roku", "roky", "let"][getCzechOrSlovakForm(c)];
          },
          mo: function (c) {
            return ["měsíc", "měsíce", "měsíce", "měsíců"][getCzechOrSlovakForm(c)];
          },
          w: function (c) {
            return ["týden", "týdne", "týdny", "týdnů"][getCzechOrSlovakForm(c)];
          },
          d: function (c) {
            return ["den", "dne", "dny", "dní"][getCzechOrSlovakForm(c)];
          },
          h: function (c) {
            return ["hodina", "hodiny", "hodiny", "hodin"][getCzechOrSlovakForm(c)];
          },
          m: function (c) {
            return ["minuta", "minuty", "minuty", "minut"][getCzechOrSlovakForm(c)];
          },
          s: function (c) {
            return ["sekunda", "sekundy", "sekundy", "sekund"][
              getCzechOrSlovakForm(c)
            ];
          },
          ms: function (c) {
            return ["milisekunda", "milisekundy", "milisekundy", "milisekund"][
              getCzechOrSlovakForm(c)
            ];
          },
          decimal: ",",
        },
        da: {
          y: "år",
          mo: function (c) {
            return "måned" + (c === 1 ? "" : "er");
          },
          w: function (c) {
            return "uge" + (c === 1 ? "" : "r");
          },
          d: function (c) {
            return "dag" + (c === 1 ? "" : "e");
          },
          h: function (c) {
            return "time" + (c === 1 ? "" : "r");
          },
          m: function (c) {
            return "minut" + (c === 1 ? "" : "ter");
          },
          s: function (c) {
            return "sekund" + (c === 1 ? "" : "er");
          },
          ms: function (c) {
            return "millisekund" + (c === 1 ? "" : "er");
          },
          decimal: ",",
        },
        de: {
          y: function (c) {
            return "Jahr" + (c === 1 ? "" : "e");
          },
          mo: function (c) {
            return "Monat" + (c === 1 ? "" : "e");
          },
          w: function (c) {
            return "Woche" + (c === 1 ? "" : "n");
          },
          d: function (c) {
            return "Tag" + (c === 1 ? "" : "e");
          },
          h: function (c) {
            return "Stunde" + (c === 1 ? "" : "n");
          },
          m: function (c) {
            return "Minute" + (c === 1 ? "" : "n");
          },
          s: function (c) {
            return "Sekunde" + (c === 1 ? "" : "n");
          },
          ms: function (c) {
            return "Millisekunde" + (c === 1 ? "" : "n");
          },
          decimal: ",",
        },
        el: greek,
        en: {
          y: function (c) {
            return "year" + (c === 1 ? "" : "s");
          },
          mo: function (c) {
            return "month" + (c === 1 ? "" : "s");
          },
          w: function (c) {
            return "week" + (c === 1 ? "" : "s");
          },
          d: function (c) {
            return "day" + (c === 1 ? "" : "s");
          },
          h: function (c) {
            return "hour" + (c === 1 ? "" : "s");
          },
          m: function (c) {
            return "minute" + (c === 1 ? "" : "s");
          },
          s: function (c) {
            return "second" + (c === 1 ? "" : "s");
          },
          ms: function (c) {
            return "millisecond" + (c === 1 ? "" : "s");
          },
          decimal: ".",
        },
        es: {
          y: function (c) {
            return "año" + (c === 1 ? "" : "s");
          },
          mo: function (c) {
            return "mes" + (c === 1 ? "" : "es");
          },
          w: function (c) {
            return "semana" + (c === 1 ? "" : "s");
          },
          d: function (c) {
            return "día" + (c === 1 ? "" : "s");
          },
          h: function (c) {
            return "hora" + (c === 1 ? "" : "s");
          },
          m: function (c) {
            return "minuto" + (c === 1 ? "" : "s");
          },
          s: function (c) {
            return "segundo" + (c === 1 ? "" : "s");
          },
          ms: function (c) {
            return "milisegundo" + (c === 1 ? "" : "s");
          },
          decimal: ",",
        },
        et: {
          y: function (c) {
            return "aasta" + (c === 1 ? "" : "t");
          },
          mo: function (c) {
            return "kuu" + (c === 1 ? "" : "d");
          },
          w: function (c) {
            return "nädal" + (c === 1 ? "" : "at");
          },
          d: function (c) {
            return "päev" + (c === 1 ? "" : "a");
          },
          h: function (c) {
            return "tund" + (c === 1 ? "" : "i");
          },
          m: function (c) {
            return "minut" + (c === 1 ? "" : "it");
          },
          s: function (c) {
            return "sekund" + (c === 1 ? "" : "it");
          },
          ms: function (c) {
            return "millisekund" + (c === 1 ? "" : "it");
          },
          decimal: ",",
        },
        fa: {
          y: "سال",
          mo: "ماه",
          w: "هفته",
          d: "روز",
          h: "ساعت",
          m: "دقیقه",
          s: "ثانیه",
          ms: "میلی ثانیه",
          decimal: ".",
        },
        fi: {
          y: function (c) {
            return c === 1 ? "vuosi" : "vuotta";
          },
          mo: function (c) {
            return c === 1 ? "kuukausi" : "kuukautta";
          },
          w: function (c) {
            return "viikko" + (c === 1 ? "" : "a");
          },
          d: function (c) {
            return "päivä" + (c === 1 ? "" : "ä");
          },
          h: function (c) {
            return "tunti" + (c === 1 ? "" : "a");
          },
          m: function (c) {
            return "minuutti" + (c === 1 ? "" : "a");
          },
          s: function (c) {
            return "sekunti" + (c === 1 ? "" : "a");
          },
          ms: function (c) {
            return "millisekunti" + (c === 1 ? "" : "a");
          },
          decimal: ",",
        },
        fo: {
          y: "ár",
          mo: function (c) {
            return c === 1 ? "mánaður" : "mánaðir";
          },
          w: function (c) {
            return c === 1 ? "vika" : "vikur";
          },
          d: function (c) {
            return c === 1 ? "dagur" : "dagar";
          },
          h: function (c) {
            return c === 1 ? "tími" : "tímar";
          },
          m: function (c) {
            return c === 1 ? "minuttur" : "minuttir";
          },
          s: "sekund",
          ms: "millisekund",
          decimal: ",",
        },
        fr: {
          y: function (c) {
            return "an" + (c >= 2 ? "s" : "");
          },
          mo: "mois",
          w: function (c) {
            return "semaine" + (c >= 2 ? "s" : "");
          },
          d: function (c) {
            return "jour" + (c >= 2 ? "s" : "");
          },
          h: function (c) {
            return "heure" + (c >= 2 ? "s" : "");
          },
          m: function (c) {
            return "minute" + (c >= 2 ? "s" : "");
          },
          s: function (c) {
            return "seconde" + (c >= 2 ? "s" : "");
          },
          ms: function (c) {
            return "milliseconde" + (c >= 2 ? "s" : "");
          },
          decimal: ",",
        },
        gr: greek,
        he: {
          y: function (c) {
            return c === 1 ? "שנה" : "שנים";
          },
          mo: function (c) {
            return c === 1 ? "חודש" : "חודשים";
          },
          w: function (c) {
            return c === 1 ? "שבוע" : "שבועות";
          },
          d: function (c) {
            return c === 1 ? "יום" : "ימים";
          },
          h: function (c) {
            return c === 1 ? "שעה" : "שעות";
          },
          m: function (c) {
            return c === 1 ? "דקה" : "דקות";
          },
          s: function (c) {
            return c === 1 ? "שניה" : "שניות";
          },
          ms: function (c) {
            return c === 1 ? "מילישנייה" : "מילישניות";
          },
          decimal: ".",
        },
        hr: {
          y: function (c) {
            if (c % 10 === 2 || c % 10 === 3 || c % 10 === 4) {
              return "godine";
            }
            return "godina";
          },
          mo: function (c) {
            if (c === 1) {
              return "mjesec";
            } else if (c === 2 || c === 3 || c === 4) {
              return "mjeseca";
            }
            return "mjeseci";
          },
          w: function (c) {
            if (c % 10 === 1 && c !== 11) {
              return "tjedan";
            }
            return "tjedna";
          },
          d: function (c) {
            return c === 1 ? "dan" : "dana";
          },
          h: function (c) {
            if (c === 1) {
              return "sat";
            } else if (c === 2 || c === 3 || c === 4) {
              return "sata";
            }
            return "sati";
          },
          m: function (c) {
            var mod10 = c % 10;
            if ((mod10 === 2 || mod10 === 3 || mod10 === 4) && (c < 10 || c > 14)) {
              return "minute";
            }
            return "minuta";
          },
          s: function (c) {
            var mod10 = c % 10;
            if (mod10 === 5 || (Math.floor(c) === c && c >= 10 && c <= 19)) {
              return "sekundi";
            } else if (mod10 === 1) {
              return "sekunda";
            } else if (mod10 === 2 || mod10 === 3 || mod10 === 4) {
              return "sekunde";
            }
            return "sekundi";
          },
          ms: function (c) {
            if (c === 1) {
              return "milisekunda";
            } else if (c % 10 === 2 || c % 10 === 3 || c % 10 === 4) {
              return "milisekunde";
            }
            return "milisekundi";
          },
          decimal: ",",
        },
        hi: {
          y: "साल",
          mo: function (c) {
            return c === 1 ? "महीना" : "महीने";
          },
          w: function (c) {
            return c === 1 ? "हफ़्ता" : "हफ्ते";
          },
          d: "दिन",
          h: function (c) {
            return c === 1 ? "घंटा" : "घंटे";
          },
          m: "मिनट",
          s: "सेकंड",
          ms: "मिलीसेकंड",
          decimal: ".",
        },
        hu: {
          y: "év",
          mo: "hónap",
          w: "hét",
          d: "nap",
          h: "óra",
          m: "perc",
          s: "másodperc",
          ms: "ezredmásodperc",
          decimal: ",",
        },
        id: {
          y: "tahun",
          mo: "bulan",
          w: "minggu",
          d: "hari",
          h: "jam",
          m: "menit",
          s: "detik",
          ms: "milidetik",
          decimal: ".",
        },
        is: {
          y: "ár",
          mo: function (c) {
            return "mánuð" + (c === 1 ? "ur" : "ir");
          },
          w: function (c) {
            return "vik" + (c === 1 ? "a" : "ur");
          },
          d: function (c) {
            return "dag" + (c === 1 ? "ur" : "ar");
          },
          h: function (c) {
            return "klukkutím" + (c === 1 ? "i" : "ar");
          },
          m: function (c) {
            return "mínút" + (c === 1 ? "a" : "ur");
          },
          s: function (c) {
            return "sekúnd" + (c === 1 ? "a" : "ur");
          },
          ms: function (c) {
            return "millisekúnd" + (c === 1 ? "a" : "ur");
          },
          decimal: ".",
        },
        it: {
          y: function (c) {
            return "ann" + (c === 1 ? "o" : "i");
          },
          mo: function (c) {
            return "mes" + (c === 1 ? "e" : "i");
          },
          w: function (c) {
            return "settiman" + (c === 1 ? "a" : "e");
          },
          d: function (c) {
            return "giorn" + (c === 1 ? "o" : "i");
          },
          h: function (c) {
            return "or" + (c === 1 ? "a" : "e");
          },
          m: function (c) {
            return "minut" + (c === 1 ? "o" : "i");
          },
          s: function (c) {
            return "second" + (c === 1 ? "o" : "i");
          },
          ms: function (c) {
            return "millisecond" + (c === 1 ? "o" : "i");
          },
          decimal: ",",
        },
        ja: {
          y: "年",
          mo: "月",
          w: "週",
          d: "日",
          h: "時間",
          m: "分",
          s: "秒",
          ms: "ミリ秒",
          decimal: ".",
        },
        ko: {
          y: "년",
          mo: "개월",
          w: "주일",
          d: "일",
          h: "시간",
          m: "분",
          s: "초",
          ms: "밀리 초",
          decimal: ".",
        },
        lo: {
          y: "ປີ",
          mo: "ເດືອນ",
          w: "ອາທິດ",
          d: "ມື້",
          h: "ຊົ່ວໂມງ",
          m: "ນາທີ",
          s: "ວິນາທີ",
          ms: "ມິນລິວິນາທີ",
          decimal: ",",
        },
        lt: {
          y: function (c) {
            return c % 10 === 0 || (c % 100 >= 10 && c % 100 <= 20)
              ? "metų"
              : "metai";
          },
          mo: function (c) {
            return ["mėnuo", "mėnesiai", "mėnesių"][getLithuanianForm(c)];
          },
          w: function (c) {
            return ["savaitė", "savaitės", "savaičių"][getLithuanianForm(c)];
          },
          d: function (c) {
            return ["diena", "dienos", "dienų"][getLithuanianForm(c)];
          },
          h: function (c) {
            return ["valanda", "valandos", "valandų"][getLithuanianForm(c)];
          },
          m: function (c) {
            return ["minutė", "minutės", "minučių"][getLithuanianForm(c)];
          },
          s: function (c) {
            return ["sekundė", "sekundės", "sekundžių"][getLithuanianForm(c)];
          },
          ms: function (c) {
            return ["milisekundė", "milisekundės", "milisekundžių"][
              getLithuanianForm(c)
            ];
          },
          decimal: ",",
        },
        lv: {
          y: function (c) {
            return getLatvianForm(c) ? "gads" : "gadi";
          },
          mo: function (c) {
            return getLatvianForm(c) ? "mēnesis" : "mēneši";
          },
          w: function (c) {
            return getLatvianForm(c) ? "nedēļa" : "nedēļas";
          },
          d: function (c) {
            return getLatvianForm(c) ? "diena" : "dienas";
          },
          h: function (c) {
            return getLatvianForm(c) ? "stunda" : "stundas";
          },
          m: function (c) {
            return getLatvianForm(c) ? "minūte" : "minūtes";
          },
          s: function (c) {
            return getLatvianForm(c) ? "sekunde" : "sekundes";
          },
          ms: function (c) {
            return getLatvianForm(c) ? "milisekunde" : "milisekundes";
          },
          decimal: ",",
        },
        ms: {
          y: "tahun",
          mo: "bulan",
          w: "minggu",
          d: "hari",
          h: "jam",
          m: "minit",
          s: "saat",
          ms: "milisaat",
          decimal: ".",
        },
        nl: {
          y: "jaar",
          mo: function (c) {
            return c === 1 ? "maand" : "maanden";
          },
          w: function (c) {
            return c === 1 ? "week" : "weken";
          },
          d: function (c) {
            return c === 1 ? "dag" : "dagen";
          },
          h: "uur",
          m: function (c) {
            return c === 1 ? "minuut" : "minuten";
          },
          s: function (c) {
            return c === 1 ? "seconde" : "seconden";
          },
          ms: function (c) {
            return c === 1 ? "milliseconde" : "milliseconden";
          },
          decimal: ",",
        },
        no: {
          y: "år",
          mo: function (c) {
            return "måned" + (c === 1 ? "" : "er");
          },
          w: function (c) {
            return "uke" + (c === 1 ? "" : "r");
          },
          d: function (c) {
            return "dag" + (c === 1 ? "" : "er");
          },
          h: function (c) {
            return "time" + (c === 1 ? "" : "r");
          },
          m: function (c) {
            return "minutt" + (c === 1 ? "" : "er");
          },
          s: function (c) {
            return "sekund" + (c === 1 ? "" : "er");
          },
          ms: function (c) {
            return "millisekund" + (c === 1 ? "" : "er");
          },
          decimal: ",",
        },
        pl: {
          y: function (c) {
            return ["rok", "roku", "lata", "lat"][getPolishForm(c)];
          },
          mo: function (c) {
            return ["miesiąc", "miesiąca", "miesiące", "miesięcy"][
              getPolishForm(c)
            ];
          },
          w: function (c) {
            return ["tydzień", "tygodnia", "tygodnie", "tygodni"][getPolishForm(c)];
          },
          d: function (c) {
            return ["dzień", "dnia", "dni", "dni"][getPolishForm(c)];
          },
          h: function (c) {
            return ["godzina", "godziny", "godziny", "godzin"][getPolishForm(c)];
          },
          m: function (c) {
            return ["minuta", "minuty", "minuty", "minut"][getPolishForm(c)];
          },
          s: function (c) {
            return ["sekunda", "sekundy", "sekundy", "sekund"][getPolishForm(c)];
          },
          ms: function (c) {
            return ["milisekunda", "milisekundy", "milisekundy", "milisekund"][
              getPolishForm(c)
            ];
          },
          decimal: ",",
        },
        pt: {
          y: function (c) {
            return "ano" + (c === 1 ? "" : "s");
          },
          mo: function (c) {
            return c === 1 ? "mês" : "meses";
          },
          w: function (c) {
            return "semana" + (c === 1 ? "" : "s");
          },
          d: function (c) {
            return "dia" + (c === 1 ? "" : "s");
          },
          h: function (c) {
            return "hora" + (c === 1 ? "" : "s");
          },
          m: function (c) {
            return "minuto" + (c === 1 ? "" : "s");
          },
          s: function (c) {
            return "segundo" + (c === 1 ? "" : "s");
          },
          ms: function (c) {
            return "milissegundo" + (c === 1 ? "" : "s");
          },
          decimal: ",",
        },
        ro: {
          y: function (c) {
            return c === 1 ? "an" : "ani";
          },
          mo: function (c) {
            return c === 1 ? "lună" : "luni";
          },
          w: function (c) {
            return c === 1 ? "săptămână" : "săptămâni";
          },
          d: function (c) {
            return c === 1 ? "zi" : "zile";
          },
          h: function (c) {
            return c === 1 ? "oră" : "ore";
          },
          m: function (c) {
            return c === 1 ? "minut" : "minute";
          },
          s: function (c) {
            return c === 1 ? "secundă" : "secunde";
          },
          ms: function (c) {
            return c === 1 ? "milisecundă" : "milisecunde";
          },
          decimal: ",",
        },
        ru: {
          y: function (c) {
            return ["лет", "год", "года"][getSlavicForm(c)];
          },
          mo: function (c) {
            return ["месяцев", "месяц", "месяца"][getSlavicForm(c)];
          },
          w: function (c) {
            return ["недель", "неделя", "недели"][getSlavicForm(c)];
          },
          d: function (c) {
            return ["дней", "день", "дня"][getSlavicForm(c)];
          },
          h: function (c) {
            return ["часов", "час", "часа"][getSlavicForm(c)];
          },
          m: function (c) {
            return ["минут", "минута", "минуты"][getSlavicForm(c)];
          },
          s: function (c) {
            return ["секунд", "секунда", "секунды"][getSlavicForm(c)];
          },
          ms: function (c) {
            return ["миллисекунд", "миллисекунда", "миллисекунды"][
              getSlavicForm(c)
            ];
          },
          decimal: ",",
        },
        uk: {
          y: function (c) {
            return ["років", "рік", "роки"][getSlavicForm(c)];
          },
          mo: function (c) {
            return ["місяців", "місяць", "місяці"][getSlavicForm(c)];
          },
          w: function (c) {
            return ["тижнів", "тиждень", "тижні"][getSlavicForm(c)];
          },
          d: function (c) {
            return ["днів", "день", "дні"][getSlavicForm(c)];
          },
          h: function (c) {
            return ["годин", "година", "години"][getSlavicForm(c)];
          },
          m: function (c) {
            return ["хвилин", "хвилина", "хвилини"][getSlavicForm(c)];
          },
          s: function (c) {
            return ["секунд", "секунда", "секунди"][getSlavicForm(c)];
          },
          ms: function (c) {
            return ["мілісекунд", "мілісекунда", "мілісекунди"][getSlavicForm(c)];
          },
          decimal: ",",
        },
        ur: {
          y: "سال",
          mo: function (c) {
            return c === 1 ? "مہینہ" : "مہینے";
          },
          w: function (c) {
            return c === 1 ? "ہفتہ" : "ہفتے";
          },
          d: "دن",
          h: function (c) {
            return c === 1 ? "گھنٹہ" : "گھنٹے";
          },
          m: "منٹ",
          s: "سیکنڈ",
          ms: "ملی سیکنڈ",
          decimal: ".",
        },
        sk: {
          y: function (c) {
            return ["rok", "roky", "roky", "rokov"][getCzechOrSlovakForm(c)];
          },
          mo: function (c) {
            return ["mesiac", "mesiace", "mesiace", "mesiacov"][
              getCzechOrSlovakForm(c)
            ];
          },
          w: function (c) {
            return ["týždeň", "týždne", "týždne", "týždňov"][
              getCzechOrSlovakForm(c)
            ];
          },
          d: function (c) {
            return ["deň", "dni", "dni", "dní"][getCzechOrSlovakForm(c)];
          },
          h: function (c) {
            return ["hodina", "hodiny", "hodiny", "hodín"][getCzechOrSlovakForm(c)];
          },
          m: function (c) {
            return ["minúta", "minúty", "minúty", "minút"][getCzechOrSlovakForm(c)];
          },
          s: function (c) {
            return ["sekunda", "sekundy", "sekundy", "sekúnd"][
              getCzechOrSlovakForm(c)
            ];
          },
          ms: function (c) {
            return ["milisekunda", "milisekundy", "milisekundy", "milisekúnd"][
              getCzechOrSlovakForm(c)
            ];
          },
          decimal: ",",
        },
        sl: {
          y: function (c) {
            if (c % 10 === 1) {
              return "leto";
            } else if (c % 100 === 2) {
              return "leti";
            } else if (
              c % 100 === 3 ||
              c % 100 === 4 ||
              (Math.floor(c) !== c && c % 100 <= 5)
            ) {
              return "leta";
            } else {
              return "let";
            }
          },
          mo: function (c) {
            if (c % 10 === 1) {
              return "mesec";
            } else if (c % 100 === 2 || (Math.floor(c) !== c && c % 100 <= 5)) {
              return "meseca";
            } else if (c % 10 === 3 || c % 10 === 4) {
              return "mesece";
            } else {
              return "mesecev";
            }
          },
          w: function (c) {
            if (c % 10 === 1) {
              return "teden";
            } else if (c % 10 === 2 || (Math.floor(c) !== c && c % 100 <= 4)) {
              return "tedna";
            } else if (c % 10 === 3 || c % 10 === 4) {
              return "tedne";
            } else {
              return "tednov";
            }
          },
          d: function (c) {
            return c % 100 === 1 ? "dan" : "dni";
          },
          h: function (c) {
            if (c % 10 === 1) {
              return "ura";
            } else if (c % 100 === 2) {
              return "uri";
            } else if (c % 10 === 3 || c % 10 === 4 || Math.floor(c) !== c) {
              return "ure";
            } else {
              return "ur";
            }
          },
          m: function (c) {
            if (c % 10 === 1) {
              return "minuta";
            } else if (c % 10 === 2) {
              return "minuti";
            } else if (
              c % 10 === 3 ||
              c % 10 === 4 ||
              (Math.floor(c) !== c && c % 100 <= 4)
            ) {
              return "minute";
            } else {
              return "minut";
            }
          },
          s: function (c) {
            if (c % 10 === 1) {
              return "sekunda";
            } else if (c % 100 === 2) {
              return "sekundi";
            } else if (c % 100 === 3 || c % 100 === 4 || Math.floor(c) !== c) {
              return "sekunde";
            } else {
              return "sekund";
            }
          },
          ms: function (c) {
            if (c % 10 === 1) {
              return "milisekunda";
            } else if (c % 100 === 2) {
              return "milisekundi";
            } else if (c % 100 === 3 || c % 100 === 4 || Math.floor(c) !== c) {
              return "milisekunde";
            } else {
              return "milisekund";
            }
          },
          decimal: ",",
        },
        sv: {
          y: "år",
          mo: function (c) {
            return "månad" + (c === 1 ? "" : "er");
          },
          w: function (c) {
            return "veck" + (c === 1 ? "a" : "or");
          },
          d: function (c) {
            return "dag" + (c === 1 ? "" : "ar");
          },
          h: function (c) {
            return "timm" + (c === 1 ? "e" : "ar");
          },
          m: function (c) {
            return "minut" + (c === 1 ? "" : "er");
          },
          s: function (c) {
            return "sekund" + (c === 1 ? "" : "er");
          },
          ms: function (c) {
            return "millisekund" + (c === 1 ? "" : "er");
          },
          decimal: ",",
        },
        sw: {
          y: function (c) {
            return c === 1 ? "mwaka" : "miaka";
          },
          mo: function (c) {
            return c === 1 ? "mwezi" : "miezi";
          },
          w: "wiki",
          d: function (c) {
            return c === 1 ? "siku" : "masiku";
          },
          h: function (c) {
            return c === 1 ? "saa" : "masaa";
          },
          m: "dakika",
          s: "sekunde",
          ms: "milisekunde",
          decimal: ".",
        },
        tr: {
          y: "yıl",
          mo: "ay",
          w: "hafta",
          d: "gün",
          h: "saat",
          m: "dakika",
          s: "saniye",
          ms: "milisaniye",
          decimal: ",",
        },
        th: {
          y: "ปี",
          mo: "เดือน",
          w: "อาทิตย์",
          d: "วัน",
          h: "ชั่วโมง",
          m: "นาที",
          s: "วินาที",
          ms: "มิลลิวินาที",
          decimal: ".",
        },
        vi: {
          y: "năm",
          mo: "tháng",
          w: "tuần",
          d: "ngày",
          h: "giờ",
          m: "phút",
          s: "giây",
          ms: "mili giây",
          decimal: ",",
        },
        zh_CN: {
          y: "年",
          mo: "个月",
          w: "周",
          d: "天",
          h: "小时",
          m: "分钟",
          s: "秒",
          ms: "毫秒",
          decimal: ".",
        },
        zh_TW: {
          y: "年",
          mo: "個月",
          w: "周",
          d: "天",
          h: "小時",
          m: "分鐘",
          s: "秒",
          ms: "毫秒",
          decimal: ".",
        },
      };

      // You can create a humanizer, which returns a function with default
      // parameters.
      function humanizer(passedOptions) {
        var result = function humanizer(ms, humanizerOptions) {
          var options = assign({}, result, humanizerOptions || {});
          return doHumanization(ms, options);
        };

        return assign(
          result,
          {
            language: "en",
            delimiter: ", ",
            spacer: " ",
            conjunction: "",
            serialComma: true,
            units: ["y", "mo", "w", "d", "h", "m", "s"],
            languages: {},
            round: false,
            unitMeasures: {
              y: 31557600000,
              mo: 2629800000,
              w: 604800000,
              d: 86400000,
              h: 3600000,
              m: 60000,
              s: 1000,
              ms: 1,
            },
          },
          passedOptions
        );
      }

      // The main function is just a wrapper around a default humanizer.
      var humanizeDuration = humanizer({});

      // Build dictionary from options
      function getDictionary(options) {
        var languagesFromOptions = [options.language];

        if (has(options, "fallbacks")) {
          if (isArray(options.fallbacks) && options.fallbacks.length) {
            languagesFromOptions = languagesFromOptions.concat(options.fallbacks);
          } else {
            throw new Error("fallbacks must be an array with at least one element");
          }
        }

        for (var i = 0; i < languagesFromOptions.length; i++) {
          var languageToTry = languagesFromOptions[i];
          if (has(options.languages, languageToTry)) {
            return options.languages[languageToTry];
          } else if (has(LANGUAGES, languageToTry)) {
            return LANGUAGES[languageToTry];
          }
        }

        throw new Error("No language found.");
      }

      // doHumanization does the bulk of the work.
      function doHumanization(ms, options) {
        var i, len, piece;

        // Make sure we have a positive number.
        // Has the nice sideffect of turning Number objects into primitives.
        ms = Math.abs(ms);

        var dictionary = getDictionary(options);
        var pieces = [];

        // Start at the top and keep removing units, bit by bit.
        var unitName, unitMS, unitCount;
        for (i = 0, len = options.units.length; i < len; i++) {
          unitName = options.units[i];
          unitMS = options.unitMeasures[unitName];

          // What's the number of full units we can fit?
          if (i + 1 === len) {
            if (has(options, "maxDecimalPoints")) {
              // We need to use this expValue to avoid rounding functionality of toFixed call
              var expValue = Math.pow(10, options.maxDecimalPoints);
              var unitCountFloat = ms / unitMS;
              unitCount = parseFloat(
                (Math.floor(expValue * unitCountFloat) / expValue).toFixed(
                  options.maxDecimalPoints
                )
              );
            } else {
              unitCount = ms / unitMS;
            }
          } else {
            unitCount = Math.floor(ms / unitMS);
          }

          // Add the string.
          pieces.push({
            unitCount: unitCount,
            unitName: unitName,
          });

          // Remove what we just figured out.
          ms -= unitCount * unitMS;
        }

        var firstOccupiedUnitIndex = 0;
        for (i = 0; i < pieces.length; i++) {
          if (pieces[i].unitCount) {
            firstOccupiedUnitIndex = i;
            break;
          }
        }

        if (options.round) {
          var ratioToLargerUnit, previousPiece;
          for (i = pieces.length - 1; i >= 0; i--) {
            piece = pieces[i];
            piece.unitCount = Math.round(piece.unitCount);

            if (i === 0) {
              break;
            }

            previousPiece = pieces[i - 1];

            ratioToLargerUnit =
              options.unitMeasures[previousPiece.unitName] /
              options.unitMeasures[piece.unitName];
            if (
              piece.unitCount % ratioToLargerUnit === 0 ||
              (options.largest && options.largest - 1 < i - firstOccupiedUnitIndex)
            ) {
              previousPiece.unitCount += piece.unitCount / ratioToLargerUnit;
              piece.unitCount = 0;
            }
          }
        }

        var result = [];
        for (i = 0, pieces.length; i < len; i++) {
          piece = pieces[i];
          if (piece.unitCount) {
            result.push(
              render(piece.unitCount, piece.unitName, dictionary, options)
            );
          }

          if (result.length === options.largest) {
            break;
          }
        }

        if (result.length) {
          if (!options.conjunction || result.length === 1) {
            return result.join(options.delimiter);
          } else if (result.length === 2) {
            return result.join(options.conjunction);
          } else if (result.length > 2) {
            return (
              result.slice(0, -1).join(options.delimiter) +
              (options.serialComma ? "," : "") +
              options.conjunction +
              result.slice(-1)
            );
          }
        } else {
          return render(
            0,
            options.units[options.units.length - 1],
            dictionary,
            options
          );
        }
      }

      function render(count, type, dictionary, options) {
        var decimal;
        if (has(options, "decimal")) {
          decimal = options.decimal;
        } else if (has(dictionary, "decimal")) {
          decimal = dictionary.decimal;
        } else {
          decimal = ".";
        }

        var countStr = count.toString().replace(".", decimal);

        var dictionaryValue = dictionary[type];
        var word;
        if (typeof dictionaryValue === "function") {
          word = dictionaryValue(count);
        } else {
          word = dictionaryValue;
        }

        return countStr + options.spacer + word;
      }

      function assign(destination) {
        var source;
        for (var i = 1; i < arguments.length; i++) {
          source = arguments[i];
          for (var prop in source) {
            if (has(source, prop)) {
              destination[prop] = source[prop];
            }
          }
        }
        return destination;
      }

      // Internal helper function for Polish language.
      function getPolishForm(c) {
        if (c === 1) {
          return 0;
        } else if (Math.floor(c) !== c) {
          return 1;
        } else if (c % 10 >= 2 && c % 10 <= 4 && !(c % 100 > 10 && c % 100 < 20)) {
          return 2;
        } else {
          return 3;
        }
      }

      // Internal helper function for Russian and Ukranian languages.
      function getSlavicForm(c) {
        if (Math.floor(c) !== c) {
          return 2;
        } else if (
          (c % 100 >= 5 && c % 100 <= 20) ||
          (c % 10 >= 5 && c % 10 <= 9) ||
          c % 10 === 0
        ) {
          return 0;
        } else if (c % 10 === 1) {
          return 1;
        } else if (c > 1) {
          return 2;
        } else {
          return 0;
        }
      }

      // Internal helper function for Slovak language.
      function getCzechOrSlovakForm(c) {
        if (c === 1) {
          return 0;
        } else if (Math.floor(c) !== c) {
          return 1;
        } else if (c % 10 >= 2 && c % 10 <= 4 && c % 100 < 10) {
          return 2;
        } else {
          return 3;
        }
      }

      // Internal helper function for Lithuanian language.
      function getLithuanianForm(c) {
        if (c === 1 || (c % 10 === 1 && c % 100 > 20)) {
          return 0;
        } else if (
          Math.floor(c) !== c ||
          (c % 10 >= 2 && c % 100 > 20) ||
          (c % 10 >= 2 && c % 100 < 10)
        ) {
          return 1;
        } else {
          return 2;
        }
      }

      // Internal helper function for Latvian language.
      function getLatvianForm(c) {
        return c % 10 === 1 && c % 100 !== 11;
      }

      // We need to make sure we support browsers that don't have
      // `Array.isArray`, so we define a fallback here.
      var isArray =
        Array.isArray ||
        function (arg) {
          return Object.prototype.toString.call(arg) === "[object Array]";
        };

      function has(obj, key) {
        return Object.prototype.hasOwnProperty.call(obj, key);
      }

      humanizeDuration.getSupportedLanguages = function getSupportedLanguages() {
        var result = [];
        for (var language in LANGUAGES) {
          if (has(LANGUAGES, language) && language !== "gr") {
            result.push(language);
          }
        }
        return result;
      };

      humanizeDuration.humanizer = humanizer;

      if (module.exports) {
        module.exports = humanizeDuration;
      } else {
        this.humanizeDuration = humanizeDuration;
      }
    })();
    });

    var MILLISECONDS_IN_MINUTE$2 = 60000;

    /**
     * Google Chrome as of 67.0.3396.87 introduced timezones with offset that includes seconds.
     * They usually appear for dates that denote time before the timezones were introduced
     * (e.g. for 'Europe/Prague' timezone the offset is GMT+00:57:44 before 1 October 1891
     * and GMT+01:00:00 after that date)
     *
     * Date#getTimezoneOffset returns the offset in minutes and would return 57 for the example above,
     * which would lead to incorrect calculations.
     *
     * This function returns the timezone offset in milliseconds that takes seconds in account.
     */
    var getTimezoneOffsetInMilliseconds = function getTimezoneOffsetInMilliseconds (dirtyDate) {
      var date = new Date(dirtyDate.getTime());
      var baseTimezoneOffset = date.getTimezoneOffset();
      date.setSeconds(0, 0);
      var millisecondsPartOfTimezoneOffset = date.getTime() % MILLISECONDS_IN_MINUTE$2;

      return baseTimezoneOffset * MILLISECONDS_IN_MINUTE$2 + millisecondsPartOfTimezoneOffset
    };

    /**
     * @category Common Helpers
     * @summary Is the given argument an instance of Date?
     *
     * @description
     * Is the given argument an instance of Date?
     *
     * @param {*} argument - the argument to check
     * @returns {Boolean} the given argument is an instance of Date
     *
     * @example
     * // Is 'mayonnaise' a Date?
     * var result = isDate('mayonnaise')
     * //=> false
     */
    function isDate (argument) {
      return argument instanceof Date
    }

    var is_date = isDate;

    var MILLISECONDS_IN_HOUR = 3600000;
    var MILLISECONDS_IN_MINUTE$1 = 60000;
    var DEFAULT_ADDITIONAL_DIGITS = 2;

    var parseTokenDateTimeDelimeter = /[T ]/;
    var parseTokenPlainTime = /:/;

    // year tokens
    var parseTokenYY = /^(\d{2})$/;
    var parseTokensYYY = [
      /^([+-]\d{2})$/, // 0 additional digits
      /^([+-]\d{3})$/, // 1 additional digit
      /^([+-]\d{4})$/ // 2 additional digits
    ];

    var parseTokenYYYY = /^(\d{4})/;
    var parseTokensYYYYY = [
      /^([+-]\d{4})/, // 0 additional digits
      /^([+-]\d{5})/, // 1 additional digit
      /^([+-]\d{6})/ // 2 additional digits
    ];

    // date tokens
    var parseTokenMM = /^-(\d{2})$/;
    var parseTokenDDD = /^-?(\d{3})$/;
    var parseTokenMMDD = /^-?(\d{2})-?(\d{2})$/;
    var parseTokenWww = /^-?W(\d{2})$/;
    var parseTokenWwwD = /^-?W(\d{2})-?(\d{1})$/;

    // time tokens
    var parseTokenHH = /^(\d{2}([.,]\d*)?)$/;
    var parseTokenHHMM = /^(\d{2}):?(\d{2}([.,]\d*)?)$/;
    var parseTokenHHMMSS = /^(\d{2}):?(\d{2}):?(\d{2}([.,]\d*)?)$/;

    // timezone tokens
    var parseTokenTimezone = /([Z+-].*)$/;
    var parseTokenTimezoneZ = /^(Z)$/;
    var parseTokenTimezoneHH = /^([+-])(\d{2})$/;
    var parseTokenTimezoneHHMM = /^([+-])(\d{2}):?(\d{2})$/;

    /**
     * @category Common Helpers
     * @summary Convert the given argument to an instance of Date.
     *
     * @description
     * Convert the given argument to an instance of Date.
     *
     * If the argument is an instance of Date, the function returns its clone.
     *
     * If the argument is a number, it is treated as a timestamp.
     *
     * If an argument is a string, the function tries to parse it.
     * Function accepts complete ISO 8601 formats as well as partial implementations.
     * ISO 8601: http://en.wikipedia.org/wiki/ISO_8601
     *
     * If all above fails, the function passes the given argument to Date constructor.
     *
     * @param {Date|String|Number} argument - the value to convert
     * @param {Object} [options] - the object with options
     * @param {0 | 1 | 2} [options.additionalDigits=2] - the additional number of digits in the extended year format
     * @returns {Date} the parsed date in the local time zone
     *
     * @example
     * // Convert string '2014-02-11T11:30:30' to date:
     * var result = parse('2014-02-11T11:30:30')
     * //=> Tue Feb 11 2014 11:30:30
     *
     * @example
     * // Parse string '+02014101',
     * // if the additional number of digits in the extended year format is 1:
     * var result = parse('+02014101', {additionalDigits: 1})
     * //=> Fri Apr 11 2014 00:00:00
     */
    function parse (argument, dirtyOptions) {
      if (is_date(argument)) {
        // Prevent the date to lose the milliseconds when passed to new Date() in IE10
        return new Date(argument.getTime())
      } else if (typeof argument !== 'string') {
        return new Date(argument)
      }

      var options = dirtyOptions || {};
      var additionalDigits = options.additionalDigits;
      if (additionalDigits == null) {
        additionalDigits = DEFAULT_ADDITIONAL_DIGITS;
      } else {
        additionalDigits = Number(additionalDigits);
      }

      var dateStrings = splitDateString(argument);

      var parseYearResult = parseYear(dateStrings.date, additionalDigits);
      var year = parseYearResult.year;
      var restDateString = parseYearResult.restDateString;

      var date = parseDate(restDateString, year);

      if (date) {
        var timestamp = date.getTime();
        var time = 0;
        var offset;

        if (dateStrings.time) {
          time = parseTime(dateStrings.time);
        }

        if (dateStrings.timezone) {
          offset = parseTimezone(dateStrings.timezone) * MILLISECONDS_IN_MINUTE$1;
        } else {
          var fullTime = timestamp + time;
          var fullTimeDate = new Date(fullTime);

          offset = getTimezoneOffsetInMilliseconds(fullTimeDate);

          // Adjust time when it's coming from DST
          var fullTimeDateNextDay = new Date(fullTime);
          fullTimeDateNextDay.setDate(fullTimeDate.getDate() + 1);
          var offsetDiff =
            getTimezoneOffsetInMilliseconds(fullTimeDateNextDay) -
            getTimezoneOffsetInMilliseconds(fullTimeDate);
          if (offsetDiff > 0) {
            offset += offsetDiff;
          }
        }

        return new Date(timestamp + time + offset)
      } else {
        return new Date(argument)
      }
    }

    function splitDateString (dateString) {
      var dateStrings = {};
      var array = dateString.split(parseTokenDateTimeDelimeter);
      var timeString;

      if (parseTokenPlainTime.test(array[0])) {
        dateStrings.date = null;
        timeString = array[0];
      } else {
        dateStrings.date = array[0];
        timeString = array[1];
      }

      if (timeString) {
        var token = parseTokenTimezone.exec(timeString);
        if (token) {
          dateStrings.time = timeString.replace(token[1], '');
          dateStrings.timezone = token[1];
        } else {
          dateStrings.time = timeString;
        }
      }

      return dateStrings
    }

    function parseYear (dateString, additionalDigits) {
      var parseTokenYYY = parseTokensYYY[additionalDigits];
      var parseTokenYYYYY = parseTokensYYYYY[additionalDigits];

      var token;

      // YYYY or ±YYYYY
      token = parseTokenYYYY.exec(dateString) || parseTokenYYYYY.exec(dateString);
      if (token) {
        var yearString = token[1];
        return {
          year: parseInt(yearString, 10),
          restDateString: dateString.slice(yearString.length)
        }
      }

      // YY or ±YYY
      token = parseTokenYY.exec(dateString) || parseTokenYYY.exec(dateString);
      if (token) {
        var centuryString = token[1];
        return {
          year: parseInt(centuryString, 10) * 100,
          restDateString: dateString.slice(centuryString.length)
        }
      }

      // Invalid ISO-formatted year
      return {
        year: null
      }
    }

    function parseDate (dateString, year) {
      // Invalid ISO-formatted year
      if (year === null) {
        return null
      }

      var token;
      var date;
      var month;
      var week;

      // YYYY
      if (dateString.length === 0) {
        date = new Date(0);
        date.setUTCFullYear(year);
        return date
      }

      // YYYY-MM
      token = parseTokenMM.exec(dateString);
      if (token) {
        date = new Date(0);
        month = parseInt(token[1], 10) - 1;
        date.setUTCFullYear(year, month);
        return date
      }

      // YYYY-DDD or YYYYDDD
      token = parseTokenDDD.exec(dateString);
      if (token) {
        date = new Date(0);
        var dayOfYear = parseInt(token[1], 10);
        date.setUTCFullYear(year, 0, dayOfYear);
        return date
      }

      // YYYY-MM-DD or YYYYMMDD
      token = parseTokenMMDD.exec(dateString);
      if (token) {
        date = new Date(0);
        month = parseInt(token[1], 10) - 1;
        var day = parseInt(token[2], 10);
        date.setUTCFullYear(year, month, day);
        return date
      }

      // YYYY-Www or YYYYWww
      token = parseTokenWww.exec(dateString);
      if (token) {
        week = parseInt(token[1], 10) - 1;
        return dayOfISOYear(year, week)
      }

      // YYYY-Www-D or YYYYWwwD
      token = parseTokenWwwD.exec(dateString);
      if (token) {
        week = parseInt(token[1], 10) - 1;
        var dayOfWeek = parseInt(token[2], 10) - 1;
        return dayOfISOYear(year, week, dayOfWeek)
      }

      // Invalid ISO-formatted date
      return null
    }

    function parseTime (timeString) {
      var token;
      var hours;
      var minutes;

      // hh
      token = parseTokenHH.exec(timeString);
      if (token) {
        hours = parseFloat(token[1].replace(',', '.'));
        return (hours % 24) * MILLISECONDS_IN_HOUR
      }

      // hh:mm or hhmm
      token = parseTokenHHMM.exec(timeString);
      if (token) {
        hours = parseInt(token[1], 10);
        minutes = parseFloat(token[2].replace(',', '.'));
        return (hours % 24) * MILLISECONDS_IN_HOUR +
          minutes * MILLISECONDS_IN_MINUTE$1
      }

      // hh:mm:ss or hhmmss
      token = parseTokenHHMMSS.exec(timeString);
      if (token) {
        hours = parseInt(token[1], 10);
        minutes = parseInt(token[2], 10);
        var seconds = parseFloat(token[3].replace(',', '.'));
        return (hours % 24) * MILLISECONDS_IN_HOUR +
          minutes * MILLISECONDS_IN_MINUTE$1 +
          seconds * 1000
      }

      // Invalid ISO-formatted time
      return null
    }

    function parseTimezone (timezoneString) {
      var token;
      var absoluteOffset;

      // Z
      token = parseTokenTimezoneZ.exec(timezoneString);
      if (token) {
        return 0
      }

      // ±hh
      token = parseTokenTimezoneHH.exec(timezoneString);
      if (token) {
        absoluteOffset = parseInt(token[2], 10) * 60;
        return (token[1] === '+') ? -absoluteOffset : absoluteOffset
      }

      // ±hh:mm or ±hhmm
      token = parseTokenTimezoneHHMM.exec(timezoneString);
      if (token) {
        absoluteOffset = parseInt(token[2], 10) * 60 + parseInt(token[3], 10);
        return (token[1] === '+') ? -absoluteOffset : absoluteOffset
      }

      return 0
    }

    function dayOfISOYear (isoYear, week, day) {
      week = week || 0;
      day = day || 0;
      var date = new Date(0);
      date.setUTCFullYear(isoYear, 0, 4);
      var fourthOfJanuaryDay = date.getUTCDay() || 7;
      var diff = week * 7 + day + 1 - fourthOfJanuaryDay;
      date.setUTCDate(date.getUTCDate() + diff);
      return date
    }

    var parse_1 = parse;

    /**
     * @category Day Helpers
     * @summary Return the start of a day for the given date.
     *
     * @description
     * Return the start of a day for the given date.
     * The result will be in the local timezone.
     *
     * @param {Date|String|Number} date - the original date
     * @returns {Date} the start of a day
     *
     * @example
     * // The start of a day for 2 September 2014 11:55:00:
     * var result = startOfDay(new Date(2014, 8, 2, 11, 55, 0))
     * //=> Tue Sep 02 2014 00:00:00
     */
    function startOfDay (dirtyDate) {
      var date = parse_1(dirtyDate);
      date.setHours(0, 0, 0, 0);
      return date
    }

    var start_of_day = startOfDay;

    var MILLISECONDS_IN_MINUTE = 60000;
    var MILLISECONDS_IN_DAY = 86400000;

    /**
     * @category Day Helpers
     * @summary Get the number of calendar days between the given dates.
     *
     * @description
     * Get the number of calendar days between the given dates.
     *
     * @param {Date|String|Number} dateLeft - the later date
     * @param {Date|String|Number} dateRight - the earlier date
     * @returns {Number} the number of calendar days
     *
     * @example
     * // How many calendar days are between
     * // 2 July 2011 23:00:00 and 2 July 2012 00:00:00?
     * var result = differenceInCalendarDays(
     *   new Date(2012, 6, 2, 0, 0),
     *   new Date(2011, 6, 2, 23, 0)
     * )
     * //=> 366
     */
    function differenceInCalendarDays (dirtyDateLeft, dirtyDateRight) {
      var startOfDayLeft = start_of_day(dirtyDateLeft);
      var startOfDayRight = start_of_day(dirtyDateRight);

      var timestampLeft = startOfDayLeft.getTime() -
        startOfDayLeft.getTimezoneOffset() * MILLISECONDS_IN_MINUTE;
      var timestampRight = startOfDayRight.getTime() -
        startOfDayRight.getTimezoneOffset() * MILLISECONDS_IN_MINUTE;

      // Round the number of days to the nearest integer
      // because the number of milliseconds in a day is not constant
      // (e.g. it's different in the day of the daylight saving time clock shift)
      return Math.round((timestampLeft - timestampRight) / MILLISECONDS_IN_DAY)
    }

    var difference_in_calendar_days = differenceInCalendarDays;

    /**
     * @category Common Helpers
     * @summary Compare the two dates and return -1, 0 or 1.
     *
     * @description
     * Compare the two dates and return 1 if the first date is after the second,
     * -1 if the first date is before the second or 0 if dates are equal.
     *
     * @param {Date|String|Number} dateLeft - the first date to compare
     * @param {Date|String|Number} dateRight - the second date to compare
     * @returns {Number} the result of the comparison
     *
     * @example
     * // Compare 11 February 1987 and 10 July 1989:
     * var result = compareAsc(
     *   new Date(1987, 1, 11),
     *   new Date(1989, 6, 10)
     * )
     * //=> -1
     *
     * @example
     * // Sort the array of dates:
     * var result = [
     *   new Date(1995, 6, 2),
     *   new Date(1987, 1, 11),
     *   new Date(1989, 6, 10)
     * ].sort(compareAsc)
     * //=> [
     * //   Wed Feb 11 1987 00:00:00,
     * //   Mon Jul 10 1989 00:00:00,
     * //   Sun Jul 02 1995 00:00:00
     * // ]
     */
    function compareAsc (dirtyDateLeft, dirtyDateRight) {
      var dateLeft = parse_1(dirtyDateLeft);
      var timeLeft = dateLeft.getTime();
      var dateRight = parse_1(dirtyDateRight);
      var timeRight = dateRight.getTime();

      if (timeLeft < timeRight) {
        return -1
      } else if (timeLeft > timeRight) {
        return 1
      } else {
        return 0
      }
    }

    var compare_asc = compareAsc;

    /**
     * @category Day Helpers
     * @summary Get the number of full days between the given dates.
     *
     * @description
     * Get the number of full days between the given dates.
     *
     * @param {Date|String|Number} dateLeft - the later date
     * @param {Date|String|Number} dateRight - the earlier date
     * @returns {Number} the number of full days
     *
     * @example
     * // How many full days are between
     * // 2 July 2011 23:00:00 and 2 July 2012 00:00:00?
     * var result = differenceInDays(
     *   new Date(2012, 6, 2, 0, 0),
     *   new Date(2011, 6, 2, 23, 0)
     * )
     * //=> 365
     */
    function differenceInDays (dirtyDateLeft, dirtyDateRight) {
      var dateLeft = parse_1(dirtyDateLeft);
      var dateRight = parse_1(dirtyDateRight);

      var sign = compare_asc(dateLeft, dateRight);
      var difference = Math.abs(difference_in_calendar_days(dateLeft, dateRight));
      dateLeft.setDate(dateLeft.getDate() - sign * difference);

      // Math.abs(diff in full days - diff in calendar days) === 1 if last calendar day is not full
      // If so, result must be decreased by 1 in absolute value
      var isLastDayNotFull = compare_asc(dateLeft, dateRight) === -sign;
      return sign * (difference - isLastDayNotFull)
    }

    var difference_in_days = differenceInDays;

    /**
     * @category Day Helpers
     * @summary Add the specified number of days to the given date.
     *
     * @description
     * Add the specified number of days to the given date.
     *
     * @param {Date|String|Number} date - the date to be changed
     * @param {Number} amount - the amount of days to be added
     * @returns {Date} the new date with the days added
     *
     * @example
     * // Add 10 days to 1 September 2014:
     * var result = addDays(new Date(2014, 8, 1), 10)
     * //=> Thu Sep 11 2014 00:00:00
     */
    function addDays (dirtyDate, dirtyAmount) {
      var date = parse_1(dirtyDate);
      var amount = Number(dirtyAmount);
      date.setDate(date.getDate() + amount);
      return date
    }

    var add_days = addDays;

    /**
     * @category Day Helpers
     * @summary Subtract the specified number of days from the given date.
     *
     * @description
     * Subtract the specified number of days from the given date.
     *
     * @param {Date|String|Number} date - the date to be changed
     * @param {Number} amount - the amount of days to be subtracted
     * @returns {Date} the new date with the days subtracted
     *
     * @example
     * // Subtract 10 days from 1 September 2014:
     * var result = subDays(new Date(2014, 8, 1), 10)
     * //=> Fri Aug 22 2014 00:00:00
     */
    function subDays (dirtyDate, dirtyAmount) {
      var amount = Number(dirtyAmount);
      return add_days(dirtyDate, -amount)
    }

    var sub_days = subDays;

    /**
     * @category Common Helpers
     * @summary Is the given date valid?
     *
     * @description
     * Returns false if argument is Invalid Date and true otherwise.
     * Invalid Date is a Date, whose time value is NaN.
     *
     * Time value of Date: http://es5.github.io/#x15.9.1.1
     *
     * @param {Date} date - the date to check
     * @returns {Boolean} the date is valid
     * @throws {TypeError} argument must be an instance of Date
     *
     * @example
     * // For the valid date:
     * var result = isValid(new Date(2014, 1, 31))
     * //=> true
     *
     * @example
     * // For the invalid date:
     * var result = isValid(new Date(''))
     * //=> false
     */
    function isValid (dirtyDate) {
      if (is_date(dirtyDate)) {
        return !isNaN(dirtyDate)
      } else {
        throw new TypeError(toString.call(dirtyDate) + ' is not an instance of Date')
      }
    }

    var is_valid = isValid;

    var helpers = createCommonjsModule(function (module, exports) {

    Object.defineProperty(exports, "__esModule", {
      value: true
    });
    exports.getDatesParameter = exports.sortDates = exports.filterInvalidDates = exports.relativeDates = undefined;



    var _start_of_day2 = _interopRequireDefault(start_of_day);



    var _sub_days2 = _interopRequireDefault(sub_days);



    var _add_days2 = _interopRequireDefault(add_days);



    var _is_valid2 = _interopRequireDefault(is_valid);

    function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

    exports.relativeDates = function relativeDates() {
      return {
        today: (0, _start_of_day2.default)(new Date()),
        yesterday: (0, _start_of_day2.default)((0, _sub_days2.default)(new Date(), 1)),
        tomorrow: (0, _start_of_day2.default)((0, _add_days2.default)(new Date(), 1))
      };
    };

    exports.filterInvalidDates = function filterInvalidDates(dates) {
      return dates.filter(function (date) {
        return !(0, _is_valid2.default)(new Date(date)) ? console.error('The date \'' + date + '\' is not in a valid date format and date-streaks is ignoring it. Browsers do not consistently support this and this package\'s results may fail. Verify the array of dates you\'re passing to date-streaks are all valid date strings. http://momentjs.com/docs/#/parsing/string/') : new Date(date);
      });
    };

    exports.sortDates = function sortDates(dates) {
      return dates.sort(function (a, b) {
        return (0, _start_of_day2.default)(b) - (0, _start_of_day2.default)(a);
      }).reverse();
    };

    exports.getDatesParameter = function getDatesParameter() {
      var param = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};

      if (Array.isArray(param)) {
        return param;
      } else {
        var dates = param.dates;

        return dates || [];
      }
    };
    });

    var _extends$1 = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };



    var _difference_in_days2 = _interopRequireDefault$2(difference_in_days);



    function _interopRequireDefault$2(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

    function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

    function _toConsumableArray$2(arr) { if (Array.isArray(arr)) { for (var i = 0, arr2 = Array(arr.length); i < arr.length; i++) { arr2[i] = arr[i]; } return arr2; } else { return Array.from(arr); } }

    function summary() {
      var datesParam = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : [];

      var dates = (0, helpers.getDatesParameter)(datesParam);

      var _relativeDates = (0, helpers.relativeDates)(),
          today = _relativeDates.today,
          yesterday = _relativeDates.yesterday;

      var allDates = (0, helpers.filterInvalidDates)(dates);
      var sortedDates = (0, helpers.sortDates)(allDates);

      var result = sortedDates.reduce(function (acc, date, index) {
        var first = new Date(date);
        var second = sortedDates[index + 1] ? new Date(sortedDates[index + 1]) : first;
        var diff = (0, _difference_in_days2.default)(second, first);
        var isToday = acc.isToday || (0, _difference_in_days2.default)(date, today) === 0;
        var isYesterday = acc.isYesterday || (0, _difference_in_days2.default)(date, yesterday) === 0;
        var isInFuture = acc.isInFuture || (0, _difference_in_days2.default)(today, date) < 0;

        if (diff === 0) {
          if (isToday) {
            acc.todayInStreak = true;
          }
        } else {
          diff === 1 ? ++acc.streaks[acc.streaks.length - 1] : acc.streaks.push(1);
        }

        return _extends$1({}, acc, {
          longestStreak: Math.max.apply(Math, _toConsumableArray$2(acc.streaks)),
          withinCurrentStreak: acc.isToday || acc.isYesterday || acc.isInFuture || isToday || isYesterday || isInFuture,
          currentStreak: isToday || isYesterday || isInFuture ? acc.streaks[acc.streaks.length - 1] : 0,
          isInFuture: isInFuture,
          isYesterday: isYesterday,
          isToday: isToday
        });
      }, {
        currentStreak: 0,
        longestStreak: 0,
        streaks: [1],
        todayInStreak: false,
        withinCurrentStreak: false,
        isInFuture: false,
        isToday: false,
        isYesterday: false
      });

      result.isToday;
          result.isYesterday;
          result.isInFuture;
          var rest = _objectWithoutProperties(result, ['isToday', 'isYesterday', 'isInFuture']);

      return rest;
    }

    var _default$2 = summary;

    var summary_1 = /*#__PURE__*/Object.defineProperty({
    	default: _default$2
    }, '__esModule', {value: true});

    var _summary3 = _interopRequireDefault$1(summary_1);

    function _interopRequireDefault$1(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

    function _toConsumableArray$1(arr) { if (Array.isArray(arr)) { for (var i = 0, arr2 = Array(arr.length); i < arr.length; i++) { arr2[i] = arr[i]; } return arr2; } else { return Array.from(arr); } }

    function streakRanges() {
      var datesParam = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : [];

      var dates = (0, helpers.getDatesParameter)(datesParam);
      if (dates.length === 0) {
        return [];
      }

      var _summary = (0, _summary3.default)({ dates: dates }),
          _summary$streaks = _summary.streaks,
          streaks = _summary$streaks === undefined ? [] : _summary$streaks;

      var allDates = [].concat(_toConsumableArray$1((0, helpers.sortDates)(dates)));

      return streaks.reduce(function (acc, streak) {
        var start = void 0,
            end = void 0;
        var days = allDates.slice(0, streak);
        allDates.splice(0, streak);

        if (days && days.length > 1) {
          start = new Date(days[0]);
          end = new Date(days[days.length - 1]);
        } else {
          start = new Date(days[0]);
          end = null;
        }

        return [].concat(_toConsumableArray$1(acc), [{
          start: start,
          end: end,
          duration: streak
        }]);
      }, []).reverse();
    }

    var _default$1 = streakRanges;

    var streakRanges_1 = /*#__PURE__*/Object.defineProperty({
    	default: _default$1
    }, '__esModule', {value: true});

    var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };



    var _start_of_day2 = _interopRequireDefault(start_of_day);



    var _sub_days2 = _interopRequireDefault(sub_days);



    function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

    function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

    function _toConsumableArray(arr) { if (Array.isArray(arr)) { for (var i = 0, arr2 = Array(arr.length); i < arr.length; i++) { arr2[i] = arr[i]; } return arr2; } else { return Array.from(arr); } }

    var trackRecord = function trackRecord(_ref) {
      var _ref$dates = _ref.dates,
          dates = _ref$dates === undefined ? [] : _ref$dates,
          _ref$length = _ref.length,
          length = _ref$length === undefined ? 7 : _ref$length,
          _ref$endDate = _ref.endDate,
          endDate = _ref$endDate === undefined ? new Date() : _ref$endDate;

      var pastDates = [].concat(_toConsumableArray(Array(length))).map(function (_, i) {
        return (0, _start_of_day2.default)((0, _sub_days2.default)(endDate, i));
      });
      var sortedDates = (0, helpers.sortDates)(dates).map(function (date) {
        return (0, _start_of_day2.default)(date).getTime();
      });

      var result = pastDates.reduce(function (acc, pastDate) {
        acc = _extends({}, acc, _defineProperty({}, pastDate, sortedDates.includes(pastDate.getTime())));
        return acc;
      }, {});

      return result;
    };

    var _default = trackRecord;

    var trackRecord_1 = /*#__PURE__*/Object.defineProperty({
    	default: _default
    }, '__esModule', {value: true});

    var dist = createCommonjsModule(function (module, exports) {

    Object.defineProperty(exports, "__esModule", {
      value: true
    });
    exports.trackRecord = exports.streakRanges = exports.summary = undefined;



    var _summary2 = _interopRequireDefault(summary_1);



    var _streakRanges2 = _interopRequireDefault(streakRanges_1);



    var _trackRecord2 = _interopRequireDefault(trackRecord_1);

    function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

    exports.summary = _summary2.default;
    exports.streakRanges = _streakRanges2.default;
    exports.trackRecord = _trackRecord2.default;
    });

    var isBetween = createCommonjsModule(function (module, exports) {
    !function(e,t){module.exports=t();}(commonjsGlobal,function(){return function(e,t,i){t.prototype.isBetween=function(e,t,s,f){var n=i(e),o=i(t),r="("===(f=f||"()")[0],u=")"===f[1];return (r?this.isAfter(n,s):!this.isBefore(n,s))&&(u?this.isBefore(o,s):!this.isAfter(o,s))||(r?this.isBefore(n,s):!this.isAfter(n,s))&&(u?this.isAfter(o,s):!this.isBefore(o,s))};}});
    });

    var en = createCommonjsModule(function (module, exports) {
    !function(e,n){module.exports=n();}(commonjsGlobal,function(){return {name:"en",weekdays:"Sunday_Monday_Tuesday_Wednesday_Thursday_Friday_Saturday".split("_"),months:"January_February_March_April_May_June_July_August_September_October_November_December".split("_")}});
    });

    dayjs_min.locale({
        ...en,
        weekStart: 1
    });
    dayjs_min.extend(isBetween);
    dayjs_min.extend(isToday);


    function totals(entries) {
        return entries.reduce((a, b) => a + (b["timerNow"] || 0), 0);
    }

    function todayTotals(entries) {
        return entries.reduce((a, b) => {
            if (!dayjs_min(b.beginningTime).isToday()) {
                return a;
            }
            return a + (b["timerNow"] || 0);
        }, 0);
    }

    function weekTotals(entries) {
        const startOf = dayjs_min().startOf("week");
        const endOf = dayjs_min().endOf("week");

        return entries.reduce((a, b) => {
            const now = dayjs_min(b.beginningTime);
            if (!now.isBetween(startOf, endOf)) {
                return a;
            }
            return a + (b["timerNow"] || 0);
        }, 0);
    }

    function monthTotals(entries) {
        const startOf = dayjs_min().startOf("month");
        const endOf = dayjs_min().endOf("month");

        return entries.reduce((a, b) => {
            const now = dayjs_min(b.beginningTime);
            if (!now.isBetween(startOf, endOf)) {
                return a;
            }
            return a + (b["timerNow"] || 0);
        }, 0);
    }

    /* src/plank/stats/summary.svelte generated by Svelte v3.35.0 */

    const { Object: Object_1 } = globals;
    const file$1 = "src/plank/stats/summary.svelte";

    function get_each_context(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[9] = list[i];
    	return child_ctx;
    }

    function get_each_context_1(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[12] = list[i];
    	return child_ctx;
    }

    // (103:4) {#each stats as stat}
    function create_each_block_1(ctx) {
    	let dl;
    	let dd0;
    	let t0_value = /*stat*/ ctx[12].name + "";
    	let t0;
    	let dd1;
    	let t1_value = /*stat*/ ctx[12].value + "";
    	let t1;

    	const block = {
    		c: function create() {
    			dl = element("dl");
    			dd0 = element("dd");
    			t0 = text(t0_value);
    			dd1 = element("dd");
    			t1 = text(t1_value);
    			attr_dev(dd0, "class", "f6 fw4 ml0 svelte-1vmbms4");
    			add_location(dd0, file$1, 104, 8, 2608);
    			attr_dev(dd1, "class", "f3 fw6 ml0 svelte-1vmbms4");
    			add_location(dd1, file$1, 105, 8, 2656);
    			attr_dev(dl, "class", "dib mr4 svelte-1vmbms4");
    			add_location(dl, file$1, 103, 6, 2579);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, dl, anchor);
    			append_dev(dl, dd0);
    			append_dev(dd0, t0);
    			append_dev(dl, dd1);
    			append_dev(dd1, t1);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*stats*/ 4 && t0_value !== (t0_value = /*stat*/ ctx[12].name + "")) set_data_dev(t0, t0_value);
    			if (dirty & /*stats*/ 4 && t1_value !== (t1_value = /*stat*/ ctx[12].value + "")) set_data_dev(t1, t1_value);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(dl);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block_1.name,
    		type: "each",
    		source: "(103:4) {#each stats as stat}",
    		ctx
    	});

    	return block;
    }

    // (143:6) {#each streakDayRecords as dayRecord}
    function create_each_block(ctx) {
    	let div;

    	const block = {
    		c: function create() {
    			div = element("div");
    			div.textContent = " ";
    			attr_dev(div, "class", "outline pa2 b--black-20 ma1 svelte-1vmbms4");
    			toggle_class(div, "bg-green", /*dayRecord*/ ctx[9].active);
    			toggle_class(div, "red", !/*dayRecord*/ ctx[9].active);
    			toggle_class(div, "bg-red", !/*dayRecord*/ ctx[9].active);
    			add_location(div, file$1, 143, 8, 3776);
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, div, anchor);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*streakDayRecords*/ 2) {
    				toggle_class(div, "bg-green", /*dayRecord*/ ctx[9].active);
    			}

    			if (dirty & /*streakDayRecords*/ 2) {
    				toggle_class(div, "red", !/*dayRecord*/ ctx[9].active);
    			}

    			if (dirty & /*streakDayRecords*/ 2) {
    				toggle_class(div, "bg-red", !/*dayRecord*/ ctx[9].active);
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
    		source: "(143:6) {#each streakDayRecords as dayRecord}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$1(ctx) {
    	let article;
    	let h1;
    	let t1;
    	let dl0;
    	let dd0;
    	let dd1;

    	let t3_value = (/*streakInfo*/ ctx[0].summary.todayInStreak
    	? "Done"
    	: "Not yet") + "";

    	let t3;
    	let t4;
    	let dl1;
    	let dd2;
    	let dd3;
    	let t6_value = humanizeDuration(/*minutes*/ ctx[3].total / /*totalAttempts*/ ctx[4], { largest: 2, maxDecimalPoints: 2 }) + "";
    	let t6;
    	let t7;
    	let h30;
    	let t9;
    	let div0;
    	let t10;
    	let h31;
    	let t12;
    	let div1;
    	let dl2;
    	let dd4;
    	let dd5;
    	let t14_value = /*streakInfo*/ ctx[0].summary.currentStreak + "";
    	let t14;
    	let t15;
    	let dl3;
    	let dd6;
    	let dd7;
    	let t17_value = /*streakInfo*/ ctx[0].summary.longestStreak + "";
    	let t17;
    	let t18;
    	let h32;
    	let t20;
    	let div2;
    	let dl4;
    	let dd8;
    	let dd9;
    	let t22;
    	let t23;
    	let dl5;
    	let dd10;
    	let dd11;
    	let t25_value = /*streakDayRecords*/ ctx[1].length - /*streakDayRecordsDone*/ ctx[5] + "";
    	let t25;
    	let t26;
    	let h33;
    	let t28;
    	let div5;
    	let div4;
    	let t29;
    	let div3;
    	let each_value_1 = /*stats*/ ctx[2];
    	validate_each_argument(each_value_1);
    	let each_blocks_1 = [];

    	for (let i = 0; i < each_value_1.length; i += 1) {
    		each_blocks_1[i] = create_each_block_1(get_each_context_1(ctx, each_value_1, i));
    	}

    	let each_value = /*streakDayRecords*/ ctx[1];
    	validate_each_argument(each_value);
    	let each_blocks = [];

    	for (let i = 0; i < each_value.length; i += 1) {
    		each_blocks[i] = create_each_block(get_each_context(ctx, each_value, i));
    	}

    	const block = {
    		c: function create() {
    			article = element("article");
    			h1 = element("h1");
    			h1.textContent = "Plank Stats";
    			t1 = space();
    			dl0 = element("dl");
    			dd0 = element("dd");
    			dd0.textContent = "Today";
    			dd1 = element("dd");
    			t3 = text(t3_value);
    			t4 = space();
    			dl1 = element("dl");
    			dd2 = element("dd");
    			dd2.textContent = "Average";
    			dd3 = element("dd");
    			t6 = text(t6_value);
    			t7 = space();
    			h30 = element("h3");
    			h30.textContent = "Time Spent";
    			t9 = space();
    			div0 = element("div");

    			for (let i = 0; i < each_blocks_1.length; i += 1) {
    				each_blocks_1[i].c();
    			}

    			t10 = space();
    			h31 = element("h3");
    			h31.textContent = "Streak";
    			t12 = space();
    			div1 = element("div");
    			dl2 = element("dl");
    			dd4 = element("dd");
    			dd4.textContent = "Current";
    			dd5 = element("dd");
    			t14 = text(t14_value);
    			t15 = space();
    			dl3 = element("dl");
    			dd6 = element("dd");
    			dd6.textContent = "Longest";
    			dd7 = element("dd");
    			t17 = text(t17_value);
    			t18 = space();
    			h32 = element("h3");
    			h32.textContent = "Days";
    			t20 = space();
    			div2 = element("div");
    			dl4 = element("dl");
    			dd8 = element("dd");
    			dd8.textContent = "Completed";
    			dd9 = element("dd");
    			t22 = text(/*streakDayRecordsDone*/ ctx[5]);
    			t23 = space();
    			dl5 = element("dl");
    			dd10 = element("dd");
    			dd10.textContent = "Missed";
    			dd11 = element("dd");
    			t25 = text(t25_value);
    			t26 = space();
    			h33 = element("h3");
    			h33.textContent = "Timeline";
    			t28 = space();
    			div5 = element("div");
    			div4 = element("div");

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			t29 = space();
    			div3 = element("div");
    			div3.textContent = "<-Today";
    			attr_dev(h1, "class", "f6 ttu tracked svelte-1vmbms4");
    			add_location(h1, file$1, 81, 2, 1972);
    			attr_dev(dd0, "class", "f6 fw4 ml0 svelte-1vmbms4");
    			add_location(dd0, file$1, 84, 4, 2079);
    			attr_dev(dd1, "class", "f3 fw6 ml0 svelte-1vmbms4");
    			add_location(dd1, file$1, 85, 4, 2117);
    			attr_dev(dl0, "class", "fl fn-l w-50 dib-l w-auto-l lh-title mr5-l svelte-1vmbms4");
    			add_location(dl0, file$1, 83, 2, 2019);
    			attr_dev(dd2, "class", "f6 fw4 ml0 svelte-1vmbms4");
    			add_location(dd2, file$1, 91, 4, 2284);
    			attr_dev(dd3, "class", "f3 fw6 ml0 svelte-1vmbms4");
    			add_location(dd3, file$1, 92, 4, 2324);
    			attr_dev(dl1, "class", "fl fn-l w-50 dib-l w-auto-l lh-title mr5-l svelte-1vmbms4");
    			add_location(dl1, file$1, 90, 2, 2224);
    			attr_dev(h30, "class", "f6 ttu tracked svelte-1vmbms4");
    			add_location(h30, file$1, 100, 2, 2485);
    			attr_dev(div0, "class", "cf svelte-1vmbms4");
    			add_location(div0, file$1, 101, 2, 2530);
    			attr_dev(h31, "class", "f6 ttu tracked svelte-1vmbms4");
    			add_location(h31, file$1, 110, 2, 2733);
    			attr_dev(dd4, "class", "f6 fw4 ml0 svelte-1vmbms4");
    			add_location(dd4, file$1, 113, 6, 2857);
    			attr_dev(dd5, "class", "f3 fw6 ml0 svelte-1vmbms4");
    			add_location(dd5, file$1, 114, 6, 2899);
    			attr_dev(dl2, "class", "fl fn-l w-50 dib-l w-auto-l lh-title mr5-l svelte-1vmbms4");
    			add_location(dl2, file$1, 112, 4, 2795);
    			attr_dev(dd6, "class", "f6 fw4 ml0 svelte-1vmbms4");
    			add_location(dd6, file$1, 117, 6, 3038);
    			attr_dev(dd7, "class", "f3 fw6 ml0 svelte-1vmbms4");
    			add_location(dd7, file$1, 118, 6, 3080);
    			attr_dev(dl3, "class", "fl fn-l w-50 dib-l w-auto-l lh-title mr5-l svelte-1vmbms4");
    			add_location(dl3, file$1, 116, 4, 2976);
    			attr_dev(div1, "class", "cf svelte-1vmbms4");
    			add_location(div1, file$1, 111, 2, 2774);
    			attr_dev(h32, "class", "f6 ttu tracked svelte-1vmbms4");
    			add_location(h32, file$1, 122, 2, 3165);
    			attr_dev(dd8, "class", "f6 fw4 ml0 svelte-1vmbms4");
    			add_location(dd8, file$1, 125, 6, 3287);
    			attr_dev(dd9, "class", "f3 fw6 ml0 svelte-1vmbms4");
    			add_location(dd9, file$1, 126, 6, 3331);
    			attr_dev(dl4, "class", "fl fn-l w-50 dib-l w-auto-l lh-title mr5-l svelte-1vmbms4");
    			add_location(dl4, file$1, 124, 4, 3225);
    			attr_dev(dd10, "class", "f6 fw4 ml0 svelte-1vmbms4");
    			add_location(dd10, file$1, 132, 6, 3475);
    			attr_dev(dd11, "class", "f3 fw6 ml0 svelte-1vmbms4");
    			add_location(dd11, file$1, 133, 6, 3516);
    			attr_dev(dl5, "class", "fl fn-l w-50 dib-l w-auto-l lh-title mr5-l svelte-1vmbms4");
    			add_location(dl5, file$1, 131, 4, 3413);
    			attr_dev(div2, "class", "cf svelte-1vmbms4");
    			add_location(div2, file$1, 123, 2, 3204);
    			attr_dev(h33, "class", "f6 ttu tracked svelte-1vmbms4");
    			add_location(h33, file$1, 139, 2, 3631);
    			attr_dev(div3, "class", "outline pa2 b--black-20 ma1 svelte-1vmbms4");
    			add_location(div3, file$1, 152, 6, 4016);
    			attr_dev(div4, "class", "flex flex-wrap svelte-1vmbms4");
    			add_location(div4, file$1, 141, 4, 3695);
    			attr_dev(div5, "class", "cf svelte-1vmbms4");
    			add_location(div5, file$1, 140, 2, 3674);
    			attr_dev(article, "class", "pa3 pa5-ns svelte-1vmbms4");
    			attr_dev(article, "data-name", "slab-stat-small");
    			add_location(article, file$1, 80, 0, 1913);
    		},
    		l: function claim(nodes) {
    			throw new Error("options.hydrate only works if the component was compiled with the `hydratable: true` option");
    		},
    		m: function mount(target, anchor) {
    			insert_dev(target, article, anchor);
    			append_dev(article, h1);
    			append_dev(article, t1);
    			append_dev(article, dl0);
    			append_dev(dl0, dd0);
    			append_dev(dl0, dd1);
    			append_dev(dd1, t3);
    			append_dev(article, t4);
    			append_dev(article, dl1);
    			append_dev(dl1, dd2);
    			append_dev(dl1, dd3);
    			append_dev(dd3, t6);
    			append_dev(article, t7);
    			append_dev(article, h30);
    			append_dev(article, t9);
    			append_dev(article, div0);

    			for (let i = 0; i < each_blocks_1.length; i += 1) {
    				each_blocks_1[i].m(div0, null);
    			}

    			append_dev(article, t10);
    			append_dev(article, h31);
    			append_dev(article, t12);
    			append_dev(article, div1);
    			append_dev(div1, dl2);
    			append_dev(dl2, dd4);
    			append_dev(dl2, dd5);
    			append_dev(dd5, t14);
    			append_dev(div1, t15);
    			append_dev(div1, dl3);
    			append_dev(dl3, dd6);
    			append_dev(dl3, dd7);
    			append_dev(dd7, t17);
    			append_dev(article, t18);
    			append_dev(article, h32);
    			append_dev(article, t20);
    			append_dev(article, div2);
    			append_dev(div2, dl4);
    			append_dev(dl4, dd8);
    			append_dev(dl4, dd9);
    			append_dev(dd9, t22);
    			append_dev(div2, t23);
    			append_dev(div2, dl5);
    			append_dev(dl5, dd10);
    			append_dev(dl5, dd11);
    			append_dev(dd11, t25);
    			append_dev(article, t26);
    			append_dev(article, h33);
    			append_dev(article, t28);
    			append_dev(article, div5);
    			append_dev(div5, div4);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(div4, null);
    			}

    			append_dev(div4, t29);
    			append_dev(div4, div3);
    		},
    		p: function update(ctx, [dirty]) {
    			if (dirty & /*streakInfo*/ 1 && t3_value !== (t3_value = (/*streakInfo*/ ctx[0].summary.todayInStreak
    			? "Done"
    			: "Not yet") + "")) set_data_dev(t3, t3_value);

    			if (dirty & /*minutes, totalAttempts*/ 24 && t6_value !== (t6_value = humanizeDuration(/*minutes*/ ctx[3].total / /*totalAttempts*/ ctx[4], { largest: 2, maxDecimalPoints: 2 }) + "")) set_data_dev(t6, t6_value);

    			if (dirty & /*stats*/ 4) {
    				each_value_1 = /*stats*/ ctx[2];
    				validate_each_argument(each_value_1);
    				let i;

    				for (i = 0; i < each_value_1.length; i += 1) {
    					const child_ctx = get_each_context_1(ctx, each_value_1, i);

    					if (each_blocks_1[i]) {
    						each_blocks_1[i].p(child_ctx, dirty);
    					} else {
    						each_blocks_1[i] = create_each_block_1(child_ctx);
    						each_blocks_1[i].c();
    						each_blocks_1[i].m(div0, null);
    					}
    				}

    				for (; i < each_blocks_1.length; i += 1) {
    					each_blocks_1[i].d(1);
    				}

    				each_blocks_1.length = each_value_1.length;
    			}

    			if (dirty & /*streakInfo*/ 1 && t14_value !== (t14_value = /*streakInfo*/ ctx[0].summary.currentStreak + "")) set_data_dev(t14, t14_value);
    			if (dirty & /*streakInfo*/ 1 && t17_value !== (t17_value = /*streakInfo*/ ctx[0].summary.longestStreak + "")) set_data_dev(t17, t17_value);
    			if (dirty & /*streakDayRecordsDone*/ 32) set_data_dev(t22, /*streakDayRecordsDone*/ ctx[5]);
    			if (dirty & /*streakDayRecords, streakDayRecordsDone*/ 34 && t25_value !== (t25_value = /*streakDayRecords*/ ctx[1].length - /*streakDayRecordsDone*/ ctx[5] + "")) set_data_dev(t25, t25_value);

    			if (dirty & /*streakDayRecords*/ 2) {
    				each_value = /*streakDayRecords*/ ctx[1];
    				validate_each_argument(each_value);
    				let i;

    				for (i = 0; i < each_value.length; i += 1) {
    					const child_ctx = get_each_context(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(div4, t29);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value.length;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(article);
    			destroy_each(each_blocks_1, detaching);
    			destroy_each(each_blocks, detaching);
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
    	let stats;
    	let streakInfo;
    	let minutes;
    	let totalAttempts;
    	let streakDayRecords;
    	let streakDayRecordsDone;
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots("Summary", slots, []);
    	let { history = [] } = $$props;

    	const orderDates = dates => {
    		dates = Object.entries(dates).sort(function (a, b) {
    			return b[0] - a[0];
    		}).reverse();

    		return dates.map(d => {
    			return {
    				day: new Date(Date.parse(d[0])),
    				active: d[1]
    			};
    		});
    	};

    	function info(history) {
    		let records = history.map(entry => {
    			// format(new Date(entry.beginningTime), "YYYY-MM-DD");
    			// https://stackoverflow.com/a/38148759
    			return new Date(entry.beginningTime).toLocaleDateString("en-CA");
    		});

    		records = [...new Set(records)];

    		return {
    			summary: dist.summary(records),
    			ranges: dist.streakRanges(records),
    			records: orderDates(dist.trackRecord({ dates: records, length: records.length }))
    		};
    	}

    	const writable_props = ["history"];

    	Object_1.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<Summary> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ("history" in $$props) $$invalidate(6, history = $$props.history);
    	};

    	$$self.$capture_state = () => ({
    		humanizeDuration,
    		summary: dist.summary,
    		streakRanges: dist.streakRanges,
    		trackRecord: dist.trackRecord,
    		monthTotals,
    		todayTotals,
    		totals,
    		weekTotals,
    		history,
    		orderDates,
    		info,
    		stats,
    		streakInfo,
    		minutes,
    		totalAttempts,
    		streakDayRecords,
    		streakDayRecordsDone
    	});

    	$$self.$inject_state = $$props => {
    		if ("history" in $$props) $$invalidate(6, history = $$props.history);
    		if ("stats" in $$props) $$invalidate(2, stats = $$props.stats);
    		if ("streakInfo" in $$props) $$invalidate(0, streakInfo = $$props.streakInfo);
    		if ("minutes" in $$props) $$invalidate(3, minutes = $$props.minutes);
    		if ("totalAttempts" in $$props) $$invalidate(4, totalAttempts = $$props.totalAttempts);
    		if ("streakDayRecords" in $$props) $$invalidate(1, streakDayRecords = $$props.streakDayRecords);
    		if ("streakDayRecordsDone" in $$props) $$invalidate(5, streakDayRecordsDone = $$props.streakDayRecordsDone);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*history*/ 64) {
    			$$invalidate(2, stats = [
    				{
    					name: "Today",
    					value: humanizeDuration(todayTotals(history), { largest: 2, maxDecimalPoints: 2 })
    				},
    				{
    					name: "Week",
    					value: humanizeDuration(weekTotals(history), { largest: 2, maxDecimalPoints: 2 })
    				},
    				{
    					name: "Month",
    					value: humanizeDuration(monthTotals(history), { largest: 2, maxDecimalPoints: 2 })
    				},
    				{
    					name: "Overall",
    					value: humanizeDuration(totals(history), { largest: 2, maxDecimalPoints: 2 })
    				}
    			]);
    		}

    		if ($$self.$$.dirty & /*history*/ 64) {
    			$$invalidate(0, streakInfo = info(history));
    		}

    		if ($$self.$$.dirty & /*history*/ 64) {
    			$$invalidate(3, minutes = {
    				total: history.map(e => e.timerNow).reduce((a, b) => a + b, 0)
    			});
    		}

    		if ($$self.$$.dirty & /*history*/ 64) {
    			$$invalidate(4, totalAttempts = history.length);
    		}

    		if ($$self.$$.dirty & /*streakInfo*/ 1) {
    			$$invalidate(1, streakDayRecords = streakInfo.records);
    		}

    		if ($$self.$$.dirty & /*streakDayRecords*/ 2) {
    			$$invalidate(5, streakDayRecordsDone = streakDayRecords.filter(e => e.active).length);
    		}
    	};

    	return [
    		streakInfo,
    		streakDayRecords,
    		stats,
    		minutes,
    		totalAttempts,
    		streakDayRecordsDone,
    		history
    	];
    }

    class Summary extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$1, create_fragment$1, safe_not_equal, { history: 6 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Summary",
    			options,
    			id: create_fragment$1.name
    		});
    	}

    	get history() {
    		throw new Error("<Summary>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set history(value) {
    		throw new Error("<Summary>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src/plank/stats/v1.svelte generated by Svelte v3.35.0 */
    const file = "src/plank/stats/v1.svelte";

    // (34:0) {#if !isLoggedIn}
    function create_if_block_1(ctx) {
    	let p;

    	const block = {
    		c: function create() {
    			p = element("p");
    			p.textContent = "Please login first";
    			add_location(p, file, 34, 2, 653);
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
    		id: create_if_block_1.name,
    		type: "if",
    		source: "(34:0) {#if !isLoggedIn}",
    		ctx
    	});

    	return block;
    }

    // (38:0) {#if isLoggedIn}
    function create_if_block(ctx) {
    	let summary;
    	let current;

    	summary = new Summary({
    			props: { history: /*history*/ ctx[1] },
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			create_component(summary.$$.fragment);
    		},
    		m: function mount(target, anchor) {
    			mount_component(summary, target, anchor);
    			current = true;
    		},
    		p: function update(ctx, dirty) {
    			const summary_changes = {};
    			if (dirty & /*history*/ 2) summary_changes.history = /*history*/ ctx[1];
    			summary.$set(summary_changes);
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(summary.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(summary.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			destroy_component(summary, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block.name,
    		type: "if",
    		source: "(38:0) {#if isLoggedIn}",
    		ctx
    	});

    	return block;
    }

    function create_fragment(ctx) {
    	let t;
    	let if_block1_anchor;
    	let current;
    	let mounted;
    	let dispose;
    	let if_block0 = !/*isLoggedIn*/ ctx[0] && create_if_block_1(ctx);
    	let if_block1 = /*isLoggedIn*/ ctx[0] && create_if_block(ctx);

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

    			if (!mounted) {
    				dispose = listen_dev(window, "beforeunload", /*beforeunload*/ ctx[2], false, false, false);
    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (!/*isLoggedIn*/ ctx[0]) {
    				if (if_block0) ; else {
    					if_block0 = create_if_block_1(ctx);
    					if_block0.c();
    					if_block0.m(t.parentNode, t);
    				}
    			} else if (if_block0) {
    				if_block0.d(1);
    				if_block0 = null;
    			}

    			if (/*isLoggedIn*/ ctx[0]) {
    				if (if_block1) {
    					if_block1.p(ctx, dirty);

    					if (dirty & /*isLoggedIn*/ 1) {
    						transition_in(if_block1, 1);
    					}
    				} else {
    					if_block1 = create_if_block(ctx);
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
    			transition_in(if_block1);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(if_block1);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (if_block0) if_block0.d(detaching);
    			if (detaching) detach_dev(t);
    			if (if_block1) if_block1.d(detaching);
    			if (detaching) detach_dev(if_block1_anchor);
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
    	let isLoggedIn;
    	let history;
    	let $store;
    	validate_store(store, "store");
    	component_subscribe($$self, store, $$value => $$invalidate(3, $store = $$value));
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots("V1", slots, []);

    	onMount(async () => {
    		if (!shared.loggedIn()) {
    			return;
    		}

    		await store.history();
    	});

    	function checkLogin() {
    		if (!shared.loggedIn()) {
    			shared.notify("info", "You need to login to see your summary", true);
    		}
    	}

    	function beforeunload(event) {
    		shared.clearNotification();
    	}

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== "$$") console.warn(`<V1> was created with unknown prop '${key}'`);
    	});

    	$$self.$capture_state = () => ({
    		loggedIn: shared.loggedIn,
    		notify: shared.notify,
    		clearNotification: shared.clearNotification,
    		store,
    		Summary,
    		onMount,
    		checkLogin,
    		beforeunload,
    		isLoggedIn,
    		history,
    		$store
    	});

    	$$self.$inject_state = $$props => {
    		if ("isLoggedIn" in $$props) $$invalidate(0, isLoggedIn = $$props.isLoggedIn);
    		if ("history" in $$props) $$invalidate(1, history = $$props.history);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*$store*/ 8) {
    			$$invalidate(1, history = $store.history);
    		}
    	};

    	checkLogin();
    	$$invalidate(0, isLoggedIn = shared.loggedIn());
    	return [isLoggedIn, history, beforeunload, $store];
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
//# sourceMappingURL=toolbox-plank-stats.js.map
