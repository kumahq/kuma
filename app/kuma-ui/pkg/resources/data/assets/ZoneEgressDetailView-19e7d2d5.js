import{d as h,c as b,r as x,o as a,a as i,w as e,h as s,b as r,G as u,q as f,g as n,t as c,e as O,F as B,s as S}from"./index-f1d26a4d.js";import{a as A,A as D,S as $,b as q}from"./SubscriptionHeader-82dae09c.js";import{g as w,D as p,S as I,A as T,o as C,q as V,E as L,r as F,_ as N}from"./RouteView.vue_vue_type_script_setup_true_lang-658e86d1.js";import{E as v}from"./EnvoyData-32123e3e.js";import{T as R}from"./TabsWidget-36b2492a.js";import{T as W}from"./TextWithCopyButton-8730651e.js";import{g as Z}from"./dataplane-30467516.js";import{_ as j}from"./RouteTitle.vue_vue_type_script_setup_true_lang-a47c6428.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-d695b393.js";import"./CopyButton-51a400ca.js";const G={class:"variable-columns"},H=h({__name:"ZoneEgressDetails",props:{zoneEgressOverview:{type:Object,required:!0}},setup(_){const t=_,{t:o}=w(),y=[{hash:"#overview",title:o("zone-egresses.routes.item.tabs.overview")},{hash:"#insights",title:o("zone-egresses.routes.item.tabs.insights")},{hash:"#xds-configuration",title:o("zone-egresses.routes.item.tabs.xds_configuration")},{hash:"#envoy-stats",title:o("zone-egresses.routes.item.tabs.stats")},{hash:"#envoy-clusters",title:o("zone-egresses.routes.item.tabs.clusters")}],m=b(()=>Z(t.zoneEgressOverview.zoneEgressInsight)),d=b(()=>{var l;const g=((l=t.zoneEgressOverview.zoneEgressInsight)==null?void 0:l.subscriptions)??[];return Array.from(g).reverse()});return(g,l)=>{const E=x("RouterLink");return a(),i(R,{tabs:y},{overview:e(()=>[s(r(u),null,{body:e(()=>[f("div",G,[s(p,null,{title:e(()=>[n(c(r(o)("http.api.property.status")),1)]),body:e(()=>[s(I,{status:m.value},null,8,["status"])]),_:1}),n(),s(p,null,{title:e(()=>[n(c(r(o)("http.api.property.name")),1)]),body:e(()=>[s(W,{text:t.zoneEgressOverview.name},{default:e(()=>[s(E,{to:{name:"zone-egress-detail-view",params:{zoneEgress:t.zoneEgressOverview.name}}},{default:e(()=>[n(c(t.zoneEgressOverview.name),1)]),_:1},8,["to"])]),_:1},8,["text"])]),_:1}),n(),s(p,null,{title:e(()=>[n(c(r(o)("http.api.property.type")),1)]),body:e(()=>[n(c(t.zoneEgressOverview.type),1)]),_:1})])]),_:1})]),insights:e(()=>[s(r(u),null,{body:e(()=>[s(A,{"initially-open":0},{default:e(()=>[(a(!0),O(B,null,S(d.value,(z,k)=>(a(),i(D,{key:k},{"accordion-header":e(()=>[s($,{subscription:z},null,8,["subscription"])]),"accordion-content":e(()=>[s(q,{subscription:z,"is-discovery-subscription":""},null,8,["subscription"])]),_:2},1024))),128))]),_:1})]),_:1})]),"xds-configuration":e(()=>[s(r(u),null,{body:e(()=>[s(v,{"data-path":"xds","zone-egress-name":t.zoneEgressOverview.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),_:1})]),"envoy-stats":e(()=>[s(r(u),null,{body:e(()=>[s(v,{"data-path":"stats","zone-egress-name":t.zoneEgressOverview.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),_:1})]),"envoy-clusters":e(()=>[s(r(u),null,{body:e(()=>[s(v,{"data-path":"clusters","zone-egress-name":t.zoneEgressOverview.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),_:1})]),_:1})}}}),te=h({__name:"ZoneEgressDetailView",setup(_){const{t}=w();return(o,y)=>(a(),i(N,{name:"zone-egress-detail-view","data-testid":"zone-egress-detail-view"},{default:e(({route:m})=>[s(T,{breadcrumbs:[{to:{name:"zone-egress-list-view"},text:r(t)("zone-egresses.routes.item.breadcrumbs")}]},{title:e(()=>[f("h1",null,[s(j,{title:r(t)("zone-egresses.routes.item.title",{name:m.params.zoneEgress}),render:!0},null,8,["title"])])]),default:e(()=>[n(),s(C,{src:`/zone-egresses/${m.params.zoneEgress}`},{default:e(({data:d,isLoading:g,error:l})=>[g?(a(),i(V,{key:0})):l!==void 0?(a(),i(L,{key:1,error:l},null,8,["error"])):d===void 0?(a(),i(F,{key:2})):(a(),i(H,{key:3,"zone-egress-overview":d,"data-testid":"detail-view-details"},null,8,["zone-egress-overview"]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{te as default};
