import{d as h,c as b,o as a,a as i,w as e,h as s,b as r,G as u,q as f,g as n,t as p,e as x,F as k,s as $}from"./index-30c3bdbc.js";import{a as O,A as B,S,b as A}from"./SubscriptionHeader-4cbb664c.js";import{g as w,D as g,S as D,A as q,o as I,q as V,E as C,r as T,_ as F}from"./RouteView.vue_vue_type_script_setup_true_lang-1dd6f4c1.js";import{E as v}from"./EnvoyData-21b8c7a9.js";import{_ as N}from"./TabsWidget.vue_vue_type_style_index_0_lang-3e0839b4.js";import{T as L}from"./TextWithCopyButton-19c2f2e4.js";import{g as Z}from"./dataplane-30467516.js";import{_ as j}from"./RouteTitle.vue_vue_type_script_setup_true_lang-cbf5001a.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-b3f1f8ad.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-bf980d9d.js";import"./CopyButton-bbd9c9e2.js";const G={class:"columns",style:{"--columns":"3"}},H=h({__name:"ZoneEgressDetails",props:{zoneEgressOverview:{type:Object,required:!0}},setup(_){const t=_,{t:o}=w(),y=[{hash:"#overview",title:o("zone-egresses.routes.item.tabs.overview")},{hash:"#insights",title:o("zone-egresses.routes.item.tabs.insights")},{hash:"#xds-configuration",title:o("zone-egresses.routes.item.tabs.xds_configuration")},{hash:"#envoy-stats",title:o("zone-egresses.routes.item.tabs.stats")},{hash:"#envoy-clusters",title:o("zone-egresses.routes.item.tabs.clusters")}],c=b(()=>Z(t.zoneEgressOverview.zoneEgressInsight)),m=b(()=>{var l;const d=((l=t.zoneEgressOverview.zoneEgressInsight)==null?void 0:l.subscriptions)??[];return Array.from(d).reverse()});return(d,l)=>(a(),i(N,{tabs:y},{overview:e(()=>[s(r(u),null,{body:e(()=>[f("div",G,[s(g,null,{title:e(()=>[n(p(r(o)("http.api.property.status")),1)]),body:e(()=>[s(D,{status:c.value},null,8,["status"])]),_:1}),n(),s(g,null,{title:e(()=>[n(p(r(o)("http.api.property.name")),1)]),body:e(()=>[s(L,{text:t.zoneEgressOverview.name},null,8,["text"])]),_:1}),n(),s(g,null,{title:e(()=>[n(p(r(o)("http.api.property.type")),1)]),body:e(()=>[n(p(t.zoneEgressOverview.type),1)]),_:1})])]),_:1})]),insights:e(()=>[s(r(u),null,{body:e(()=>[s(O,{"initially-open":0},{default:e(()=>[(a(!0),x(k,null,$(m.value,(z,E)=>(a(),i(B,{key:E},{"accordion-header":e(()=>[s(S,{subscription:z},null,8,["subscription"])]),"accordion-content":e(()=>[s(A,{subscription:z,"is-discovery-subscription":""},null,8,["subscription"])]),_:2},1024))),128))]),_:1})]),_:1})]),"xds-configuration":e(()=>[s(r(u),null,{body:e(()=>[s(v,{src:`/zone-egresses/${t.zoneEgressOverview.name}/data-path/xds`,"query-key":"envoy-data-xds-zone-egress"},null,8,["src"])]),_:1})]),"envoy-stats":e(()=>[s(r(u),null,{body:e(()=>[s(v,{src:`/zone-egresses/${t.zoneEgressOverview.name}/data-path/stats`,"query-key":"envoy-data-stats-zone-egress"},null,8,["src"])]),_:1})]),"envoy-clusters":e(()=>[s(r(u),null,{body:e(()=>[s(v,{src:`/zone-egresses/${t.zoneEgressOverview.name}/data-path/clusters`,"query-key":"envoy-data-clusters-zone-egress"},null,8,["src"])]),_:1})]),_:1}))}}),se=h({__name:"ZoneEgressDetailView",setup(_){const{t}=w();return(o,y)=>(a(),i(F,{name:"zone-egress-detail-view","data-testid":"zone-egress-detail-view"},{default:e(({route:c})=>[s(q,{breadcrumbs:[{to:{name:"zone-egress-list-view"},text:r(t)("zone-egresses.routes.item.breadcrumbs")}]},{title:e(()=>[f("h1",null,[s(j,{title:r(t)("zone-egresses.routes.item.title",{name:c.params.zoneEgress}),render:!0},null,8,["title"])])]),default:e(()=>[n(),s(I,{src:`/zone-egresses/${c.params.zoneEgress}`},{default:e(({data:m,isLoading:d,error:l})=>[d?(a(),i(V,{key:0})):l!==void 0?(a(),i(C,{key:1,error:l},null,8,["error"])):m===void 0?(a(),i(T,{key:2})):(a(),i(H,{key:3,"zone-egress-overview":m,"data-testid":"detail-view-details"},null,8,["zone-egress-overview"]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{se as default};
