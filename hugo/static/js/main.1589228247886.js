!function(t){"use strict";function e(t,e){return localStorage.hasOwnProperty(t)?JSON.parse(localStorage.getItem(t)):e}function n(t,e){localStorage.setItem(t,JSON.stringify(e))}function r(){}function i(t){return t()}function o(){return Object.create(null)}function a(t){t.forEach(i)}function s(t){return"function"==typeof t}function l(t,e){return t!=t?e==e:t!==e||t&&"object"==typeof t||"function"==typeof t}function c(t,e,n){t.$$.on_destroy.push(function(t,...e){if(null==t)return r;const n=t.subscribe(...e);return n.unsubscribe?()=>n.unsubscribe():n}(e,n))}function u(t,e){t.appendChild(e)}function d(t,e,n){t.insertBefore(e,n||null)}function f(t){t.parentNode.removeChild(t)}function m(t){return document.createElement(t)}function h(t){return document.createElementNS("http://www.w3.org/2000/svg",t)}function p(t){return document.createTextNode(t)}function g(){return p(" ")}function b(){return p("")}function v(t,e,n,r){return t.addEventListener(e,n,r),()=>t.removeEventListener(e,n,r)}function w(t,e,n){null==n?t.removeAttribute(e):t.getAttribute(e)!==n&&t.setAttribute(e,n)}function y(t,e){(null!=e||t.value)&&(t.value=e)}function k(t,e,n,r){t.style.setProperty(e,n,r?"important":"")}function x(t,e,n){t.classList[n?"add":"remove"](e)}let $;function _(t){$=t}const C=[],z=[],S=[],E=[],L=Promise.resolve();let M=!1;function T(t){S.push(t)}let N=!1;const j=new Set;function O(){if(!N){N=!0;do{for(let t=0;t<C.length;t+=1){const e=C[t];_(e),P(e.$$)}for(C.length=0;z.length;)z.pop()();for(let t=0;t<S.length;t+=1){const e=S[t];j.has(e)||(j.add(e),e())}S.length=0}while(C.length);for(;E.length;)E.pop()();M=!1,N=!1,j.clear()}}function P(t){if(null!==t.fragment){t.update(),a(t.before_update);const e=t.dirty;t.dirty=[-1],t.fragment&&t.fragment.p(t.ctx,e),t.after_update.forEach(T)}}const H=new Set;function R(t,e){-1===t.$$.dirty[0]&&(C.push(t),M||(M=!0,L.then(O)),t.$$.dirty.fill(0)),t.$$.dirty[e/31|0]|=1<<e%31}function A(t,e,n,l,c,u,d=[-1]){const m=$;_(t);const h=e.props||{},p=t.$$={fragment:null,ctx:null,props:u,update:r,not_equal:c,bound:o(),on_mount:[],on_destroy:[],before_update:[],after_update:[],context:new Map(m?m.$$.context:[]),callbacks:o(),dirty:d};let g=!1;if(p.ctx=n?n(t,h,(e,n,...r)=>{const i=r.length?r[0]:n;return p.ctx&&c(p.ctx[e],p.ctx[e]=i)&&(p.bound[e]&&p.bound[e](i),g&&R(t,e)),n}):[],p.update(),g=!0,a(p.before_update),p.fragment=!!l&&l(p.ctx),e.target){if(e.hydrate){const t=function(t){return Array.from(t.childNodes)}(e.target);p.fragment&&p.fragment.l(t),t.forEach(f)}else p.fragment&&p.fragment.c();e.intro&&((b=t.$$.fragment)&&b.i&&(H.delete(b),b.i(v))),function(t,e,n){const{fragment:r,on_mount:o,on_destroy:l,after_update:c}=t.$$;r&&r.m(e,n),T(()=>{const e=o.map(i).filter(s);l?l.push(...e):a(e),t.$$.on_mount=[]}),c.forEach(T)}(t,e.target,e.anchor),O()}var b,v;_(m)}let I;function B(t){let e,n,r,i,o,a,s,l,c,b;return{c(){e=m("div"),n=h("svg"),r=h("title"),i=p("info icon"),o=h("path"),s=g(),l=m("span"),c=p(t[2]),w(o,"d",a=t[5](t[1].level)),w(n,"class","w1"),w(n,"data-icon","info"),w(n,"viewBox","0 0 24 24"),k(n,"fill","currentcolor"),k(n,"width","2em"),k(n,"height","2em"),w(l,"class","lh-title ml3"),w(e,"class","flex items-center justify-center pa3 navy"),x(e,"info","info"===t[0]),x(e,"error","error"===t[0])},m(a,f,m){d(a,e,f),u(e,n),u(n,r),u(r,i),u(n,o),u(e,s),u(e,l),u(l,c),m&&b(),b=v(e,"click",t[4])},p(t,n){2&n&&a!==(a=t[5](t[1].level))&&w(o,"d",a),4&n&&function(t,e){e=""+e,t.data!==e&&(t.data=e)}(c,t[2]),1&n&&x(e,"info","info"===t[0]),1&n&&x(e,"error","error"===t[0])},d(t){t&&f(e),b()}}}function J(t){let e,n=t[3]&&B(t);return{c(){n&&n.c(),e=b(),this.c=r},m(t,r){n&&n.m(t,r),d(t,e,r)},p(t,[r]){t[3]?n?n.p(t,r):(n=B(t),n.c(),n.m(e.parentNode,e)):n&&(n.d(1),n=null)},i:r,o:r,d(t){n&&n.d(t),t&&f(e)}}}function q(e,n,r){let i;c(e,t.notifications,t=>r(1,i=t));let o,a,s;return e.$$.update=()=>{2&e.$$.dirty&&r(0,o=i.level),2&e.$$.dirty&&r(2,a=i.message),1&e.$$.dirty&&r(3,s=""!=o)},[o,i,a,s,function(){t.notifications.clear()},function(t){return""==t?"":"info"==t?"M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z":"M11 15h2v2h-2zm0-8h2v6h-2zm.99-5C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z"}]}"function"==typeof HTMLElement&&(I=class extends HTMLElement{constructor(){super(),this.attachShadow({mode:"open"})}connectedCallback(){for(const t in this.$$.slotted)this.appendChild(this.$$.slotted[t])}attributeChangedCallback(t,e,n){this[t]=n}$destroy(){!function(t,e){const n=t.$$;null!==n.fragment&&(a(n.on_destroy),n.fragment&&n.fragment.d(e),n.on_destroy=n.fragment=null,n.ctx=[])}(this,1),this.$destroy=r}$on(t,e){const n=this.$$.callbacks[t]||(this.$$.callbacks[t]=[]);return n.push(e),()=>{const t=n.indexOf(e);-1!==t&&n.splice(t,1)}}$set(){}});function U(t){let e,n;return{c(){e=m("a"),n=p("Login"),w(e,"title","Click to login"),w(e,"href",t[0]),w(e,"class","f6 fw6 hover-red link black-70 mr2 mr3-m mr4-l dib")},m(t,r){d(t,e,r),u(e,n)},p(t,n){1&n&&w(e,"href",t[0])},d(t){t&&f(e)}}}function D(t){let e,n,i,o,a,s;return{c(){e=m("a"),e.textContent="Create",n=g(),i=m("a"),i.textContent="My Lists",o=g(),a=m("a"),a.textContent="Logout",w(e,"title","create, edit, share"),w(e,"href","/editor.html"),w(e,"class","f6 fw6 hover-blue link black-70 ml0 mr2-l di"),w(i,"title","Lists created by you"),w(i,"href","/lists-by-me.html"),w(i,"class","f6 fw6 hover-blue link black-70 di"),w(a,"title","Logout"),w(a,"href","/logout.html"),w(a,"class","f6 fw6 hover-blue link black-70 di ml3")},m(t,r,l){d(t,e,r),d(t,n,r),d(t,i,r),d(t,o,r),d(t,a,r),l&&s(),s=v(a,"click",F)},p:r,d(t){t&&f(e),t&&f(n),t&&f(i),t&&f(o),t&&f(a),s()}}}function V(e){let n,i;function o(e,n){return null==i&&(i=!!t.loggedIn()),i?D:window.location.pathname!=e[0]?U:void 0}let a=o(e),s=a&&a(e);return{c(){n=m("div"),s&&s.c(),this.c=r,w(n,"class","fr mt0")},m(t,e){d(t,n,e),s&&s.m(n,null)},p(t,[e]){a===(a=o(t))&&s?s.p(t,e):(s&&s.d(1),s=a&&a(t),s&&(s.c(),s.m(n,null)))},i:r,o:r,d(t){t&&f(n),s&&s.d()}}}function F(){localStorage.clear(),console.log("It should still click")}function G(t,e,n){let{loginurl:r="/login.html"}=e;return t.$set=t=>{"loginurl"in t&&n(0,r=t.loginurl)},[r]}async function K(t,n){const r={status:400,body:{}},i={username:t,password:n},o=function(){const t=e("settings.server",null);if(null===t)throw new Error("settings.server.missing");return t}()+"/api/v1/user/login",a=await fetch(o,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(i)}),s=await a.json();switch(a.status){case 200:case 403:case 400:return r.status=a.status,r.body=s,r}throw new Error("Unexpected response from the server")}function Q(t){let e,n,r,i,o,s,l,c,h,p,b,k,x,$;return{c(){e=m("form"),n=m("fieldset"),r=m("div"),i=m("label"),i.textContent="Username",o=g(),s=m("input"),l=g(),c=m("div"),h=m("label"),h.textContent="Password",p=g(),b=m("input"),k=g(),x=m("div"),x.innerHTML='<div class="w-100 items-end"><div class="fr"><div class="flex items-center mb2"><button class="db w-100" type="submit">Login</button></div> \n          <div class="flex items-center mb2"><span class="f6 link dib black">\n              or with\n              <a target="_blank" href="https://learnalist.net/api/v1/oauth/google/redirect" class="f6 link underline dib black">\n                google\n              </a></span></div></div></div>',w(i,"class","db fw6 lh-copy f6"),w(i,"for","username"),w(s,"class","pa2 input-reset ba bg-transparent b--black-20 w-100 br2"),w(s,"type","text"),w(s,"name","username"),w(s,"id","username"),w(s,"autocapitalize","none"),w(r,"class","mt3"),w(h,"class","db fw6 lh-copy f6"),w(h,"for","password"),w(b,"class","b pa2 input-reset ba bg-transparent b--black-20 w-100 br2"),w(b,"type","password"),w(b,"name","password"),w(b,"autocomplete","off"),w(b,"id","password"),w(c,"class","mv3"),w(n,"id","sign_up"),w(n,"class","ba b--transparent ph0 mh0"),w(x,"class","measure flex"),w(e,"class","measure center")},m(f,m,g){var w;d(f,e,m),u(e,n),u(n,r),u(r,i),u(r,o),u(r,s),y(s,t[0]),u(n,l),u(n,c),u(c,h),u(c,p),u(c,b),y(b,t[1]),u(e,k),u(e,x),g&&a($),$=[v(s,"input",t[4]),v(b,"input",t[5]),v(e,"submit",(w=t[2],function(t){return t.preventDefault(),w.call(this,t)}))]},p(t,e){1&e&&s.value!==t[0]&&y(s,t[0]),2&e&&b.value!==t[1]&&y(b,t[1])},d(t){t&&f(e),a($)}}}function W(t){let e;let n=Q(t);return{c(){n.c(),e=b(),this.c=r},m(t,r){n.m(t,r),d(t,e,r)},p(t,[e]){n.p(t,e)},i:r,o:r,d(t){n.d(t),t&&f(e)}}}function X(e,r,i){let o,a="",s="";return[a,s,async function(){if(""===a||""===s)return o="Please enter in a username and password",void t.notify("error",o);let e=await K(a,s);200==e.status?(n("app.user.uuid",e.body.user_uuid),n("app.user.authentication",e.body.token),t.login("/welcome.html")):t.notify("error","Please try again")},o,function(){a=this.value,i(0,a)},function(){s=this.value,i(1,s)}]}null===e("settings.install.defaults",null)&&function(){localStorage.clear(),n("settings.install.defaults",!0);const t=document.querySelector('meta[name="api.server"]');n("settings.server",t?t.content:"https://learnalist.net"),n("my.edited.lists",[]),n("my.lists",[])}(),customElements.define("login-header",class extends I{constructor(t){super(),this.shadowRoot.innerHTML="<style>a{background-color:transparent}a,div{box-sizing:border-box}.di{display:inline}.dib{display:inline-block}.fr{_display:inline}.fr{float:right}.fw6{font-weight:600}.link{text-decoration:none}.link,.link:active,.link:focus,.link:hover,.link:link,.link:visited{transition:color .15s ease-in}.link:focus{outline:1px dotted currentColor}.black-70{color:rgba(0,0,0,.7)}.hover-red:focus,.hover-red:hover{color:#ff4136}.hover-blue:focus,.hover-blue:hover{color:#357edd}.ml0{margin-left:0}.ml3{margin-left:1rem}.mr2{margin-right:.5rem}.mt0{margin-top:0}.f6{font-size:.875rem}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){.mr3-m{margin-right:1rem}}@media screen and (min-width:60em){.mr2-l{margin-right:.5rem}.mr4-l{margin-right:2rem}}</style>",A(this,{target:this.shadowRoot},G,V,l,{loginurl:0}),t&&(t.target&&d(t.target,this,t.anchor),t.props&&(this.$set(t.props),O()))}static get observedAttributes(){return["loginurl"]}get loginurl(){return this.$$.ctx[0]}set loginurl(t){this.$set({loginurl:t}),O()}}),customElements.define("user-login",class extends I{constructor(t){super(),this.shadowRoot.innerHTML="<style>a{background-color:transparent}button,input{font-family:inherit;font-size:100%;line-height:1.15;margin:0}button,input{overflow:visible}button{text-transform:none}button{-webkit-appearance:button}button::-moz-focus-inner{border-style:none;padding:0}button:-moz-focusring{outline:1px dotted ButtonText}fieldset{padding:.35em .75em .625em}a,div,fieldset,form,p{box-sizing:border-box}.ba{border-style:solid;border-width:1px}.b--black-20{border-color:rgba(0,0,0,.2)}.b--transparent{border-color:transparent}.br2{border-radius:.25rem}.db{display:block}.dib{display:inline-block}.flex{display:flex}.items-end{align-items:flex-end}.items-center{align-items:center}.fr{_display:inline}.fr{float:right}.b{font-weight:700}.fw6{font-weight:600}.input-reset{-webkit-appearance:none;-moz-appearance:none}.input-reset::-moz-focus-inner{border:0;padding:0}.lh-copy{line-height:1.5}.link{text-decoration:none}.link,.link:active,.link:focus,.link:hover,.link:link,.link:visited{transition:color .15s ease-in}.link:focus{outline:1px dotted currentColor}.w-100{width:100%}.black{color:#000}.bg-transparent{background-color:transparent}.pa2{padding:.5rem}.ph0{padding-left:0;padding-right:0}.mb2{margin-bottom:.5rem}.mt3{margin-top:1rem}.mv3{margin-top:1rem;margin-bottom:1rem}.mh0{margin-left:0;margin-right:0}.underline{text-decoration:underline}.f6{font-size:.875rem}.measure{max-width:30em}.center{margin-left:auto}.center{margin-right:auto}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){}@media screen and (min-width:60em){}</style>",A(this,{target:this.shadowRoot},X,W,l,{}),t&&t.target&&d(t.target,this,t.anchor)}}),customElements.define("notification-center",class extends I{constructor(t){super(),this.shadowRoot.innerHTML="<style>div{box-sizing:border-box}.flex{display:flex}.items-center{align-items:center}.justify-center{justify-content:center}.lh-title{line-height:1.25}.w1{width:1rem}.navy{color:#001b44}.pa3{padding:1rem}.ml3{margin-left:1rem}@media screen and (min-width:30em){}@media screen and (min-width:30em) and (max-width:60em){}@media screen and (min-width:60em){}.error{background-color:#ffdfdf}.info{background-color:#96ccff}</style>",A(this,{target:this.shadowRoot},q,J,l,{}),t&&t.target&&d(t.target,this,t.anchor)}})}(superstore);
