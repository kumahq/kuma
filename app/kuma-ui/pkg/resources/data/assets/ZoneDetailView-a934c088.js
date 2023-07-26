import{d as x,c as v,r as N,o as t,a,w as e,q as V,g as d,h as r,t as z,e as k,s as I,b as f,F as O,u as T,j as A}from"./index-9551a132.js";import{L as $,j as B}from"./kongponents.es-8c883b13.js";import{A as E,_ as Z,S as j,a as R}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-a8f38967.js";import{_ as F}from"./CodeBlock.vue_vue_type_style_index_0_lang-4a1bfa0a.js";import{a as q,D as G}from"./DefinitionListItem-25818dd8.js";import{T as M}from"./TabsWidget-045a1489.js";import{T as S}from"./TextWithCopyButton-893964c0.js";import{_ as P}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-296bb0f0.js";import{g as D,F as W,o as H,G as J,H as K,n as U,h as Q,A as X,_ as Y}from"./RouteView.vue_vue_type_script_setup_true_lang-d5114b59.js";import{_ as ee}from"./RouteTitle.vue_vue_type_script_setup_true_lang-e2244129.js";import{_ as ne}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-cc9bff92.js";import{E as te}from"./ErrorBlock-1ffa56a4.js";const se={class:"entity-heading"},ae=x({__name:"ZoneDetails",props:{zoneOverview:{type:Object,required:!0}},setup(C){const i=C,{t:h}=D(),w=W(),p=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Insights"},{hash:"#config",title:"Config"},{hash:"#warnings",title:"Warnings"}],_=v(()=>({name:"zone-cp-detail-view",params:{zone:i.zoneOverview.name}})),c=v(()=>{const{type:n,name:s}=i.zoneOverview,l=H(i.zoneOverview.zoneInsight);return{type:n,name:s,status:l,"Authentication Type":J(i.zoneOverview)}}),y=v(()=>{var s;const n=((s=i.zoneOverview.zoneInsight)==null?void 0:s.subscriptions)??[];return Array.from(n).reverse()}),b=v(()=>{var l;const n=[],s=((l=i.zoneOverview.zoneInsight)==null?void 0:l.subscriptions)??[];if(s.length>0){const o=s[s.length-1],u=o.version.kumaCp.version||"-",{kumaCpGlobalCompatible:L=!0}=o.version.kumaCp;L||n.push({kind:K,payload:{zoneCpVersion:u,globalCpVersion:w("KUMA_VERSION")}})}return n}),g=v(()=>{var s;const n=((s=i.zoneOverview.zoneInsight)==null?void 0:s.subscriptions)??[];if(n.length>0){const l=n[n.length-1];if(l.config)return JSON.stringify(JSON.parse(l.config),null,2)}return null}),m=v(()=>b.value.length===0?p.filter(n=>n.hash!=="#warnings"):p);return(n,s)=>{const l=N("router-link");return t(),a(M,{tabs:m.value},{tabHeader:e(()=>[V("h1",se,[d(`
        Zone Control Plane:

        `),r(S,{text:c.value.name},{default:e(()=>[r(l,{to:_.value},{default:e(()=>[d(z(c.value.name),1)]),_:1},8,["to"])]),_:1},8,["text"])])]),overview:e(()=>[r(G,null,{default:e(()=>[(t(!0),k(O,null,I(c.value,(o,u)=>(t(),a(q,{key:u,term:f(h)(`http.api.property.${u}`)},{default:e(()=>[u==="status"?(t(),a(f($),{key:0,appearance:o==="offline"?"danger":"success"},{default:e(()=>[d(z(o),1)]),_:2},1032,["appearance"])):u==="name"?(t(),a(S,{key:1,text:o},null,8,["text"])):(t(),k(O,{key:2},[d(z(o),1)],64))]),_:2},1032,["term"]))),128))]),_:1})]),insights:e(()=>[r(R,{"initially-open":0},{default:e(()=>[(t(!0),k(O,null,I(y.value,(o,u)=>(t(),a(E,{key:u},{"accordion-header":e(()=>[r(Z,{details:o},null,8,["details"])]),"accordion-content":e(()=>[r(j,{details:o},null,8,["details"])]),_:2},1024))),128))]),_:1})]),config:e(()=>[g.value!==null?(t(),a(F,{key:0,id:"code-block-zone-config",language:"json",code:g.value,"is-searchable":"","query-key":"zone-config"},null,8,["code"])):(t(),a(f(B),{key:1,"data-testid":"warning-no-subscriptions",appearance:"warning"},{alertMessage:e(()=>[d(z(f(h)("zone-cps.routes.item.config.no-subscriptions")),1)]),_:1}))]),warnings:e(()=>[r(P,{warnings:b.value},null,8,["warnings"])]),_:1},8,["tabs"])}}}),oe={class:"zone-details"},re={key:3,class:"kcard-border","data-testid":"detail-view-details"},we=x({__name:"ZoneDetailView",setup(C){const i=U(),h=T(),{t:w}=D(),p=A(null),_=A(!0),c=A(null);y();function y(){b()}async function b(){_.value=!0,c.value=null;const g=h.params.zone;try{p.value=await i.getZoneOverview({name:g})}catch(m){p.value=null,m instanceof Error?c.value=m:console.error(m)}finally{_.value=!1}}return(g,m)=>(t(),a(Y,null,{default:e(({route:n})=>[r(ee,{title:f(w)("zone-cps.routes.item.title",{name:n.params.zone})},null,8,["title"]),d(),r(X,{breadcrumbs:[{to:{name:"zone-cp-list-view"},text:f(w)("zone-cps.routes.item.breadcrumbs")}]},{default:e(()=>[V("div",oe,[_.value?(t(),a(Q,{key:0})):c.value!==null?(t(),a(te,{key:1,error:c.value},null,8,["error"])):p.value===null?(t(),a(ne,{key:2})):(t(),k("div",re,[r(ae,{"zone-overview":p.value},null,8,["zone-overview"])]))])]),_:1},8,["breadcrumbs"])]),_:1}))}});export{we as default};
