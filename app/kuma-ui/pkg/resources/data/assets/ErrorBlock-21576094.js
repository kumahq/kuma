import{d as g,cu as d,g as h,o as a,c as t,a as i,ch as y,u as r,w as n,e as o,b as s,bV as c,j as u,F as v,cd as E,m as w,$ as B,i as x,c3 as k,bX as S,bY as V,k as C}from"./index-08ba2993.js";const b=e=>(S("data-v-9de3b600"),e=e(),V(),e),I={class:"error-block"},N=b(()=>o("p",null,"An error has occurred while trying to load this data.",-1)),j={class:"error-block-details"},A=b(()=>o("summary",null,"Details",-1)),D={key:0},F={key:0,class:"badge-list"},$=g({__name:"ErrorBlock",props:{error:{type:[Error,d],required:!1,default:null}},setup(e){const l=e,_=h(()=>l.error instanceof Error),m=h(()=>l.error instanceof d?l.error.causes:[]);return(q,z)=>(a(),t("div",I,[i(r(B),{"cta-is-hidden":""},y({title:n(()=>[i(r(w),{class:"mb-3",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"42"}),s(),N]),_:2},[r(_)||r(m).length>0?{name:"message",fn:n(()=>[o("details",j,[A,s(),r(_)?(a(),t("p",D,c(e.error.message),1)):u("",!0),s(),o("ul",null,[(a(!0),t(v,null,E(r(m),(p,f)=>(a(),t("li",{key:f},[o("b",null,[o("code",null,c(p.field),1)]),s(": "+c(p.message),1)]))),128))])])]),key:"0"}:void 0]),1024),s(),e.error instanceof r(d)?(a(),t("div",F,[e.error.code?(a(),x(r(k),{key:0,appearance:"warning"},{default:n(()=>[s(c(e.error.code),1)]),_:1})):u("",!0),s(),i(r(k),{appearance:"warning"},{default:n(()=>[s(c(e.error.statusCode),1)]),_:1})])):u("",!0)]))}});const O=C($,[["__scopeId","data-v-9de3b600"]]);export{O as E};
