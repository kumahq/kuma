import{a3 as _,a4 as L,a5 as O,a6 as P,a7 as v,a8 as $,a9 as C,aa as I,d as R,ab as V,ac as M,v as N,ad as B,r as T,o as F,p as W,w as A,b as q,m as K,e as X,a as j,_ as D}from"./index-COF03--z.js";const x=$?window:void 0;function y(o){var a;const n=v(o);return(a=n==null?void 0:n.$el)!=null?a:n}function S(...o){let a,n,s,p;if(typeof o[0]=="string"||Array.isArray(o[0])?([n,s,p]=o,a=x):[a,n,s,p]=o,!a)return _;n=C(n),s=C(s);const d=[],i=()=>{d.forEach(u=>u()),d.length=0},r=(u,f,m,w)=>(u.addEventListener(f,m,w),()=>u.removeEventListener(f,m,w)),c=O(()=>[y(a),v(p)],([u,f])=>{if(i(),!u)return;const m=P(f)?{...f}:f;d.push(...n.flatMap(w=>s.map(b=>r(u,w,b,m))))},{immediate:!0,flush:"post"}),h=()=>{c(),i()};return I(h),h}let g=!1;function z(o,a,n={}){const{window:s=x,ignore:p=[],capture:d=!0,detectIframe:i=!1}=n;if(!s)return _;L&&!g&&(g=!0,Array.from(s.document.body.children).forEach(e=>e.addEventListener("click",_)),s.document.documentElement.addEventListener("click",_));let r=!0;const c=e=>v(p).some(t=>{if(typeof t=="string")return Array.from(s.document.querySelectorAll(t)).some(l=>l===e.target||e.composedPath().includes(l));{const l=y(t);return l&&(e.target===l||e.composedPath().includes(l))}});function h(e){const t=v(e);return t&&t.$.subTree.shapeFlag===16}function u(e,t){const l=v(e),E=l.$.subTree&&l.$.subTree.children;return E==null||!Array.isArray(E)?!1:E.some(k=>k.el===t.target||t.composedPath().includes(k.el))}const f=e=>{const t=y(o);if(e.target!=null&&!(!(t instanceof Element)&&h(o)&&u(o,e))&&!(!t||t===e.target||e.composedPath().includes(t))){if(e.detail===0&&(r=!c(e)),!r){r=!0;return}a(e)}};let m=!1;const w=[S(s,"click",e=>{m||(m=!0,setTimeout(()=>{m=!1},0),f(e))},{passive:!0,capture:d}),S(s,"pointerdown",e=>{const t=y(o);r=!c(e)&&!!(t&&!e.composedPath().includes(t))},{passive:!0}),i&&S(s,"blur",e=>{setTimeout(()=>{var t;const l=y(o);((t=s.document.activeElement)==null?void 0:t.tagName)==="IFRAME"&&!(l!=null&&l.contains(s.document.activeElement))&&a(e)},0)})].filter(Boolean);return()=>w.forEach(e=>e())}const G=R({__name:"SummaryView",props:{width:{default:"560px"}},emits:["close"],setup(o,{emit:a}){const n=V("summary-view-title");M("app-summary-view",n);const s=N(null);z(s,B(i=>{var c;const r=i.target;(((c=window.getSelection())==null?void 0:c.isCollapsed)??!0)&&!i.defaultPrevented&&i.isTrusted&&r.nodeName.toLowerCase()!=="a"&&d("close")},1,!0,!1));const p=o,d=a;return(i,r)=>{const c=T("XTeleportSlot"),h=T("KSlideout");return F(),W(h,{ref_key:"slideOutRef",ref:s,class:"summary-slideout","close-on-blur":!1,"has-overlay":!1,visible:"","max-width":p.width,"offset-top":"var(--app-slideout-offset-top, 0)","data-testid":"summary",onClose:r[0]||(r[0]=u=>d("close"))},{title:A(()=>[q(c,{name:K(n)},null,8,["name"])]),default:A(()=>[r[1]||(r[1]=X()),j(i.$slots,"default",{},void 0,!0)]),_:3},8,["max-width"])}}}),J=D(G,[["__scopeId","data-v-0b1b5d96"]]);export{J as S};
