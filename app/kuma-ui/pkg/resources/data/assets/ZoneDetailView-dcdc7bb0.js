import{d as S,c as u,r as V,o as n,a as o,w as e,q as B,g as p,h as a,t as g,e as h,F as b,s as C,b as m,R as N,X as T}from"./index-a928d02c.js";import{a as D,A as E,_ as L,S as Z}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-8c0b83b6.js";import{_ as R}from"./CodeBlock.vue_vue_type_style_index_0_lang-8fce21bd.js";import{D as q,a as F}from"./DefinitionListItem-e5ff9064.js";import{T as M}from"./TabsWidget-c30a7efe.js";import{T as I}from"./TextWithCopyButton-406f01b5.js";import{_ as P}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-7356eda1.js";import{g as A,z as W,n as j,B as G,E as J,A as X,q as H,_ as K}from"./RouteView.vue_vue_type_script_setup_true_lang-f622f9ae.js";import{_ as U}from"./DataSource.vue_vue_type_script_setup_true_lang-8eeb0b78.js";import{_ as Q}from"./RouteTitle.vue_vue_type_script_setup_true_lang-a99b1649.js";import{_ as Y}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-be8f92ee.js";import{E as ee}from"./ErrorBlock-57542f11.js";const ne={class:"entity-heading"},te=S({__name:"ZoneDetails",props:{zoneOverview:{type:Object,required:!0}},setup(z){const i=z,{t:w}=A(),k=W(),_=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Insights"},{hash:"#config",title:"Config"},{hash:"#warnings",title:"Warnings"}],f=u(()=>({name:"zone-cp-detail-view",params:{zone:i.zoneOverview.name}})),d=u(()=>{const{type:t,name:s}=i.zoneOverview,c=j(i.zoneOverview.zoneInsight);return{type:t,name:s,status:c,"Authentication Type":G(i.zoneOverview)}}),v=u(()=>{var s;const t=((s=i.zoneOverview.zoneInsight)==null?void 0:s.subscriptions)??[];return Array.from(t).reverse()}),y=u(()=>{var c;const t=[],s=((c=i.zoneOverview.zoneInsight)==null?void 0:c.subscriptions)??[];if(s.length>0){const r=s[s.length-1],l=r.version.kumaCp.version||"-",{kumaCpGlobalCompatible:x=!0}=r.version.kumaCp;x||t.push({kind:J,payload:{zoneCpVersion:l,globalCpVersion:k("KUMA_VERSION")}})}return t}),O=u(()=>{var s;const t=((s=i.zoneOverview.zoneInsight)==null?void 0:s.subscriptions)??[];if(t.length>0){const c=t[t.length-1];if(c.config)return JSON.stringify(JSON.parse(c.config),null,2)}return null}),$=u(()=>y.value.length===0?_.filter(t=>t.hash!=="#warnings"):_);return(t,s)=>{const c=V("router-link");return n(),o(M,{tabs:$.value},{tabHeader:e(()=>[B("h1",ne,[p(`
        Zone Control Plane:

        `),a(I,{text:d.value.name},{default:e(()=>[a(c,{to:f.value},{default:e(()=>[p(g(d.value.name),1)]),_:1},8,["to"])]),_:1},8,["text"])])]),overview:e(()=>[a(q,null,{default:e(()=>[(n(!0),h(b,null,C(d.value,(r,l)=>(n(),o(F,{key:l,term:m(w)(`http.api.property.${l}`)},{default:e(()=>[l==="status"?(n(),o(m(N),{key:0,appearance:r==="offline"?"danger":"success"},{default:e(()=>[p(g(r),1)]),_:2},1032,["appearance"])):l==="name"?(n(),o(I,{key:1,text:r},null,8,["text"])):(n(),h(b,{key:2},[p(g(r),1)],64))]),_:2},1032,["term"]))),128))]),_:1})]),insights:e(()=>[a(D,{"initially-open":0},{default:e(()=>[(n(!0),h(b,null,C(v.value,(r,l)=>(n(),o(E,{key:l},{"accordion-header":e(()=>[a(L,{details:r},null,8,["details"])]),"accordion-content":e(()=>[a(Z,{details:r},null,8,["details"])]),_:2},1024))),128))]),_:1})]),config:e(()=>[O.value!==null?(n(),o(R,{key:0,id:"code-block-zone-config",language:"json",code:O.value,"is-searchable":"","query-key":"zone-config"},null,8,["code"])):(n(),o(m(T),{key:1,"data-testid":"warning-no-subscriptions",appearance:"warning"},{alertMessage:e(()=>[p(g(m(w)("zone-cps.routes.item.config.no-subscriptions")),1)]),_:1}))]),warnings:e(()=>[a(P,{warnings:y.value},null,8,["warnings"])]),_:1},8,["tabs"])}}}),se={key:3,class:"kcard-border","data-testid":"detail-view-details"},ve=S({__name:"ZoneDetailView",setup(z){const{t:i}=A();return(w,k)=>(n(),o(K,{name:"zone-cp-detail-view"},{default:e(({route:_})=>[a(Q,{title:m(i)("zone-cps.routes.item.title",{name:_.params.zone})},null,8,["title"]),p(),a(X,{breadcrumbs:[{to:{name:"zone-cp-list-view"},text:m(i)("zone-cps.routes.item.breadcrumbs")}]},{default:e(()=>[a(U,{src:`/zone-cps/${_.params.zone}`},{default:e(({data:f,isLoading:d,error:v})=>[d?(n(),o(H,{key:0})):v!==void 0?(n(),o(ee,{key:1,error:v},null,8,["error"])):f===void 0?(n(),o(Y,{key:2})):(n(),h("div",se,[a(te,{"zone-overview":f},null,8,["zone-overview"])]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{ve as default};
