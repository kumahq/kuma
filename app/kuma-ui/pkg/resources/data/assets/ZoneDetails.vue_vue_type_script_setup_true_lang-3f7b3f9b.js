import{O as S}from"./kongponents.es-c741eab8.js";import{A as C,a as I}from"./AccordionList-1e0cd60d.js";import{_ as A}from"./CodeBlock.vue_vue_type_style_index_0_lang-90b7b8c9.js";import{D as N,a as D}from"./DefinitionListItem-87124c14.js";import{_ as V,S as T}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-dceeb786.js";import{T as B}from"./TabsWidget-597e6684.js";import{_ as L}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-0a78fe3e.js";import{u as Z}from"./store-5ee7e2bf.js";import{d as x,c as a,D as E,I as j,J as q,b as c,w as t,o,i as F,t as l,g as u,j as p,q as v,e as J,h as _,F as m,f as $}from"./index-0fbacd76.js";const G={class:"entity-heading"},Y=x({__name:"ZoneDetails",props:{zoneOverview:{type:Object,required:!0}},setup(b){const r=b,w=Z(),g=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Insights"},{hash:"#config",title:"Config"},{hash:"#warnings",title:"Warnings"}],f=a(()=>{const{type:n,name:s}=r.zoneOverview,e=E(r.zoneOverview.zoneInsight);return{type:n,name:s,status:e,"Authentication Type":j(r.zoneOverview)}}),O=a(()=>{var s;const n=((s=r.zoneOverview.zoneInsight)==null?void 0:s.subscriptions)??[];return Array.from(n).reverse()}),d=a(()=>{var e;const n=[],s=((e=r.zoneOverview.zoneInsight)==null?void 0:e.subscriptions)??[];if(s.length>0){const i=s[s.length-1],y=i.version.kumaCp.version||"-",{kumaCpGlobalCompatible:z=!0}=i.version.kumaCp;z||n.push({kind:q,payload:{zoneCpVersion:y,globalCpVersion:w.getters["config/getVersion"]}})}return n}),h=a(()=>{var e;const n=((e=r.zoneOverview.zoneInsight)==null?void 0:e.subscriptions)??[],s=n[n.length-1];return s.config?JSON.stringify(JSON.parse(s.config),null,2):null}),k=a(()=>d.value.length===0?g.filter(n=>n.hash!=="#warnings"):g);return(n,s)=>(o(),c(B,{tabs:k.value},{tabHeader:t(()=>[F("h1",G,`
        Zone: `+l(f.value.name),1)]),overview:t(()=>[u(D,null,{default:t(()=>[(o(!0),p(m,null,v(f.value,(e,i)=>(o(),c(N,{key:i,term:i},{default:t(()=>[i==="status"?(o(),c(J(S),{key:0,appearance:e==="Offline"?"danger":"success"},{default:t(()=>[_(l(e),1)]),_:2},1032,["appearance"])):(o(),p(m,{key:1},[_(l(e),1)],64))]),_:2},1032,["term"]))),128))]),_:1})]),insights:t(()=>[u(I,{"initially-open":0},{default:t(()=>[(o(!0),p(m,null,v(O.value,(e,i)=>(o(),c(C,{key:i},{"accordion-header":t(()=>[u(V,{details:e},null,8,["details"])]),"accordion-content":t(()=>[u(T,{details:e},null,8,["details"])]),_:2},1024))),128))]),_:1})]),config:t(()=>[h.value!==null?(o(),c(A,{key:0,id:"code-block-zone-config",language:"json",code:h.value,"is-searchable":"","query-key":"zone-config"},null,8,["code"])):$("",!0)]),warnings:t(()=>[u(L,{warnings:d.value},null,8,["warnings"])]),_:1},8,["tabs"]))}});export{Y as _};
