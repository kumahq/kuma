import{af as y,ag as k,V as C,ah as x,ai as S,aj as L,ak as O,d as b,ad as g,al as I,v as T,am as V,r as v,o as N,m as P,w as _,b as $,l as q,e as B,a as R,q as W}from"./index-sgqUZBhH.js";function h(n){var r;const s=S(n);return(r=s==null?void 0:s.$el)!=null?r:s}const A=L?window:void 0;function w(...n){let r,s,t,d;if(typeof n[0]=="string"||Array.isArray(n[0])?([s,t,d]=n,r=A):[r,s,t,d]=n,!r)return y;Array.isArray(s)||(s=[s]),Array.isArray(t)||(t=[t]);const c=[],u=()=>{c.forEach(l=>l()),c.length=0},a=(l,m,e,o)=>(l.addEventListener(m,e,o),()=>l.removeEventListener(m,e,o)),f=C(()=>[h(r),S(d)],([l,m])=>{if(u(),!l)return;const e=x(m)?{...m}:m;c.push(...s.flatMap(o=>t.map(i=>a(l,o,i,e))))},{immediate:!0,flush:"post"}),p=()=>{f(),u()};return O(p),p}let E=!1;function j(n,r,s={}){const{window:t=A,ignore:d=[],capture:c=!0,detectIframe:u=!1}=s;if(!t)return y;k&&!E&&(E=!0,Array.from(t.document.body.children).forEach(e=>e.addEventListener("click",y)),t.document.documentElement.addEventListener("click",y));let a=!0;const f=e=>d.some(o=>{if(typeof o=="string")return Array.from(t.document.querySelectorAll(o)).some(i=>i===e.target||e.composedPath().includes(i));{const i=h(o);return i&&(e.target===i||e.composedPath().includes(i))}}),l=[w(t,"click",e=>{const o=h(n);if(!(!o||o===e.target||e.composedPath().includes(o))){if(e.detail===0&&(a=!f(e)),!a){a=!0;return}r(e)}},{passive:!0,capture:c}),w(t,"pointerdown",e=>{const o=h(n);a=!f(e)&&!!(o&&!e.composedPath().includes(o))},{passive:!0}),u&&w(t,"blur",e=>{setTimeout(()=>{var o;const i=h(n);((o=t.document.activeElement)==null?void 0:o.tagName)==="IFRAME"&&!(i!=null&&i.contains(t.document.activeElement))&&r(e)},0)})].filter(Boolean);return()=>l.forEach(e=>e())}const F=b({__name:"SummaryView",props:{width:{default:"560px"}},emits:["close"],setup(n,{emit:r}){const s=g("summary-view-title");I("app-summary-view",s);const t=T(null);j(t,V(u=>{const a=u.target;u.isTrusted&&a.nodeName.toLowerCase()!=="a"&&c("close")},1,!0,!1));const d=n,c=r;return(u,a)=>{const f=v("XTeleportSlot"),p=v("KSlideout");return N(),P(p,{ref_key:"slideOutRef",ref:t,class:"summary-slideout","close-on-blur":!1,"has-overlay":!1,visible:"","max-width":d.width,"offset-top":"var(--app-slideout-offset-top, 0)","data-testid":"summary",onClose:a[0]||(a[0]=l=>c("close"))},{title:_(()=>[$(f,{name:q(s)},null,8,["name"])]),default:_(()=>[B(),R(u.$slots,"default",{},void 0,!0)]),_:3},8,["max-width"])}}}),M=W(F,[["__scopeId","data-v-1eac95d3"]]);export{M as S};
