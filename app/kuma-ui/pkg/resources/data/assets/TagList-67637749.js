import{_ as f}from"./kongponents.es-3ba46133.js";import{a as b}from"./production-554ae9d4.js";import{d as g,c as y,o as r,h as u,F as k,l as h,a as l,w as i,m as x,f as L,t as _,g as T,u as p}from"./runtime-dom.esm-bundler-9284044f.js";import{_ as w}from"./_plugin-vue_export-helper-c27b6911.js";function B(o){return Object.entries(o??{}).map(([a,s])=>({label:a,value:s}))}const C={class:"tag-list"},j=g({__name:"TagList",props:{tags:{type:Object,required:!0}},setup(o){const a=o,s=b(),m=y(()=>(Array.isArray(a.tags)?a.tags:B(a.tags)).map(n=>{const{label:t,value:c}=n,v=d(n);return{label:t,value:c,route:v}}));function d(e){if(e.value!=="*")try{switch(e.label){case"kuma.io/zone":return s.resolve({name:"zones",query:{ns:e.value}});case"kuma.io/service":return s.resolve({name:"service-detail-view",params:{service:e.value}});default:return}}catch{return}}return(e,n)=>(r(),u("span",C,[(r(!0),u(k,null,h(p(m),(t,c)=>(r(),l(p(f),{key:c,class:"tag-badge"},{default:i(()=>[(r(),l(x(t.route?"router-link":"span"),{to:t.route},{default:i(()=>[L(_(t.label)+":",1),T("b",null,_(t.value),1)]),_:2},1032,["to"]))]),_:2},1024))),128))]))}});const F=w(j,[["__scopeId","data-v-dc93a777"]]);export{F as T};
