import{a as x,A as D,_ as O,S as $}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-6640b218.js";import{D as A,a as B}from"./DefinitionListItem-aa9fbb97.js";import{E as d}from"./EnvoyData-9bef97ec.js";import{T as C}from"./TabsWidget-9d1f7c85.js";import{T as L}from"./TextWithCopyButton-7c3f1ae7.js";import{g as E,A as S,p as T,r as q,E as I,s as V,_ as Z}from"./RouteView.vue_vue_type_script_setup_true_lang-b0370148.js";import{d as k,c as y,r as N,o as t,a as n,w as e,h as s,e as u,F as v,s as h,b as g,g as p,t as w,q as R}from"./index-fd0688ab.js";import{_ as F}from"./RouteTitle.vue_vue_type_script_setup_true_lang-0a897f5f.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-4bbdcb3f.js";import"./CopyButton-ce689fe8.js";const W=k({__name:"ZoneEgressDetails",props:{zoneEgressOverview:{type:Object,required:!0}},setup(_){const r=_,{t:f}=E(),z=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Egress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],o=y(()=>{const{type:i,name:a}=r.zoneEgressOverview;return{type:i,name:a}}),c=y(()=>{var a;const i=((a=r.zoneEgressOverview.zoneEgressInsight)==null?void 0:a.subscriptions)??[];return Array.from(i).reverse()});return(i,a)=>{const b=N("RouterLink");return t(),n(C,{tabs:z},{overview:e(()=>[s(A,null,{default:e(()=>[(t(!0),u(v,null,h(o.value,(m,l)=>(t(),n(B,{key:l,term:g(f)(`http.api.property.${l}`)},{default:e(()=>[l==="name"?(t(),n(L,{key:0,text:r.zoneEgressOverview.name},{default:e(()=>[s(b,{to:{name:"zone-egress-detail-view",params:{zoneEgress:r.zoneEgressOverview.name}}},{default:e(()=>[p(w(r.zoneEgressOverview.name),1)]),_:1},8,["to"])]),_:1},8,["text"])):(t(),u(v,{key:1},[p(w(m),1)],64))]),_:2},1032,["term"]))),128))]),_:1})]),insights:e(()=>[s(x,{"initially-open":0},{default:e(()=>[(t(!0),u(v,null,h(c.value,(m,l)=>(t(),n(D,{key:l},{"accordion-header":e(()=>[s(O,{details:m},null,8,["details"])]),"accordion-content":e(()=>[s($,{details:m,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),"xds-configuration":e(()=>[s(d,{"data-path":"xds","zone-egress-name":o.value.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),"envoy-stats":e(()=>[s(d,{"data-path":"stats","zone-egress-name":o.value.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),"envoy-clusters":e(()=>[s(d,{"data-path":"clusters","zone-egress-name":o.value.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),_:1})}}}),j={key:3,class:"kcard-border","data-testid":"detail-view-details"},ee=k({__name:"ZoneEgressDetailView",setup(_){const{t:r}=E();return(f,z)=>(t(),n(Z,{name:"zone-egress-detail-view","data-testid":"zone-egress-detail-view"},{default:e(({route:o})=>[s(S,{breadcrumbs:[{to:{name:"zone-egress-list-view"},text:g(r)("zone-egresses.routes.item.breadcrumbs")}]},{title:e(()=>[R("h1",null,[s(F,{title:g(r)("zone-egresses.routes.item.title",{name:o.params.zoneEgress}),render:!0},null,8,["title"])])]),default:e(()=>[p(),s(T,{src:`/zone-egresses/${o.params.zoneEgress}`},{default:e(({data:c,isLoading:i,error:a})=>[i?(t(),n(q,{key:0})):a!==void 0?(t(),n(I,{key:1,error:a},null,8,["error"])):c===void 0?(t(),n(V,{key:2})):(t(),u("div",j,[s(W,{"zone-egress-overview":c},null,8,["zone-egress-overview"])]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{ee as default};
