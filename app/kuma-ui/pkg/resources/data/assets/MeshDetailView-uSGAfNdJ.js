import{d as qe,J as or,c as Dt,Q as ar,R as ir,S as cr,T as sr,U as ur,h as J,o as w,a as ve,P as xt,w as E,g as re,t as B,b as F,l as lr,e as V,i as Je,j as M,k as C,z as We,G as kt,F as ye,B as Xe,V as St,W as Ye,A as fr}from"./index-CMlVV7ds.js";import{_ as dr}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-CHvkb6Jy.js";import"./CodeBlock-BzjwXdv1.js";import"./toYaml-DB9FPXFY.js";const pr=(t,n)=>{const o=t.__vccOpts||t;for(const[s,l]of n)o[s]=l;return o},vr={class:"about-section-title","data-testid":"about-section-title"},br={class:"about-section-header-end"},hr={key:0,class:"about-section-timestamps-created","data-testid":"about-section-timestamps-created"},mr={key:1,class:"about-section-timestamps-arrow"},yr={key:2,class:"about-section-timestamps-modified","data-testid":"about-section-timestamps-modified"},gr={key:1,class:"about-section-actions","data-testid":"about-section-actions"},_r={key:0,"data-testid":"about-section-loading-skeleton"},wr={key:1},Or={key:0,class:"about-section-description","data-testid":"about-section-description"},kr={key:1,class:"about-section-content","data-testid":"about-section-content"},Sr={key:2,"data-testid":"about-divider-section-separator"},Er={key:3,class:"about-divider-section","data-testid":"about-divider-section"},jr=qe({__name:"AppAboutSection",props:{title:{type:String,default:""},description:{type:String,default:""},created:{type:String,default:""},createdLabel:{type:String,default:"Created"},modified:{type:String,default:""},modifiedLabel:{type:String,default:"Modified"},isLoading:{type:Boolean,default:!1}},setup(t){const n=t,o=Dt(()=>!!n.modified&&n.modified!==n.created);return(s,l)=>{const i=J("KSkeletonBox"),p=J("KCard");return w(),ve(p,{class:"kong-ui-app-about-section","title-tag":"h2"},xt({default:E(()=>[t.isLoading?(w(),F("div",_r,[M(i,{height:"2",width:"100"}),M(i,{height:"2",width:"100"})])):(w(),F("div",wr,[t.description?(w(),F("p",Or,B(t.description),1)):V("",!0),s.$slots.default?(w(),F("div",kr,[Je(s.$slots,"default",{},void 0,!0)])):V("",!0),s.$slots["divider-section"]?(w(),F("hr",Sr)):V("",!0),s.$slots["divider-section"]?(w(),F("div",Er,[Je(s.$slots,"divider-section",{},void 0,!0)])):V("",!0)]))]),_:2},[t.title?{name:"title",fn:E(()=>[re("span",vr,B(t.title),1)]),key:"0"}:void 0,s.$slots.actions||t.created||t.modified?{name:"actions",fn:E(()=>[re("div",br,[t.created||t.modified?(w(),F("div",{key:0,class:lr(["about-section-timestamps",{"has-actions":s.$slots.actions}])},[t.created?(w(),F("span",hr,B(t.createdLabel)+": "+B(t.created),1)):V("",!0),t.created&&o.value?(w(),F("span",mr,"->")):V("",!0),o.value?(w(),F("span",yr,B(t.modifiedLabel)+": "+B(t.modified),1)):V("",!0)],2)):V("",!0),s.$slots.actions?(w(),F("div",gr,[Je(s.$slots,"actions",{},void 0,!0)])):V("",!0)])]),key:"1"}:void 0]),1024)}}}),Nr=pr(jr,[["__scopeId","data-v-521dcc27"]]);/*!
* tabbable 6.2.0
* @license MIT, https://github.com/focus-trap/tabbable/blob/master/LICENSE
*/var Rt=["input:not([inert])","select:not([inert])","textarea:not([inert])","a[href]:not([inert])","button:not([inert])","[tabindex]:not(slot):not([inert])","audio[controls]:not([inert])","video[controls]:not([inert])",'[contenteditable]:not([contenteditable="false"]):not([inert])',"details>summary:first-of-type:not([inert])","details:not([inert])"],Ae=Rt.join(","),Ct=typeof Element>"u",ae=Ct?function(){}:Element.prototype.matches||Element.prototype.msMatchesSelector||Element.prototype.webkitMatchesSelector,De=!Ct&&Element.prototype.getRootNode?function(t){var n;return t==null||(n=t.getRootNode)===null||n===void 0?void 0:n.call(t)}:function(t){return t==null?void 0:t.ownerDocument},xe=function t(n,o){var s;o===void 0&&(o=!0);var l=n==null||(s=n.getAttribute)===null||s===void 0?void 0:s.call(n,"inert"),i=l===""||l==="true",p=i||o&&n&&t(n.parentNode);return p},Fr=function(t){var n,o=t==null||(n=t.getAttribute)===null||n===void 0?void 0:n.call(t,"contenteditable");return o===""||o==="true"},Pt=function(t,n,o){if(xe(t))return[];var s=Array.prototype.slice.apply(t.querySelectorAll(Ae));return n&&ae.call(t,Ae)&&s.unshift(t),s=s.filter(o),s},It=function t(n,o,s){for(var l=[],i=Array.from(n);i.length;){var p=i.shift();if(!xe(p,!1))if(p.tagName==="SLOT"){var v=p.assignedElements(),y=v.length?v:p.children,j=t(y,!0,s);s.flatten?l.push.apply(l,j):l.push({scopeParent:p,candidates:j})}else{var T=ae.call(p,Ae);T&&s.filter(p)&&(o||!n.includes(p))&&l.push(p);var A=p.shadowRoot||typeof s.getShadowRoot=="function"&&s.getShadowRoot(p),O=!xe(A,!1)&&(!s.shadowRootFilter||s.shadowRootFilter(p));if(A&&O){var P=t(A===!0?p.children:A.children,!0,s);s.flatten?l.push.apply(l,P):l.push({scopeParent:p,candidates:P})}else i.unshift.apply(i,p.children)}}return l},Bt=function(t){return!isNaN(parseInt(t.getAttribute("tabindex"),10))},oe=function(t){if(!t)throw new Error("No node provided");return t.tabIndex<0&&(/^(AUDIO|VIDEO|DETAILS)$/.test(t.tagName)||Fr(t))&&!Bt(t)?0:t.tabIndex},Tr=function(t,n){var o=oe(t);return o<0&&n&&!Bt(t)?0:o},Ar=function(t,n){return t.tabIndex===n.tabIndex?t.documentOrder-n.documentOrder:t.tabIndex-n.tabIndex},Lt=function(t){return t.tagName==="INPUT"},Dr=function(t){return Lt(t)&&t.type==="hidden"},xr=function(t){var n=t.tagName==="DETAILS"&&Array.prototype.slice.apply(t.children).some(function(o){return o.tagName==="SUMMARY"});return n},Rr=function(t,n){for(var o=0;o<t.length;o++)if(t[o].checked&&t[o].form===n)return t[o]},Cr=function(t){if(!t.name)return!0;var n=t.form||De(t),o=function(i){return n.querySelectorAll('input[type="radio"][name="'+i+'"]')},s;if(typeof window<"u"&&typeof window.CSS<"u"&&typeof window.CSS.escape=="function")s=o(window.CSS.escape(t.name));else try{s=o(t.name)}catch(i){return console.error("Looks like you have a radio button with a name attribute containing invalid CSS selector characters and need the CSS.escape polyfill: %s",i.message),!1}var l=Rr(s,t.form);return!l||l===t},Pr=function(t){return Lt(t)&&t.type==="radio"},Ir=function(t){return Pr(t)&&!Cr(t)},Br=function(t){var n,o=t&&De(t),s=(n=o)===null||n===void 0?void 0:n.host,l=!1;if(o&&o!==t){var i,p,v;for(l=!!((i=s)!==null&&i!==void 0&&(p=i.ownerDocument)!==null&&p!==void 0&&p.contains(s)||t!=null&&(v=t.ownerDocument)!==null&&v!==void 0&&v.contains(t));!l&&s;){var y,j,T;o=De(s),s=(y=o)===null||y===void 0?void 0:y.host,l=!!((j=s)!==null&&j!==void 0&&(T=j.ownerDocument)!==null&&T!==void 0&&T.contains(s))}}return l},Et=function(t){var n=t.getBoundingClientRect(),o=n.width,s=n.height;return o===0&&s===0},Lr=function(t,n){var o=n.displayCheck,s=n.getShadowRoot;if(getComputedStyle(t).visibility==="hidden")return!0;var l=ae.call(t,"details>summary:first-of-type"),i=l?t.parentElement:t;if(ae.call(i,"details:not([open]) *"))return!0;if(!o||o==="full"||o==="legacy-full"){if(typeof s=="function"){for(var p=t;t;){var v=t.parentElement,y=De(t);if(v&&!v.shadowRoot&&s(v)===!0)return Et(t);t.assignedSlot?t=t.assignedSlot:!v&&y!==t.ownerDocument?t=y.host:t=v}t=p}if(Br(t))return!t.getClientRects().length;if(o!=="legacy-full")return!0}else if(o==="non-zero-area")return Et(t);return!1},Mr=function(t){if(/^(INPUT|BUTTON|SELECT|TEXTAREA)$/.test(t.tagName))for(var n=t.parentElement;n;){if(n.tagName==="FIELDSET"&&n.disabled){for(var o=0;o<n.children.length;o++){var s=n.children.item(o);if(s.tagName==="LEGEND")return ae.call(n,"fieldset[disabled] *")?!0:!s.contains(t)}return!0}n=n.parentElement}return!1},Re=function(t,n){return!(n.disabled||xe(n)||Dr(n)||Lr(n,t)||xr(n)||Mr(n))},Qe=function(t,n){return!(Ir(n)||oe(n)<0||!Re(t,n))},$r=function(t){var n=parseInt(t.getAttribute("tabindex"),10);return!!(isNaN(n)||n>=0)},Gr=function t(n){var o=[],s=[];return n.forEach(function(l,i){var p=!!l.scopeParent,v=p?l.scopeParent:l,y=Tr(v,p),j=p?t(l.candidates):v;y===0?p?o.push.apply(o,j):o.push(v):s.push({documentOrder:i,tabIndex:y,item:l,isScope:p,content:j})}),s.sort(Ar).reduce(function(l,i){return i.isScope?l.push.apply(l,i.content):l.push(i.content),l},[]).concat(o)},Kr=function(t,n){n=n||{};var o;return n.getShadowRoot?o=It([t],n.includeContainer,{filter:Qe.bind(null,n),flatten:!1,getShadowRoot:n.getShadowRoot,shadowRootFilter:$r}):o=Pt(t,n.includeContainer,Qe.bind(null,n)),Gr(o)},Vr=function(t,n){n=n||{};var o;return n.getShadowRoot?o=It([t],n.includeContainer,{filter:Re.bind(null,n),flatten:!0,getShadowRoot:n.getShadowRoot}):o=Pt(t,n.includeContainer,Re.bind(null,n)),o},pe=function(t,n){if(n=n||{},!t)throw new Error("No node provided");return ae.call(t,Ae)===!1?!1:Qe(n,t)},Ur=Rt.concat("iframe").join(","),Ze=function(t,n){if(n=n||{},!t)throw new Error("No node provided");return ae.call(t,Ur)===!1?!1:Re(n,t)};/*!
* focus-trap 7.5.4
* @license MIT, https://github.com/focus-trap/focus-trap/blob/master/LICENSE
*/function jt(t,n){var o=Object.keys(t);if(Object.getOwnPropertySymbols){var s=Object.getOwnPropertySymbols(t);n&&(s=s.filter(function(l){return Object.getOwnPropertyDescriptor(t,l).enumerable})),o.push.apply(o,s)}return o}function Nt(t){for(var n=1;n<arguments.length;n++){var o=arguments[n]!=null?arguments[n]:{};n%2?jt(Object(o),!0).forEach(function(s){Hr(t,s,o[s])}):Object.getOwnPropertyDescriptors?Object.defineProperties(t,Object.getOwnPropertyDescriptors(o)):jt(Object(o)).forEach(function(s){Object.defineProperty(t,s,Object.getOwnPropertyDescriptor(o,s))})}return t}function Hr(t,n,o){return n=Jr(n),n in t?Object.defineProperty(t,n,{value:o,enumerable:!0,configurable:!0,writable:!0}):t[n]=o,t}function zr(t,n){if(typeof t!="object"||t===null)return t;var o=t[Symbol.toPrimitive];if(o!==void 0){var s=o.call(t,n||"default");if(typeof s!="object")return s;throw new TypeError("@@toPrimitive must return a primitive value.")}return(n==="string"?String:Number)(t)}function Jr(t){var n=zr(t,"string");return typeof n=="symbol"?n:String(n)}var Ft={activateTrap:function(t,n){if(t.length>0){var o=t[t.length-1];o!==n&&o.pause()}var s=t.indexOf(n);s===-1||t.splice(s,1),t.push(n)},deactivateTrap:function(t,n){var o=t.indexOf(n);o!==-1&&t.splice(o,1),t.length>0&&t[t.length-1].unpause()}},Wr=function(t){return t.tagName&&t.tagName.toLowerCase()==="input"&&typeof t.select=="function"},Xr=function(t){return(t==null?void 0:t.key)==="Escape"||(t==null?void 0:t.key)==="Esc"||(t==null?void 0:t.keyCode)===27},_e=function(t){return(t==null?void 0:t.key)==="Tab"||(t==null?void 0:t.keyCode)===9},Yr=function(t){return _e(t)&&!t.shiftKey},Zr=function(t){return _e(t)&&t.shiftKey},Tt=function(t){return setTimeout(t,0)},At=function(t,n){var o=-1;return t.every(function(s,l){return n(s)?(o=l,!1):!0}),o},ge=function(t){for(var n=arguments.length,o=new Array(n>1?n-1:0),s=1;s<n;s++)o[s-1]=arguments[s];return typeof t=="function"?t.apply(void 0,o):t},Fe=function(t){return t.target.shadowRoot&&typeof t.composedPath=="function"?t.composedPath()[0]:t.target},Qr=[],qr=function(t,n){var o=(n==null?void 0:n.document)||document,s=(n==null?void 0:n.trapStack)||Qr,l=Nt({returnFocusOnDeactivate:!0,escapeDeactivates:!0,delayInitialFocus:!0,isKeyForward:Yr,isKeyBackward:Zr},n),i={containers:[],containerGroups:[],tabbableGroups:[],nodeFocusedBeforeActivation:null,mostRecentlyFocusedNode:null,active:!1,paused:!1,delayInitialFocusTimer:void 0,recentNavEvent:void 0},p,v=function(u,c,d){return u&&u[c]!==void 0?u[c]:l[d||c]},y=function(u,c){var d=typeof(c==null?void 0:c.composedPath)=="function"?c.composedPath():void 0;return i.containerGroups.findIndex(function(m){var h=m.container,_=m.tabbableNodes;return h.contains(u)||(d==null?void 0:d.includes(h))||_.find(function(k){return k===u})})},j=function(u){var c=l[u];if(typeof c=="function"){for(var d=arguments.length,m=new Array(d>1?d-1:0),h=1;h<d;h++)m[h-1]=arguments[h];c=c.apply(void 0,m)}if(c===!0&&(c=void 0),!c){if(c===void 0||c===!1)return c;throw new Error("`".concat(u,"` was specified but was not a node, or did not return a node"))}var _=c;if(typeof c=="string"&&(_=o.querySelector(c),!_))throw new Error("`".concat(u,"` as selector refers to no known node"));return _},T=function(){var u=j("initialFocus");if(u===!1)return!1;if(u===void 0||!Ze(u,l.tabbableOptions))if(y(o.activeElement)>=0)u=o.activeElement;else{var c=i.tabbableGroups[0],d=c&&c.firstTabbableNode;u=d||j("fallbackFocus")}if(!u)throw new Error("Your focus-trap needs to have at least one focusable element");return u},A=function(){if(i.containerGroups=i.containers.map(function(u){var c=Kr(u,l.tabbableOptions),d=Vr(u,l.tabbableOptions),m=c.length>0?c[0]:void 0,h=c.length>0?c[c.length-1]:void 0,_=d.find(function(S){return pe(S)}),k=d.slice().reverse().find(function(S){return pe(S)}),R=!!c.find(function(S){return oe(S)>0});return{container:u,tabbableNodes:c,focusableNodes:d,posTabIndexesFound:R,firstTabbableNode:m,lastTabbableNode:h,firstDomTabbableNode:_,lastDomTabbableNode:k,nextTabbableNode:function(S){var Z=arguments.length>1&&arguments[1]!==void 0?arguments[1]:!0,b=c.indexOf(S);return b<0?Z?d.slice(d.indexOf(S)+1).find(function(H){return pe(H)}):d.slice(0,d.indexOf(S)).reverse().find(function(H){return pe(H)}):c[b+(Z?1:-1)]}}}),i.tabbableGroups=i.containerGroups.filter(function(u){return u.tabbableNodes.length>0}),i.tabbableGroups.length<=0&&!j("fallbackFocus"))throw new Error("Your focus-trap must have at least one container with at least one tabbable node in it at all times");if(i.containerGroups.find(function(u){return u.posTabIndexesFound})&&i.containerGroups.length>1)throw new Error("At least one node with a positive tabindex was found in one of your focus-trap's multiple containers. Positive tabindexes are only supported in single-container focus-traps.")},O=function u(c){var d=c.activeElement;if(d)return d.shadowRoot&&d.shadowRoot.activeElement!==null?u(d.shadowRoot):d},P=function u(c){if(c!==!1&&c!==O(document)){if(!c||!c.focus){u(T());return}c.focus({preventScroll:!!l.preventScroll}),i.mostRecentlyFocusedNode=c,Wr(c)&&c.select()}},g=function(u){var c=j("setReturnFocus",u);return c||(c===!1?!1:u)},U=function(u){var c=u.target,d=u.event,m=u.isBackward,h=m===void 0?!1:m;c=c||Fe(d),A();var _=null;if(i.tabbableGroups.length>0){var k=y(c,d),R=k>=0?i.containerGroups[k]:void 0;if(k<0)h?_=i.tabbableGroups[i.tabbableGroups.length-1].lastTabbableNode:_=i.tabbableGroups[0].firstTabbableNode;else if(h){var S=At(i.tabbableGroups,function(ce){var se=ce.firstTabbableNode;return c===se});if(S<0&&(R.container===c||Ze(c,l.tabbableOptions)&&!pe(c,l.tabbableOptions)&&!R.nextTabbableNode(c,!1))&&(S=k),S>=0){var Z=S===0?i.tabbableGroups.length-1:S-1,b=i.tabbableGroups[Z];_=oe(c)>=0?b.lastTabbableNode:b.lastDomTabbableNode}else _e(d)||(_=R.nextTabbableNode(c,!1))}else{var H=At(i.tabbableGroups,function(ce){var se=ce.lastTabbableNode;return c===se});if(H<0&&(R.container===c||Ze(c,l.tabbableOptions)&&!pe(c,l.tabbableOptions)&&!R.nextTabbableNode(c))&&(H=k),H>=0){var Pe=H===i.tabbableGroups.length-1?0:H+1,G=i.tabbableGroups[Pe];_=oe(c)>=0?G.firstTabbableNode:G.firstDomTabbableNode}else _e(d)||(_=R.nextTabbableNode(c))}}else _=j("fallbackFocus");return _},D=function(u){var c=Fe(u);if(!(y(c,u)>=0)){if(ge(l.clickOutsideDeactivates,u)){p.deactivate({returnFocus:l.returnFocusOnDeactivate});return}ge(l.allowOutsideClick,u)||u.preventDefault()}},x=function(u){var c=Fe(u),d=y(c,u)>=0;if(d||c instanceof Document)d&&(i.mostRecentlyFocusedNode=c);else{u.stopImmediatePropagation();var m,h=!0;if(i.mostRecentlyFocusedNode)if(oe(i.mostRecentlyFocusedNode)>0){var _=y(i.mostRecentlyFocusedNode),k=i.containerGroups[_].tabbableNodes;if(k.length>0){var R=k.findIndex(function(S){return S===i.mostRecentlyFocusedNode});R>=0&&(l.isKeyForward(i.recentNavEvent)?R+1<k.length&&(m=k[R+1],h=!1):R-1>=0&&(m=k[R-1],h=!1))}}else i.containerGroups.some(function(S){return S.tabbableNodes.some(function(Z){return oe(Z)>0})})||(h=!1);else h=!1;h&&(m=U({target:i.mostRecentlyFocusedNode,isBackward:l.isKeyBackward(i.recentNavEvent)})),P(m||i.mostRecentlyFocusedNode||T())}i.recentNavEvent=void 0},$=function(u){var c=arguments.length>1&&arguments[1]!==void 0?arguments[1]:!1;i.recentNavEvent=u;var d=U({event:u,isBackward:c});d&&(_e(u)&&u.preventDefault(),P(d))},W=function(u){if(Xr(u)&&ge(l.escapeDeactivates,u)!==!1){u.preventDefault(),p.deactivate();return}(l.isKeyForward(u)||l.isKeyBackward(u))&&$(u,l.isKeyBackward(u))},X=function(u){var c=Fe(u);y(c,u)>=0||ge(l.clickOutsideDeactivates,u)||ge(l.allowOutsideClick,u)||(u.preventDefault(),u.stopImmediatePropagation())},be=function(){if(i.active)return Ft.activateTrap(s,p),i.delayInitialFocusTimer=l.delayInitialFocus?Tt(function(){P(T())}):P(T()),o.addEventListener("focusin",x,!0),o.addEventListener("mousedown",D,{capture:!0,passive:!1}),o.addEventListener("touchstart",D,{capture:!0,passive:!1}),o.addEventListener("click",X,{capture:!0,passive:!1}),o.addEventListener("keydown",W,{capture:!0,passive:!1}),p},q=function(){if(i.active)return o.removeEventListener("focusin",x,!0),o.removeEventListener("mousedown",D,!0),o.removeEventListener("touchstart",D,!0),o.removeEventListener("click",X,!0),o.removeEventListener("keydown",W,!0),p},we=function(u){var c=u.some(function(d){var m=Array.from(d.removedNodes);return m.some(function(h){return h===i.mostRecentlyFocusedNode})});c&&P(T())},ie=typeof window<"u"&&"MutationObserver"in window?new MutationObserver(we):void 0,Y=function(){ie&&(ie.disconnect(),i.active&&!i.paused&&i.containers.map(function(u){ie.observe(u,{subtree:!0,childList:!0})}))};return p={get active(){return i.active},get paused(){return i.paused},activate:function(u){if(i.active)return this;var c=v(u,"onActivate"),d=v(u,"onPostActivate"),m=v(u,"checkCanFocusTrap");m||A(),i.active=!0,i.paused=!1,i.nodeFocusedBeforeActivation=o.activeElement,c==null||c();var h=function(){m&&A(),be(),Y(),d==null||d()};return m?(m(i.containers.concat()).then(h,h),this):(h(),this)},deactivate:function(u){if(!i.active)return this;var c=Nt({onDeactivate:l.onDeactivate,onPostDeactivate:l.onPostDeactivate,checkCanReturnFocus:l.checkCanReturnFocus},u);clearTimeout(i.delayInitialFocusTimer),i.delayInitialFocusTimer=void 0,q(),i.active=!1,i.paused=!1,Y(),Ft.deactivateTrap(s,p);var d=v(c,"onDeactivate"),m=v(c,"onPostDeactivate"),h=v(c,"checkCanReturnFocus"),_=v(c,"returnFocus","returnFocusOnDeactivate");d==null||d();var k=function(){Tt(function(){_&&P(g(i.nodeFocusedBeforeActivation)),m==null||m()})};return _&&h?(h(g(i.nodeFocusedBeforeActivation)).then(k,k),this):(k(),this)},pause:function(u){if(i.paused||!i.active)return this;var c=v(u,"onPause"),d=v(u,"onPostPause");return i.paused=!0,c==null||c(),q(),Y(),d==null||d(),this},unpause:function(u){if(!i.paused||!i.active)return this;var c=v(u,"onUnpause"),d=v(u,"onPostUnpause");return i.paused=!1,c==null||c(),A(),be(),Y(),d==null||d(),this},updateContainerElements:function(u){var c=[].concat(u).filter(Boolean);return i.containers=c.map(function(d){return typeof d=="string"?o.querySelector(d):d}),i.active&&A(),Y(),this}},p.updateContainerElements(t),p};/*!
  * focus-trap-vue v4.0.2
  * (c) 2023 Eduardo San Martin Morote
  * @license MIT
  */const eo={escapeDeactivates:{type:Boolean,default:!0},returnFocusOnDeactivate:{type:Boolean,default:!0},allowOutsideClick:{type:[Boolean,Function],default:!0},clickOutsideDeactivates:[Boolean,Function],initialFocus:[String,Function,Boolean],fallbackFocus:[String,Function],checkCanFocusTrap:Function,checkCanReturnFocus:Function,delayInitialFocus:{type:Boolean,default:!0},document:Object,preventScroll:Boolean,setReturnFocus:[Object,String,Boolean,Function],tabbableOptions:Object};qe({props:Object.assign({active:{type:Boolean,default:!0}},eo),emits:["update:active","activate","postActivate","deactivate","postDeactivate"],render(){return this.renderImpl()},setup(t,{slots:n,emit:o}){let s;const l=or(null),i=Dt(()=>{const v=l.value;return v&&(v instanceof HTMLElement?v:v.$el)});function p(){return s||(s=qr(i.value,{escapeDeactivates:t.escapeDeactivates,allowOutsideClick:t.allowOutsideClick,returnFocusOnDeactivate:t.returnFocusOnDeactivate,clickOutsideDeactivates:t.clickOutsideDeactivates,onActivate:()=>{o("update:active",!0),o("activate")},onDeactivate:()=>{o("update:active",!1),o("deactivate")},onPostActivate:()=>o("postActivate"),onPostDeactivate:()=>o("postDeactivate"),initialFocus:t.initialFocus,fallbackFocus:t.fallbackFocus,tabbableOptions:t.tabbableOptions,delayInitialFocus:t.delayInitialFocus,preventScroll:t.preventScroll}))}return ar(()=>{ir(()=>t.active,v=>{v&&i.value?p().activate():s&&(s.deactivate(),(!i.value||i.value.nodeType===Node.COMMENT_NODE)&&(s=null))},{immediate:!0,flush:"post"})}),cr(()=>{s&&s.deactivate(),s=null}),{activate(){p().activate()},deactivate(){s&&s.deactivate()},renderImpl(){if(!n.default)return null;const v=n.default().filter(y=>y.type!==sr);return!v||!v.length||v.length>1?(console.error("[focus-trap-vue]: FocusTrap requires exactly one child."),v):ur(v[0],{ref:l})}}}});var Te=typeof globalThis<"u"?globalThis:typeof window<"u"?window:typeof global<"u"?global:typeof self<"u"?self:{},Ce={exports:{}};Ce.exports;(function(t,n){var o=200,s="__lodash_hash_undefined__",l=9007199254740991,i="[object Arguments]",p="[object Array]",v="[object Boolean]",y="[object Date]",j="[object Error]",T="[object Function]",A="[object GeneratorFunction]",O="[object Map]",P="[object Number]",g="[object Object]",U="[object Promise]",D="[object RegExp]",x="[object Set]",$="[object String]",W="[object Symbol]",X="[object WeakMap]",be="[object ArrayBuffer]",q="[object DataView]",we="[object Float32Array]",ie="[object Float64Array]",Y="[object Int8Array]",u="[object Int16Array]",c="[object Int32Array]",d="[object Uint8Array]",m="[object Uint8ClampedArray]",h="[object Uint16Array]",_="[object Uint32Array]",k=/[\\^$.*+?()[\]{}|]/g,R=/\w*$/,S=/^\[object .+?Constructor\]$/,Z=/^(?:0|[1-9]\d*)$/,b={};b[i]=b[p]=b[be]=b[q]=b[v]=b[y]=b[we]=b[ie]=b[Y]=b[u]=b[c]=b[O]=b[P]=b[g]=b[D]=b[x]=b[$]=b[W]=b[d]=b[m]=b[h]=b[_]=!0,b[j]=b[T]=b[X]=!1;var H=typeof Te=="object"&&Te&&Te.Object===Object&&Te,Pe=typeof self=="object"&&self&&self.Object===Object&&self,G=H||Pe||Function("return this")(),ce=n&&!n.nodeType&&n,se=ce&&!0&&t&&!t.nodeType&&t,Mt=se&&se.exports===ce;function $t(e,r){return e.set(r[0],r[1]),e}function Gt(e,r){return e.add(r),e}function Kt(e,r){for(var a=-1,f=e?e.length:0;++a<f&&r(e[a],a,e)!==!1;);return e}function Vt(e,r){for(var a=-1,f=r.length,N=e.length;++a<f;)e[N+a]=r[a];return e}function et(e,r,a,f){var N=-1,I=e?e.length:0;for(f&&I&&(a=e[++N]);++N<I;)a=r(a,e[N],N,e);return a}function Ut(e,r){for(var a=-1,f=Array(e);++a<e;)f[a]=r(a);return f}function Ht(e,r){return e==null?void 0:e[r]}function tt(e){var r=!1;if(e!=null&&typeof e.toString!="function")try{r=!!(e+"")}catch{}return r}function nt(e){var r=-1,a=Array(e.size);return e.forEach(function(f,N){a[++r]=[N,f]}),a}function Ie(e,r){return function(a){return e(r(a))}}function rt(e){var r=-1,a=Array(e.size);return e.forEach(function(f){a[++r]=f}),a}var zt=Array.prototype,Jt=Function.prototype,Oe=Object.prototype,Be=G["__core-js_shared__"],ot=function(){var e=/[^.]+$/.exec(Be&&Be.keys&&Be.keys.IE_PROTO||"");return e?"Symbol(src)_1."+e:""}(),at=Jt.toString,Q=Oe.hasOwnProperty,ke=Oe.toString,Wt=RegExp("^"+at.call(Q).replace(k,"\\$&").replace(/hasOwnProperty|(function).*?(?=\\\()| for .+?(?=\\\])/g,"$1.*?")+"$"),it=Mt?G.Buffer:void 0,ct=G.Symbol,st=G.Uint8Array,Xt=Ie(Object.getPrototypeOf,Object),Yt=Object.create,Zt=Oe.propertyIsEnumerable,Qt=zt.splice,ut=Object.getOwnPropertySymbols,qt=it?it.isBuffer:void 0,en=Ie(Object.keys,Object),Le=fe(G,"DataView"),he=fe(G,"Map"),Me=fe(G,"Promise"),$e=fe(G,"Set"),Ge=fe(G,"WeakMap"),me=fe(Object,"create"),tn=ne(Le),nn=ne(he),rn=ne(Me),on=ne($e),an=ne(Ge),lt=ct?ct.prototype:void 0,ft=lt?lt.valueOf:void 0;function ee(e){var r=-1,a=e?e.length:0;for(this.clear();++r<a;){var f=e[r];this.set(f[0],f[1])}}function cn(){this.__data__=me?me(null):{}}function sn(e){return this.has(e)&&delete this.__data__[e]}function un(e){var r=this.__data__;if(me){var a=r[e];return a===s?void 0:a}return Q.call(r,e)?r[e]:void 0}function ln(e){var r=this.__data__;return me?r[e]!==void 0:Q.call(r,e)}function fn(e,r){var a=this.__data__;return a[e]=me&&r===void 0?s:r,this}ee.prototype.clear=cn,ee.prototype.delete=sn,ee.prototype.get=un,ee.prototype.has=ln,ee.prototype.set=fn;function z(e){var r=-1,a=e?e.length:0;for(this.clear();++r<a;){var f=e[r];this.set(f[0],f[1])}}function dn(){this.__data__=[]}function pn(e){var r=this.__data__,a=Se(r,e);if(a<0)return!1;var f=r.length-1;return a==f?r.pop():Qt.call(r,a,1),!0}function vn(e){var r=this.__data__,a=Se(r,e);return a<0?void 0:r[a][1]}function bn(e){return Se(this.__data__,e)>-1}function hn(e,r){var a=this.__data__,f=Se(a,e);return f<0?a.push([e,r]):a[f][1]=r,this}z.prototype.clear=dn,z.prototype.delete=pn,z.prototype.get=vn,z.prototype.has=bn,z.prototype.set=hn;function ue(e){var r=-1,a=e?e.length:0;for(this.clear();++r<a;){var f=e[r];this.set(f[0],f[1])}}function mn(){this.__data__={hash:new ee,map:new(he||z),string:new ee}}function yn(e){return Ee(this,e).delete(e)}function gn(e){return Ee(this,e).get(e)}function _n(e){return Ee(this,e).has(e)}function wn(e,r){return Ee(this,e).set(e,r),this}ue.prototype.clear=mn,ue.prototype.delete=yn,ue.prototype.get=gn,ue.prototype.has=_n,ue.prototype.set=wn;function le(e){this.__data__=new z(e)}function On(){this.__data__=new z}function kn(e){return this.__data__.delete(e)}function Sn(e){return this.__data__.get(e)}function En(e){return this.__data__.has(e)}function jn(e,r){var a=this.__data__;if(a instanceof z){var f=a.__data__;if(!he||f.length<o-1)return f.push([e,r]),this;a=this.__data__=new ue(f)}return a.set(e,r),this}le.prototype.clear=On,le.prototype.delete=kn,le.prototype.get=Sn,le.prototype.has=En,le.prototype.set=jn;function Nn(e,r){var a=Ue(e)||Zn(e)?Ut(e.length,String):[],f=a.length,N=!!f;for(var I in e)(r||Q.call(e,I))&&!(N&&(I=="length"||Jn(I,f)))&&a.push(I);return a}function dt(e,r,a){var f=e[r];(!(Q.call(e,r)&&ht(f,a))||a===void 0&&!(r in e))&&(e[r]=a)}function Se(e,r){for(var a=e.length;a--;)if(ht(e[a][0],r))return a;return-1}function Fn(e,r){return e&&pt(r,He(r),e)}function Ke(e,r,a,f,N,I,K){var L;if(f&&(L=I?f(e,N,I,K):f(e)),L!==void 0)return L;if(!je(e))return e;var gt=Ue(e);if(gt){if(L=Un(e),!r)return Gn(e,L)}else{var de=te(e),_t=de==T||de==A;if(qn(e))return Cn(e,r);if(de==g||de==i||_t&&!I){if(tt(e))return I?e:{};if(L=Hn(_t?{}:e),!r)return Kn(e,Fn(L,e))}else{if(!b[de])return I?e:{};L=zn(e,de,Ke,r)}}K||(K=new le);var wt=K.get(e);if(wt)return wt;if(K.set(e,L),!gt)var Ot=a?Vn(e):He(e);return Kt(Ot||e,function(ze,Ne){Ot&&(Ne=ze,ze=e[Ne]),dt(L,Ne,Ke(ze,r,a,f,Ne,e,K))}),L}function Tn(e){return je(e)?Yt(e):{}}function An(e,r,a){var f=r(e);return Ue(e)?f:Vt(f,a(e))}function Dn(e){return ke.call(e)}function xn(e){if(!je(e)||Xn(e))return!1;var r=yt(e)||tt(e)?Wt:S;return r.test(ne(e))}function Rn(e){if(!bt(e))return en(e);var r=[];for(var a in Object(e))Q.call(e,a)&&a!="constructor"&&r.push(a);return r}function Cn(e,r){if(r)return e.slice();var a=new e.constructor(e.length);return e.copy(a),a}function Ve(e){var r=new e.constructor(e.byteLength);return new st(r).set(new st(e)),r}function Pn(e,r){var a=r?Ve(e.buffer):e.buffer;return new e.constructor(a,e.byteOffset,e.byteLength)}function In(e,r,a){var f=r?a(nt(e),!0):nt(e);return et(f,$t,new e.constructor)}function Bn(e){var r=new e.constructor(e.source,R.exec(e));return r.lastIndex=e.lastIndex,r}function Ln(e,r,a){var f=r?a(rt(e),!0):rt(e);return et(f,Gt,new e.constructor)}function Mn(e){return ft?Object(ft.call(e)):{}}function $n(e,r){var a=r?Ve(e.buffer):e.buffer;return new e.constructor(a,e.byteOffset,e.length)}function Gn(e,r){var a=-1,f=e.length;for(r||(r=Array(f));++a<f;)r[a]=e[a];return r}function pt(e,r,a,f){a||(a={});for(var N=-1,I=r.length;++N<I;){var K=r[N],L=f?f(a[K],e[K],K,a,e):void 0;dt(a,K,L===void 0?e[K]:L)}return a}function Kn(e,r){return pt(e,vt(e),r)}function Vn(e){return An(e,He,vt)}function Ee(e,r){var a=e.__data__;return Wn(r)?a[typeof r=="string"?"string":"hash"]:a.map}function fe(e,r){var a=Ht(e,r);return xn(a)?a:void 0}var vt=ut?Ie(ut,Object):nr,te=Dn;(Le&&te(new Le(new ArrayBuffer(1)))!=q||he&&te(new he)!=O||Me&&te(Me.resolve())!=U||$e&&te(new $e)!=x||Ge&&te(new Ge)!=X)&&(te=function(e){var r=ke.call(e),a=r==g?e.constructor:void 0,f=a?ne(a):void 0;if(f)switch(f){case tn:return q;case nn:return O;case rn:return U;case on:return x;case an:return X}return r});function Un(e){var r=e.length,a=e.constructor(r);return r&&typeof e[0]=="string"&&Q.call(e,"index")&&(a.index=e.index,a.input=e.input),a}function Hn(e){return typeof e.constructor=="function"&&!bt(e)?Tn(Xt(e)):{}}function zn(e,r,a,f){var N=e.constructor;switch(r){case be:return Ve(e);case v:case y:return new N(+e);case q:return Pn(e,f);case we:case ie:case Y:case u:case c:case d:case m:case h:case _:return $n(e,f);case O:return In(e,f,a);case P:case $:return new N(e);case D:return Bn(e);case x:return Ln(e,f,a);case W:return Mn(e)}}function Jn(e,r){return r=r??l,!!r&&(typeof e=="number"||Z.test(e))&&e>-1&&e%1==0&&e<r}function Wn(e){var r=typeof e;return r=="string"||r=="number"||r=="symbol"||r=="boolean"?e!=="__proto__":e===null}function Xn(e){return!!ot&&ot in e}function bt(e){var r=e&&e.constructor,a=typeof r=="function"&&r.prototype||Oe;return e===a}function ne(e){if(e!=null){try{return at.call(e)}catch{}try{return e+""}catch{}}return""}function Yn(e){return Ke(e,!0,!0)}function ht(e,r){return e===r||e!==e&&r!==r}function Zn(e){return Qn(e)&&Q.call(e,"callee")&&(!Zt.call(e,"callee")||ke.call(e)==i)}var Ue=Array.isArray;function mt(e){return e!=null&&er(e.length)&&!yt(e)}function Qn(e){return tr(e)&&mt(e)}var qn=qt||rr;function yt(e){var r=je(e)?ke.call(e):"";return r==T||r==A}function er(e){return typeof e=="number"&&e>-1&&e%1==0&&e<=l}function je(e){var r=typeof e;return!!e&&(r=="object"||r=="function")}function tr(e){return!!e&&typeof e=="object"}function He(e){return mt(e)?Nn(e):Rn(e)}function nr(){return[]}function rr(){return!1}t.exports=Yn})(Ce,Ce.exports);Ce.exports;const to=["innerHTML"],no=["innerHTML"],ro={class:"stack"},oo={class:"columns"},ao={class:"stack"},io={class:"columns"},co=qe({__name:"MeshDetailView",props:{mesh:{}},setup(t){const n=t;return(o,s)=>{const l=J("RouteTitle"),i=J("XAction"),p=J("KBadge"),v=J("KCard"),y=J("DataSource"),j=J("AppView"),T=J("RouteView");return w(),ve(T,{name:"mesh-detail-view",params:{mesh:""}},{default:E(({route:A,t:O,uri:P})=>[M(l,{title:O("meshes.routes.overview.title"),render:!1},null,8,["title"]),C(),M(y,{src:P(We(kt),"/mesh-insights/:name",{name:A.params.mesh})},{default:E(({data:g})=>[(w(!0),F(ye,null,Xe([["MeshTrafficPermission","TrafficPermission"].reduce((U,D)=>{var x,$;return U+((($=(x=g==null?void 0:g.policies)==null?void 0:x[D])==null?void 0:$.total)??0)},0)===0],U=>(w(),ve(j,{key:U},xt({default:E(()=>[C(),re("div",ro,[M(We(Nr),{title:O("meshes.routes.item.subtitle",{name:n.mesh.name}),created:O("common.formats.datetime",{value:Date.parse(n.mesh.creationTime)}),modified:O("common.formats.datetime",{value:Date.parse(n.mesh.modificationTime)})},{default:E(()=>[re("div",oo,[(w(),F(ye,null,Xe(["MeshTrafficPermission","MeshMetric","MeshAccessLog","MeshTrace"],D=>(w(),F(ye,{key:D},[(w(!0),F(ye,null,Xe([Object.entries((g==null?void 0:g.policies)??{}).find(([x,$])=>x===D)],x=>(w(),ve(St,{key:x},{title:E(()=>[M(i,{to:{name:"policy-list-view",params:{mesh:A.params.mesh,policyPath:`${D.toLowerCase()}s`}}},{default:E(()=>[C(B(D),1)]),_:2},1032,["to"])]),body:E(()=>[M(p,{appearance:"neutral"},{default:E(()=>[C(B(O(x?"meshes.detail.enabled":"meshes.detail.disabled")),1)]),_:2},1024)]),_:2},1024))),128))],64))),64))])]),_:2},1032,["title","created","modified"]),C(),M(v,null,{default:E(()=>[re("div",ao,[re("div",io,[M(Ye,{total:(g==null?void 0:g.services.total)??0,"data-testid":"services-status"},{title:E(()=>[C(B(O("meshes.detail.services")),1)]),_:2},1032,["total"]),C(),M(Ye,{total:(g==null?void 0:g.dataplanesByType.standard.total)??0,online:(g==null?void 0:g.dataplanesByType.standard.online)??0,"data-testid":"data-plane-proxies-status"},{title:E(()=>[C(B(O("meshes.detail.data_plane_proxies")),1)]),_:2},1032,["total","online"]),C(),M(Ye,{total:(g==null?void 0:g.totalPolicyCount)??0,"data-testid":"policies-status"},{title:E(()=>[C(B(O("meshes.detail.policies")),1)]),_:2},1032,["total"]),C(),M(St,null,{title:E(()=>[C(B(O("http.api.property.mtls")),1)]),body:E(()=>[n.mesh.mtlsBackend?(w(),F(ye,{key:1},[C(B(n.mesh.mtlsBackend.type)+" / "+B(n.mesh.mtlsBackend.name),1)],64)):(w(),ve(p,{key:0,appearance:"neutral"},{default:E(()=>[C(B(O("meshes.detail.disabled")),1)]),_:2},1024))]),_:2},1024)])])]),_:2},1024),C(),M(dr,{resource:o.mesh.config},{default:E(({copy:D,copying:x})=>[x?(w(),ve(y,{key:0,src:P(We(kt),"/meshes/:name/as/kubernetes",{name:A.params.mesh},{cacheControl:"no-store"}),onChange:$=>{D(W=>W($))},onError:$=>{D((W,X)=>X($))}},null,8,["src","onChange","onError"])):V("",!0)]),_:2},1032,["resource"])])]),_:2},[!n.mesh.mtlsBackend||U?{name:"notifications",fn:E(()=>[re("ul",null,[n.mesh.mtlsBackend?V("",!0):(w(),F("li",{key:0,innerHTML:O("meshes.routes.item.mtls-warning")},null,8,to)),C(),n.mesh.mtlsBackend&&U?(w(),F("li",{key:1,innerHTML:O("meshes.routes.item.mtp-warning")},null,8,no)):V("",!0)])]),key:"0"}:void 0]),1024))),128))]),_:2},1032,["src"])]),_:1})}}}),po=fr(co,[["__scopeId","data-v-cfd653fe"]]);export{po as default};
