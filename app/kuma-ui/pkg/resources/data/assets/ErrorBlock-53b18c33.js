import{m as k,E as g,_}from"./kongponents.es-3ba46133.js";import{A as p}from"./production-554ae9d4.js";import{d as y,c as v,o as a,h as t,e as n,a4 as E,u as o,w as l,g as s,f as r,t as c,b as d,F as w,l as b,a as B,p as x,k as S}from"./runtime-dom.esm-bundler-9284044f.js";import{_ as C}from"./_plugin-vue_export-helper-c27b6911.js";const f=e=>(x("data-v-cdd36506"),e=e(),S(),e),I={class:"error-block"},N=f(()=>s("p",null,"An error has occurred while trying to load this data.",-1)),V={class:"error-block-details"},A=f(()=>s("summary",null,"Details",-1)),D={key:0},F={key:0,class:"badge-list"},q=y({__name:"ErrorBlock",props:{error:{type:[Error,null],required:!1,default:null}},setup(e){const i=e,u=v(()=>i.error instanceof p?i.error.causes:[]);return(z,L)=>(a(),t("div",I,[n(o(g),{"cta-is-hidden":""},E({title:l(()=>[n(o(k),{class:"mb-3",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"42"}),r(),N]),_:2},[e.error!==null||o(u).length>0?{name:"message",fn:l(()=>[s("details",V,[A,r(),e.error!==null?(a(),t("p",D,c(e.error.message),1)):d("",!0),r(),s("ul",null,[(a(!0),t(w,null,b(o(u),(m,h)=>(a(),t("li",{key:h},[s("b",null,[s("code",null,c(m.field),1)]),r(": "+c(m.message),1)]))),128))])])]),key:"0"}:void 0]),1024),r(),e.error instanceof o(p)?(a(),t("div",F,[e.error.code?(a(),B(o(_),{key:0,appearance:"warning"},{default:l(()=>[r(c(e.error.code),1)]),_:1})):d("",!0),r(),n(o(_),{appearance:"warning"},{default:l(()=>[r(c(e.error.statusCode),1)]),_:1})])):d("",!0)]))}});const J=C(q,[["__scopeId","data-v-cdd36506"]]);export{J as E};
