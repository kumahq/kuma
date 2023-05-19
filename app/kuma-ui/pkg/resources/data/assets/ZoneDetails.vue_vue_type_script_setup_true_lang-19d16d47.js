import{O as z}from"./kongponents.es-186331d0.js";import{A as C,a as I}from"./AccordionList-33bf8c88.js";import{_ as A}from"./CodeBlock.vue_vue_type_style_index_0_lang-976cc725.js";import{a as N,D as V}from"./DefinitionListItem-7ae31aca.js";import{_ as D,S as T}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-17a8f1eb.js";import{T as B}from"./TabsWidget-aa4c3028.js";import{_ as L}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-7605d700.js";import{u as Z}from"./store-ba47d389.js";import{d as x,c as a,S as E,a7 as j,a8 as q,o,b as c,w as t,i as F,t as l,g as u,j as p,q as v,e as $,h as _,F as m,f as G}from"./index-f9d036ef.js";const J={class:"entity-heading"},Y=x({__name:"ZoneDetails",props:{zoneOverview:{type:Object,required:!0}},setup(b){const r=b,w=Z(),g=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Insights"},{hash:"#config",title:"Config"},{hash:"#warnings",title:"Warnings"}],f=a(()=>{const{type:n,name:s}=r.zoneOverview,e=E(r.zoneOverview.zoneInsight);return{type:n,name:s,status:e,"Authentication Type":j(r.zoneOverview)}}),O=a(()=>{var s;const n=((s=r.zoneOverview.zoneInsight)==null?void 0:s.subscriptions)??[];return Array.from(n).reverse()}),d=a(()=>{var e;const n=[],s=((e=r.zoneOverview.zoneInsight)==null?void 0:e.subscriptions)??[];if(s.length>0){const i=s[s.length-1],k=i.version.kumaCp.version||"-",{kumaCpGlobalCompatible:y=!0}=i.version.kumaCp;y||n.push({kind:q,payload:{zoneCpVersion:k,globalCpVersion:w.getters["config/getVersion"]}})}return n}),h=a(()=>{var e;const n=((e=r.zoneOverview.zoneInsight)==null?void 0:e.subscriptions)??[],s=n[n.length-1];return s.config?JSON.stringify(JSON.parse(s.config),null,2):null}),S=a(()=>d.value.length===0?g.filter(n=>n.hash!=="#warnings"):g);return(n,s)=>(o(),c(B,{tabs:S.value},{tabHeader:t(()=>[F("h1",J,`
        Zone: `+l(f.value.name),1)]),overview:t(()=>[u(V,null,{default:t(()=>[(o(!0),p(m,null,v(f.value,(e,i)=>(o(),c(N,{key:i,term:i},{default:t(()=>[i==="status"?(o(),c($(z),{key:0,appearance:e==="Offline"?"danger":"success"},{default:t(()=>[_(l(e),1)]),_:2},1032,["appearance"])):(o(),p(m,{key:1},[_(l(e),1)],64))]),_:2},1032,["term"]))),128))]),_:1})]),insights:t(()=>[u(I,{"initially-open":0},{default:t(()=>[(o(!0),p(m,null,v(O.value,(e,i)=>(o(),c(C,{key:i},{"accordion-header":t(()=>[u(D,{details:e},null,8,["details"])]),"accordion-content":t(()=>[u(T,{details:e},null,8,["details"])]),_:2},1024))),128))]),_:1})]),config:t(()=>[h.value!==null?(o(),c(A,{key:0,id:"code-block-zone-config",language:"json",code:h.value,"is-searchable":"","query-key":"zone-config"},null,8,["code"])):G("",!0)]),warnings:t(()=>[u(L,{warnings:d.value},null,8,["warnings"])]),_:1},8,["tabs"]))}});export{Y as _};
