import{ah as E,l as k,ai as x,a7 as O,d as P,A as C,a as L,o as T,b as I,w as D,r as W,_ as j}from"./index-fb2eded6.js";function V(e){return E()?(x(e),!0):!1}function y(e){return typeof e=="function"?e():k(e)}const S=typeof window<"u"&&typeof document<"u";typeof WorkerGlobalScope<"u"&&globalThis instanceof WorkerGlobalScope;const F=Object.prototype.toString,M=e=>F.call(e)==="[object Object]",w=()=>{},R=$();function $(){var e,n;return S&&((e=window==null?void 0:window.navigator)==null?void 0:e.userAgent)&&(/iP(ad|hone|od)/.test(window.navigator.userAgent)||((n=window==null?void 0:window.navigator)==null?void 0:n.maxTouchPoints)>2&&/iPad|Macintosh/.test(window==null?void 0:window.navigator.userAgent))}function B(e,n){function r(...o){return new Promise((a,i)=>{Promise.resolve(e(()=>n.apply(this,o),{fn:n,thisArg:this,args:o})).then(a).catch(i)})}return r}function G(e,n=!0,r=!0,o=!1){let a=0,i,u=!0,c=w,f;const m=()=>{i&&(clearTimeout(i),i=void 0,c(),c=w)};return p=>{const t=y(e),s=Date.now()-a,l=()=>f=p();return m(),t<=0?(a=Date.now(),l()):(s>t&&(r||!u)?(a=Date.now(),l()):n&&(f=new Promise((_,b)=>{c=o?b:_,i=setTimeout(()=>{a=Date.now(),u=!0,_(l()),m()},Math.max(0,t-s))})),!r&&!i&&(i=setTimeout(()=>u=!0,t)),u=!1,f)}}function K(e,n=200,r=!1,o=!0,a=!1){return B(G(n,r,o,a),e)}function h(e){var n;const r=y(e);return(n=r==null?void 0:r.$el)!=null?n:r}const A=S?window:void 0;function v(...e){let n,r,o,a;if(typeof e[0]=="string"||Array.isArray(e[0])?([r,o,a]=e,n=A):[n,r,o,a]=e,!n)return w;Array.isArray(r)||(r=[r]),Array.isArray(o)||(o=[o]);const i=[],u=()=>{i.forEach(d=>d()),i.length=0},c=(d,p,t,s)=>(d.addEventListener(p,t,s),()=>d.removeEventListener(p,t,s)),f=O(()=>[h(n),y(a)],([d,p])=>{if(u(),!d)return;const t=M(p)?{...p}:p;i.push(...r.flatMap(s=>o.map(l=>c(d,s,l,t))))},{immediate:!0,flush:"post"}),m=()=>{f(),u()};return V(m),m}let g=!1;function N(e,n,r={}){const{window:o=A,ignore:a=[],capture:i=!0,detectIframe:u=!1}=r;if(!o)return w;R&&!g&&(g=!0,Array.from(o.document.body.children).forEach(t=>t.addEventListener("click",w)),o.document.documentElement.addEventListener("click",w));let c=!0;const f=t=>a.some(s=>{if(typeof s=="string")return Array.from(o.document.querySelectorAll(s)).some(l=>l===t.target||t.composedPath().includes(l));{const l=h(s);return l&&(t.target===l||t.composedPath().includes(l))}}),d=[v(o,"click",t=>{const s=h(e);if(!(!s||s===t.target||t.composedPath().includes(s))){if(t.detail===0&&(c=!f(t)),!c){c=!0;return}n(t)}},{passive:!0,capture:i}),v(o,"pointerdown",t=>{const s=h(e);c=!f(t)&&!!(s&&!t.composedPath().includes(s))},{passive:!0}),u&&v(o,"blur",t=>{setTimeout(()=>{var s;const l=h(e);((s=o.document.activeElement)==null?void 0:s.tagName)==="IFRAME"&&!(l!=null&&l.contains(o.document.activeElement))&&n(t)},0)})].filter(Boolean);return()=>d.forEach(t=>t())}const q=P({__name:"SummaryView",props:{width:{default:"560px"}},emits:["close"],setup(e,{emit:n}){const r=C(null);N(r,K(i=>{const u=i.target;i.isTrusted&&u.nodeName.toLowerCase()!=="a"&&a("close")},1,!0,!1));const o=e,a=n;return(i,u)=>{const c=L("KSlideout");return T(),I(c,{ref_key:"slideOutRef",ref:r,class:"summary-slideout","prevent-close-on-blur":"","close-button-alignment":"end","has-overlay":!1,"is-visible":"","max-width":o.width,"offset-top":"var(--app-slideout-offset-top, 0)","data-testid":"summary",onClose:u[0]||(u[0]=f=>a("close"))},{default:D(()=>[W(i.$slots,"default",{},void 0,!0)]),_:3},8,["max-width"])}}});const H=j(q,[["__scopeId","data-v-3a795769"]]);export{H as S};
