import{s as u,q as k}from"./kongponents.es-e4c2d233.js";import{d as g,r as s,k as b,Q as w,o,j as B,i as m,g as C,w as y,b as d,e as l,h as E,z as n}from"./index-ae17c915.js";const $={class:"scanner"},q={class:"scanner-content"},N={class:"mb-2"},S=g({__name:"EntityScanner",props:{interval:{type:Number,required:!1,default:1e3},retries:{type:Number,required:!1,default:3600},hasError:{type:Boolean,default:!1},loaderFunction:{type:Function,required:!0},canComplete:{type:Boolean,default:!1}},emits:["hide-siblings"],setup(a,{emit:p}){const r=a,f=s(0),t=s(!1),v=s(!1),i=s(null);b(function(){h()}),w(function(){c()});function h(){t.value=!0,v.value=!1,c(),i.value=window.setInterval(async()=>{f.value++,await r.loaderFunction(),(f.value===r.retries||r.canComplete===!0)&&(c(),t.value=!1,v.value=!0,p("hide-siblings",!0))},r.interval)}function c(){i.value!==null&&window.clearInterval(i.value)}return(e,z)=>(o(),B("div",$,[m("div",q,[C(l(k),{"cta-is-hidden":""},{title:y(()=>[m("div",N,[t.value?(o(),d(l(u),{key:0,icon:"spinner",color:"var(--grey-300)",size:"42"})):a.hasError?(o(),d(l(u),{key:1,icon:"errorFilled",color:"var(--red-500)",size:"42"})):(o(),d(l(u),{key:2,icon:"circleCheck",color:"var(--green-500)",size:"42"}))]),E(),t.value?n(e.$slots,"loading-title",{key:0}):a.hasError?n(e.$slots,"error-title",{key:1}):n(e.$slots,"complete-title",{key:2})]),message:y(()=>[t.value?n(e.$slots,"loading-content",{key:0}):a.hasError?n(e.$slots,"error-content",{key:1}):n(e.$slots,"complete-content",{key:2})]),_:3})])]))}});export{S as _};
