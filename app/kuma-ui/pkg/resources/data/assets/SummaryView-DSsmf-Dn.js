import{Y as h,Z as A,M as I,$ as g,a0 as m,a1 as M,a2 as L,a3 as R,p as $,a4 as V,d as N,a5 as B,a6 as F,y as W,a7 as q,r as _,o as K,m as X,w as O,b as j,e as D,a as Y,_ as Z}from"./index-Bi3CXAeE.js";const x=L?window:void 0;function w(n){var u;const a=m(n);return(u=a==null?void 0:a.$el)!=null?u:a}function y(...n){const u=[],a=()=>{u.forEach(o=>o()),u.length=0},r=(o,t,l,d)=>(o.addEventListener(t,l,d),()=>o.removeEventListener(t,l,d)),f=I(()=>{const o=g(m(n[0])).filter(t=>t!=null);return o.every(t=>typeof t!="string")?o:void 0}),p=M(()=>{var o,t;return[(t=(o=f.value)==null?void 0:o.map(l=>w(l)))!=null?t:[x].filter(l=>l!=null),g(m(f.value?n[1]:n[0])),g($(f.value?n[2]:n[1])),m(f.value?n[3]:n[2])]},([o,t,l,d])=>{if(a(),!(o!=null&&o.length)||!(t!=null&&t.length)||!(l!=null&&l.length))return;const b=R(d)?{...d}:d;u.push(...o.flatMap(S=>t.flatMap(v=>l.map(E=>r(S,v,E,b)))))},{flush:"post"}),c=()=>{p(),a()};return V(a),c}let P=!1;function z(n,u,a={}){const{window:r=x,ignore:f=[],capture:p=!0,detectIframe:c=!1,controls:o=!1}=a;if(!r)return o?{stop:h,cancel:h,trigger:h}:h;if(A&&!P){P=!0;const e={passive:!0};Array.from(r.document.body.children).forEach(s=>y(s,"click",h,e)),y(r.document.documentElement,"click",h,e)}let t=!0;const l=e=>m(f).some(s=>{if(typeof s=="string")return Array.from(r.document.querySelectorAll(s)).some(i=>i===e.target||e.composedPath().includes(i));{const i=w(s);return i&&(e.target===i||e.composedPath().includes(i))}});function d(e){const s=m(e);return s&&s.$.subTree.shapeFlag===16}function b(e,s){const i=m(e),T=i.$.subTree&&i.$.subTree.children;return T==null||!Array.isArray(T)?!1:T.some(C=>C.el===s.target||s.composedPath().includes(C.el))}const S=e=>{const s=w(n);if(e.target!=null&&!(!(s instanceof Element)&&d(n)&&b(n,e))&&!(!s||s===e.target||e.composedPath().includes(s))){if("detail"in e&&e.detail===0&&(t=!l(e)),!t){t=!0;return}u(e)}};let v=!1;const E=[y(r,"click",e=>{v||(v=!0,setTimeout(()=>{v=!1},0),S(e))},{passive:!0,capture:p}),y(r,"pointerdown",e=>{const s=w(n);t=!l(e)&&!!(s&&!e.composedPath().includes(s))},{passive:!0}),c&&y(r,"blur",e=>{setTimeout(()=>{var s;const i=w(n);((s=r.document.activeElement)==null?void 0:s.tagName)==="IFRAME"&&!(i!=null&&i.contains(r.document.activeElement))&&u(e)},0)},{passive:!0})].filter(Boolean),k=()=>E.forEach(e=>e());return o?{stop:k,cancel:()=>{t=!1},trigger:e=>{t=!0,S(e),t=!1}}:k}const G=N({__name:"SummaryView",props:{width:{default:"560px"}},emits:["close"],setup(n,{emit:u}){const a=B("summary-view-title");F("app-summary-view",a);const r=W(null);z(r,q(c=>{var t;const o=c.target;(((t=window.getSelection())==null?void 0:t.isCollapsed)??!0)&&!c.defaultPrevented&&c.isTrusted&&o.nodeName.toLowerCase()!=="a"&&p("close")},1,!0,!1));const f=n,p=u;return(c,o)=>{const t=_("XTeleportSlot"),l=_("KSlideout");return K(),X(l,{ref_key:"slideOutRef",ref:r,class:"summary-slideout","close-on-blur":!1,"has-overlay":!1,visible:"","max-width":f.width,"offset-top":"var(--app-slideout-offset-top, 0)","data-testid":"summary",onClose:o[0]||(o[0]=d=>p("close"))},{title:O(()=>[j(t,{name:$(a)},null,8,["name"])]),default:O(()=>[o[1]||(o[1]=D()),Y(c.$slots,"default",{},void 0,!0)]),_:3},8,["max-width"])}}}),J=Z(G,[["__scopeId","data-v-0b1b5d96"]]);export{J as S};
