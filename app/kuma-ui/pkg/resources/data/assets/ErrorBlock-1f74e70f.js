import{c as k,q as g,O as _}from"./kongponents.es-6cc20401.js";import{d as y,c as v,an as f,j as s,g as l,E as w,w as n,e as t,h as r,b as E,f as d,o as a,i as o,t as c,F as b,q as B,p as x,m as S}from"./index-c271a676.js";import{_ as q}from"./_plugin-vue_export-helper-c27b6911.js";const p=e=>(x("data-v-4154f15d"),e=e(),S(),e),C={class:"error-block"},I=p(()=>o("p",null,"An error has occurred while trying to load this data.",-1)),N={class:"error-block-details"},V=p(()=>o("summary",null,"Details",-1)),A={key:0},D={key:0,class:"badge-list"},F=y({__name:"ErrorBlock",props:{error:{type:[Error,null],required:!1,default:null}},setup(e){const i=e,u=v(()=>i.error instanceof f?i.error.causes:[]);return(O,j)=>(a(),s("div",C,[l(t(g),{"cta-is-hidden":""},w({title:n(()=>[l(t(k),{class:"mb-3",icon:"warning",color:"var(--black-500)","secondary-color":"var(--yellow-300)",size:"42"}),r(),I]),_:2},[e.error!==null||u.value.length>0?{name:"message",fn:n(()=>[o("details",N,[V,r(),e.error!==null?(a(),s("p",A,c(e.error.message),1)):d("",!0),r(),o("ul",null,[(a(!0),s(b,null,B(u.value,(m,h)=>(a(),s("li",{key:h},[o("b",null,[o("code",null,c(m.field),1)]),r(": "+c(m.message),1)]))),128))])])]),key:"0"}:void 0]),1024),r(),e.error instanceof t(f)?(a(),s("div",D,[e.error.code?(a(),E(t(_),{key:0,appearance:"warning"},{default:n(()=>[r(c(e.error.code),1)]),_:1})):d("",!0),r(),l(t(_),{appearance:"warning"},{default:n(()=>[r(c(e.error.statusCode),1)]),_:1})])):d("",!0)]))}});const G=q(F,[["__scopeId","data-v-4154f15d"]]);export{G as E};
