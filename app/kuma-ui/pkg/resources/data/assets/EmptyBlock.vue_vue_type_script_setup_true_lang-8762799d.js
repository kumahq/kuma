import{_ as c}from"./_plugin-vue_export-helper-c27b6911.js";import{o as i,j as _,J as s,d,c as u,i as n,t as f,b as o,h as l,e as h,H as g,w as r,g as v}from"./index-8b82889e.js";import{d as y,a as b}from"./kongponents.es-a6650bd6.js";const L={},$={class:"definition-list"};function D(t,e){return i(),_("dl",$,[s(t.$slots,"default",{},void 0,!0)])}const A=c(L,[["render",D],["__scopeId","data-v-48665ce3"]]);function k(t){const e=t.split(/([A-Z][a-z]+)/).join(" ").replace(/\s+/g," ").trim();return e.charAt(0).toUpperCase()+e.substring(1)}function I(t,e){return e.termLabels[t]??k(t)}function w(){return{termLabels:{mtls:"mTLS"}}}const B={class:"definition-list-item"},S={class:"definition-list-item__term"},T={class:"definition-list-item__details"},x=d({__name:"DefinitionListItem",props:{term:{type:String,required:!0}},setup(t){const e=t,a=w(),m=u(()=>I(e.term,a));return(p,N)=>(i(),_("div",B,[n("dt",S,f(o(m)),1),l(),n("dd",T,[s(p.$slots,"default",{},void 0,!0)])]))}});const E=c(x,[["__scopeId","data-v-81c62ba5"]]),C=n("p",null,"There is no data to display.",-1),q=d({__name:"EmptyBlock",setup(t){return(e,a)=>(i(),h(o(b),{"cta-is-hidden":""},g({title:r(()=>[s(e.$slots,"title",{},()=>[v(o(y),{class:"mb-3",icon:"warning",color:"var(--black-500)","secondary-color":"var(--yellow-300)",size:"42"}),l(),C])]),_:2},[e.$slots.message?{name:"message",fn:r(()=>[s(e.$slots,"message")]),key:"0"}:void 0]),1024))}});export{E as D,q as _,A as a};
