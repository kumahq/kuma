import{u,j as k}from"./kongponents.es-76ff1c1d.js";import{d as g,r as s,k as b,a2 as w,j as B,i as m,g as C,w as y,b as o,o as l,e as d,h as E,z as n}from"./index-e1c5e7d3.js";const $={class:"scanner"},N={class:"scanner-content"},z={class:"mb-2"},j=g({__name:"EntityScanner",props:{interval:{type:Number,required:!1,default:1e3},retries:{type:Number,required:!1,default:3600},hasError:{type:Boolean,default:!1},loaderFunction:{type:Function,required:!0},canComplete:{type:Boolean,default:!1}},emits:["hide-siblings"],setup(a,{emit:p}){const r=a,f=s(0),t=s(!1),v=s(!1),i=s(null);b(function(){h()}),w(function(){c()});function h(){t.value=!0,v.value=!1,c(),i.value=window.setInterval(async()=>{f.value++,await r.loaderFunction(),(f.value===r.retries||r.canComplete===!0)&&(c(),t.value=!1,v.value=!0,p("hide-siblings",!0))},r.interval)}function c(){i.value!==null&&window.clearInterval(i.value)}return(e,F)=>(l(),B("div",$,[m("div",N,[C(o(k),{"cta-is-hidden":""},{title:y(()=>[m("div",z,[t.value?(l(),d(o(u),{key:0,icon:"spinner",color:"var(--grey-300)",size:"42"})):a.hasError?(l(),d(o(u),{key:1,icon:"errorFilled",color:"var(--red-500)",size:"42"})):(l(),d(o(u),{key:2,icon:"circleCheck",color:"var(--green-500)",size:"42"}))]),E(),t.value?n(e.$slots,"loading-title",{key:0}):a.hasError?n(e.$slots,"error-title",{key:1}):n(e.$slots,"complete-title",{key:2})]),message:y(()=>[t.value?n(e.$slots,"loading-content",{key:0}):a.hasError?n(e.$slots,"error-content",{key:1}):n(e.$slots,"complete-content",{key:2})]),_:3})])]))}});export{j as _};
