import{d as k,N as n,j as g,e as s,a as d,G as h,m as a,w as l,c as v,n as p,o as r,f as o,F as i,h as w,t,b as c,C as B,M as E,O as y,x,y as b,_ as S}from"./index.04875eef.js";const C=e=>(x("data-v-7a254205"),e=e(),b(),e),I={class:"error-block"},N={class:"card-icon mb-3"},V=C(()=>o("summary",null,"Details",-1)),K={key:0,class:"badge-list"},A=k({__name:"ErrorBlock",props:{error:{type:[Error,n],required:!1,default:null}},setup(e){const u=e,_=g(()=>u.error instanceof n?u.error.causes:[]);return(D,F)=>(r(),s("div",I,[d(a(E),{"cta-is-hidden":""},h({title:l(()=>[o("div",N,[d(a(B),{class:"kong-icon--centered",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"42"})]),o("p",null,[e.error instanceof a(n)?(r(),s(i,{key:0},[c(t(e.error.message),1)],64)):(r(),s(i,{key:1},[c(" An error has occurred while trying to load this data. ")],64))])]),_:2},[a(_).length>0?{name:"message",fn:l(()=>[o("details",null,[V,o("ul",null,[(r(!0),s(i,null,w(a(_),(m,f)=>(r(),s("li",{key:f},[o("b",null,[o("code",null,t(m.field),1)]),c(": "+t(m.message),1)]))),128))])])]),key:"0"}:void 0]),1024),e.error instanceof a(n)?(r(),s("div",K,[e.error.code?(r(),v(a(y),{key:0,appearance:"warning"},{default:l(()=>[c(t(e.error.code),1)]),_:1})):p("",!0),d(a(y),{appearance:"warning"},{default:l(()=>[c(t(e.error.statusCode),1)]),_:1})])):p("",!0)]))}});const q=S(A,[["__scopeId","data-v-7a254205"]]);export{q as E};
