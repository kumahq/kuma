import{l as y,Q as b,P as _}from"./kongponents.es-7e228e6a.js";import{d as g,m as v,ax as p,o as a,f as s,a as n,x as w,w as c,e as o,b as r,t,g as d,F as x,k as B,u as l,y as E,c as S,p as C,j as I}from"./index-2aa994fe.js";import{_ as N}from"./RouteView.vue_vue_type_script_setup_true_lang-49cffe7a.js";const f=e=>(C("data-v-176c6beb"),e=e(),I(),e),V={class:"error-block"},A=f(()=>o("p",null,"An error has occurred while trying to load this data.",-1)),D={class:"error-block-details"},F=f(()=>o("summary",null,"Details",-1)),P={key:0},Q={key:0,class:"badge-list"},j=g({__name:"ErrorBlock",props:{error:{type:[Error,null],required:!1,default:null}},setup(e){const i=e,u=v(()=>i.error instanceof p?i.error.causes:[]);return(h,q)=>(a(),s("div",V,[n(l(b),{"cta-is-hidden":""},w({title:c(()=>[n(l(y),{class:"mb-3",icon:"warning",color:"var(--black-500)","secondary-color":"var(--yellow-300)",size:"42"}),r(),E(h.$slots,"default",{},()=>[A],!0)]),_:2},[e.error!==null||u.value.length>0?{name:"message",fn:c(()=>[o("details",D,[F,r(),e.error!==null?(a(),s("p",P,t(e.error.message),1)):d("",!0),r(),o("ul",null,[(a(!0),s(x,null,B(u.value,(m,k)=>(a(),s("li",{key:k},[o("b",null,[o("code",null,t(m.field),1)]),r(": "+t(m.message),1)]))),128))])])]),key:"0"}:void 0]),1024),r(),e.error instanceof l(p)?(a(),s("div",Q,[e.error.code?(a(),S(l(_),{key:0,appearance:"warning"},{default:c(()=>[r(t(e.error.code),1)]),_:1})):d("",!0),r(),n(l(_),{appearance:"warning"},{default:c(()=>[r(t(e.error.statusCode),1)]),_:1})])):d("",!0)]))}});const $=N(j,[["__scopeId","data-v-176c6beb"]]);export{$ as E};
