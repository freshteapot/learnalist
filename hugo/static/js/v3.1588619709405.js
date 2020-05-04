!function(t){"use strict";function e(){}function n(t){return t()}function r(){return Object.create(null)}function o(t){t.forEach(n)}function i(t){return"function"==typeof t}function l(t,e){return t!=t?e==e:t!==e||t&&"object"==typeof t||"function"==typeof t}function c(t,n,r){t.$$.on_destroy.push(function(t,...n){if(null==t)return e;const r=t.subscribe(...n);return r.unsubscribe?()=>r.unsubscribe():r}(n,r))}function s(t,e){t.appendChild(e)}function a(t,e,n){t.insertBefore(e,n||null)}function u(t){t.parentNode.removeChild(t)}function d(t){return document.createElement(t)}function f(t){return document.createElementNS("http://www.w3.org/2000/svg",t)}function m(t){return document.createTextNode(t)}function h(){return m(" ")}function g(t,e,n,r){return t.addEventListener(e,n,r),()=>t.removeEventListener(e,n,r)}function p(t,e,n){null==n?t.removeAttribute(e):t.getAttribute(e)!==n&&t.setAttribute(e,n)}function b(t,e,n,r){t.style.setProperty(e,n,r?"important":"")}function $(t,e,n){t.classList[n?"add":"remove"](e)}let y;function v(t){y=t}const w=[],k=[],x=[],_=[],C=Promise.resolve();let E=!1;function L(t){x.push(t)}let z=!1;const M=new Set;function S(){if(!z){z=!0;do{for(let t=0;t<w.length;t+=1){const e=w[t];v(e),j(e.$$)}for(w.length=0;k.length;)k.pop()();for(let t=0;t<x.length;t+=1){const e=x[t];M.has(e)||(M.add(e),e())}x.length=0}while(w.length);for(;_.length;)_.pop()();E=!1,z=!1,M.clear()}}function j(t){if(null!==t.fragment){t.update(),o(t.before_update);const e=t.dirty;t.dirty=[-1],t.fragment&&t.fragment.p(t.ctx,e),t.after_update.forEach(L)}}const A=new Set;function N(t,e){-1===t.$$.dirty[0]&&(w.push(t),E||(E=!0,C.then(S)),t.$$.dirty.fill(0)),t.$$.dirty[e/31|0]|=1<<e%31}function T(t,l,c,s,a,d,f=[-1]){const m=y;v(t);const h=l.props||{},g=t.$$={fragment:null,ctx:null,props:d,update:e,not_equal:a,bound:r(),on_mount:[],on_destroy:[],before_update:[],after_update:[],context:new Map(m?m.$$.context:[]),callbacks:r(),dirty:f};let p=!1;if(g.ctx=c?c(t,h,(e,n,...r)=>{const o=r.length?r[0]:n;return g.ctx&&a(g.ctx[e],g.ctx[e]=o)&&(g.bound[e]&&g.bound[e](o),p&&N(t,e)),n}):[],g.update(),p=!0,o(g.before_update),g.fragment=!!s&&s(g.ctx),l.target){if(l.hydrate){const t=function(t){return Array.from(t.childNodes)}(l.target);g.fragment&&g.fragment.l(t),t.forEach(u)}else g.fragment&&g.fragment.c();l.intro&&((b=t.$$.fragment)&&b.i&&(A.delete(b),b.i($))),function(t,e,r){const{fragment:l,on_mount:c,on_destroy:s,after_update:a}=t.$$;l&&l.m(e,r),L(()=>{const e=c.map(n).filter(i);s?s.push(...e):o(e),t.$$.on_mount=[]}),a.forEach(L)}(t,l.target,l.anchor),S()}var b,$;v(m)}let H;function R(t){let e,n;return{c(){e=d("a"),n=m("Login"),p(e,"title","Click to login"),p(e,"href",t[0]),p(e,"class","f6 fw6 hover-red link black-70 mr2 mr3-m mr4-l dib")},m(t,r){a(t,e,r),s(e,n)},p(t,n){1&n&&p(e,"href",t[0])},d(t){t&&u(e)}}}function B(t){let n,r,o,i,l,c;return{c(){n=d("a"),n.textContent="Create",r=h(),o=d("a"),o.textContent="My Lists",i=h(),l=d("a"),l.textContent="Logout",p(n,"title","create, edit, share"),p(n,"href","/editor.html"),p(n,"class","f6 fw6 hover-blue link black-70 ml0 mr2-l di"),p(o,"title","Lists created by you"),p(o,"href","/lists-by-me.html"),p(o,"class","f6 fw6 hover-blue link black-70 di"),p(l,"title","Logout"),p(l,"href","/logout.html"),p(l,"class","f6 fw6 hover-blue link black-70 di ml3")},m(t,e,s){a(t,n,e),a(t,r,e),a(t,o,e),a(t,i,e),a(t,l,e),s&&c(),c=g(l,"click",O)},p:e,d(t){t&&u(n),t&&u(r),t&&u(o),t&&u(i),t&&u(l),c()}}}function I(n){let r,o;function i(e,n){return null==o&&(o=!!t.loggedIn()),o?B:window.location.pathname!=e[0]?R:void 0}let l=i(n),c=l&&l(n);return{c(){r=d("div"),c&&c.c(),this.c=e,p(r,"class","fr mt0")},m(t,e){a(t,r,e),c&&c.m(r,null)},p(t,[e]){l===(l=i(t))&&c?c.p(t,e):(c&&c.d(1),c=l&&l(t),c&&(c.c(),c.m(r,null)))},i:e,o:e,d(t){t&&u(r),c&&c.d()}}}function O(){localStorage.clear(),console.log("It should still click")}function P(t,e,n){let{loginurl:r="/login.html"}=e;return t.$set=t=>{"loginurl"in t&&n(0,r=t.loginurl)},[r]}"function"==typeof HTMLElement&&(H=class extends HTMLElement{constructor(){super(),this.attachShadow({mode:"open"})}connectedCallback(){for(const t in this.$$.slotted)this.appendChild(this.$$.slotted[t])}attributeChangedCallback(t,e,n){this[t]=n}$destroy(){!function(t,e){const n=t.$$;null!==n.fragment&&(o(n.on_destroy),n.fragment&&n.fragment.d(e),n.on_destroy=n.fragment=null,n.ctx=[])}(this,1),this.$destroy=e}$on(t,e){const n=this.$$.callbacks[t]||(this.$$.callbacks[t]=[]);return n.push(e),()=>{const t=n.indexOf(e);-1!==t&&n.splice(t,1)}}$set(){}});function q(t){let e,n,r,o,i,l,c,y,v,w;return{c(){e=d("div"),n=f("svg"),r=f("title"),o=m("info icon"),i=f("path"),c=h(),y=d("span"),v=m(t[2]),p(i,"d",l=t[4](t[1].level)),p(n,"class","w1"),p(n,"data-icon","info"),p(n,"viewBox","0 0 24 24"),b(n,"fill","currentcolor"),b(n,"width","2em"),b(n,"height","2em"),p(y,"class","lh-title ml3"),p(e,"class","flex items-center justify-center pa3 navy"),$(e,"info","info"===t[0]),$(e,"error","error"===t[0])},m(t,l,u){a(t,e,l),s(e,n),s(n,r),s(r,o),s(n,i),s(e,c),s(e,y),s(y,v),u&&w(),w=g(e,"click",D)},p(t,n){2&n&&l!==(l=t[4](t[1].level))&&p(i,"d",l),4&n&&function(t,e){e=""+e,t.data!==e&&(t.data=e)}(v,t[2]),1&n&&$(e,"info","info"===t[0]),1&n&&$(e,"error","error"===t[0])},d(t){t&&u(e),w()}}}function V(t){let n,r=t[3]&&q(t);return{c(){r&&r.c(),n=m(""),this.c=e},m(t,e){r&&r.m(t,e),a(t,n,e)},p(t,[e]){t[3]?r?r.p(t,e):(r=q(t),r.c(),r.m(n.parentNode,n)):r&&(r.d(1),r=null)},i:e,o:e,d(t){r&&r.d(t),t&&u(n)}}}function D(){t.notifications.clear()}function F(e,n,r){let o;c(e,t.notifications,t=>r(1,o=t));let i,l,s;return e.$$.update=()=>{2&e.$$.dirty&&r(0,i=o.level),2&e.$$.dirty&&r(2,l=o.message),1&e.$$.dirty&&r(3,s=""!=i)},[i,o,l,s,function(t){return""==t?"":"info"==t?"M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z":"M11 15h2v2h-2zm0-8h2v6h-2zm.99-5C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z"}]}customElements.define("login-header",class extends H{constructor(t){super(),this.shadowRoot.innerHTML="<style>a{background-color:transparent}a,div{box-sizing:border-box}.di{display:inline}.dib{display:inline-block}.fr{_display:inline}.fr{float:right}.fw6{font-weight:600}.link{text-decoration:none}.link,.link:active,.link:focus,.link:hover,.link:link,.link:visited{transition:color .15s ease-in}.link:focus{outline:1px dotted currentColor}.black-70{color:rgba(0,0,0,.7)}.hover-red:focus,.hover-red:hover{color:#ff4136}.hover-blue:focus,.hover-blue:hover{color:#357edd}.ml0{margin-left:0}.ml3{margin-left:1rem}.mr2{margin-right:.5rem}.mt0{margin-top:0}.f6{font-size:.875rem}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){.mr3-m{margin-right:1rem}}@media screen and (min-width:60em){.mr2-l{margin-right:.5rem}.mr4-l{margin-right:2rem}}</style>",T(this,{target:this.shadowRoot},P,I,l,{loginurl:0}),t&&(t.target&&a(t.target,this,t.anchor),t.props&&(this.$set(t.props),S()))}static get observedAttributes(){return["loginurl"]}get loginurl(){return this.$$.ctx[0]}set loginurl(t){this.$set({loginurl:t}),S()}}),customElements.define("notification-center",class extends H{constructor(t){super(),this.shadowRoot.innerHTML="<style>div{box-sizing:border-box}.flex{display:flex}.items-center{align-items:center}.justify-center{justify-content:center}.lh-title{line-height:1.25}.w1{width:1rem}.navy{color:#001b44}.pa3{padding:1rem}.ml3{margin-left:1rem}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){}@media screen and (min-width:60em){}.error{background-color:#ffdfdf}.info{background-color:#96ccff}</style>",T(this,{target:this.shadowRoot},F,V,l,{}),t&&t.target&&a(t.target,this,t.anchor)}})}(superstore);
