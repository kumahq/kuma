import{d as x,c as k,o as r,a as i,w as e,q as b,h as s,b as n,H as g,g as a,t as v,e as z,F as y,s as $}from"./index-e262a3eb.js";import{a as A,A as B,S,b as D}from"./SubscriptionHeader-17bd9ec2.js";import{i as O,D as _,S as q,T as w,A as E,o as V,s as Z,E as C,t as T,_ as F}from"./RouteView.vue_vue_type_script_setup_true_lang-64a2b575.js";import{E as I}from"./EnvoyData-0a19149d.js";import{_ as N}from"./TabsWidget.vue_vue_type_style_index_0_lang-0ad41a39.js";import{g as H}from"./dataplane-30467516.js";import{_ as L}from"./RouteTitle.vue_vue_type_script_setup_true_lang-3bfddf35.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-506bff80.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-19a04e43.js";const P={class:"stack"},j={class:"columns",style:{"--columns":"4"}},R=x({__name:"ZoneIngressDetails",props:{zoneIngressOverview:{type:Object,required:!0}},setup(f){const t=f,{t:o}=O(),h=[{hash:"#overview",title:o("zone-ingresses.routes.item.tabs.overview")},{hash:"#xds-configuration",title:o("zone-ingresses.routes.item.tabs.xds_configuration")},{hash:"#envoy-stats",title:o("zone-ingresses.routes.item.tabs.stats")},{hash:"#envoy-clusters",title:o("zone-ingresses.routes.item.tabs.clusters")}],u=k(()=>H(t.zoneIngressOverview.zoneIngressInsight)),m=k(()=>{var c;const p=((c=t.zoneIngressOverview.zoneIngressInsight)==null?void 0:c.subscriptions)??[];return Array.from(p).reverse()});return(p,c)=>(r(),i(N,{tabs:h},{overview:e(()=>[b("div",P,[s(n(g),null,{body:e(()=>[b("div",j,[s(_,null,{title:e(()=>[a(v(n(o)("http.api.property.status")),1)]),body:e(()=>[s(q,{status:u.value},null,8,["status"])]),_:1}),a(),s(_,null,{title:e(()=>[a(v(n(o)("http.api.property.name")),1)]),body:e(()=>[s(w,{text:t.zoneIngressOverview.name},null,8,["text"])]),_:1}),a(),s(_,null,{title:e(()=>[a(v(n(o)("http.api.property.address")),1)]),body:e(()=>{var l,d;return[(l=t.zoneIngressOverview.zoneIngress.networking)!=null&&l.address&&((d=t.zoneIngressOverview.zoneIngress.networking)!=null&&d.port)?(r(),i(w,{key:0,text:`${t.zoneIngressOverview.zoneIngress.networking.address}:${t.zoneIngressOverview.zoneIngress.networking.port}`},null,8,["text"])):(r(),z(y,{key:1},[a(v(n(o)("common.detail.none")),1)],64))]}),_:1}),a(),s(_,null,{title:e(()=>[a(v(n(o)("http.api.property.advertisedAddress")),1)]),body:e(()=>{var l,d;return[(l=t.zoneIngressOverview.zoneIngress.networking)!=null&&l.advertisedAddress&&((d=t.zoneIngressOverview.zoneIngress.networking)!=null&&d.advertisedPort)?(r(),i(w,{key:0,text:`${t.zoneIngressOverview.zoneIngress.networking.advertisedAddress}:${t.zoneIngressOverview.zoneIngress.networking.advertisedPort}`},null,8,["text"])):(r(),z(y,{key:1},[a(v(n(o)("common.detail.none")),1)],64))]}),_:1})])]),_:1}),a(),s(n(g),null,{body:e(()=>[s(A,{"initially-open":0},{default:e(()=>[(r(!0),z(y,null,$(m.value,(l,d)=>(r(),i(B,{key:d},{"accordion-header":e(()=>[s(S,{subscription:l},null,8,["subscription"])]),"accordion-content":e(()=>[s(D,{subscription:l,"is-discovery-subscription":""},null,8,["subscription"])]),_:2},1024))),128))]),_:1})]),_:1})])]),"xds-configuration":e(()=>[s(n(g),null,{body:e(()=>[s(I,{status:u.value,resource:"Zone",src:`/zone-ingresses/${t.zoneIngressOverview.name}/data-path/xds`,"query-key":"envoy-data-xds-zone-ingress"},null,8,["status","src"])]),_:1})]),"envoy-stats":e(()=>[s(n(g),null,{body:e(()=>[s(I,{status:u.value,resource:"Zone",src:`/zone-ingresses/${t.zoneIngressOverview.name}/data-path/stats`,"query-key":"envoy-data-stats-zone-ingress"},null,8,["status","src"])]),_:1})]),"envoy-clusters":e(()=>[s(n(g),null,{body:e(()=>[s(I,{status:u.value,resource:"Zone",src:`/zone-ingresses/${t.zoneIngressOverview.name}/data-path/clusters`,"query-key":"envoy-data-clusters-zone-ingress"},null,8,["status","src"])]),_:1})]),_:1}))}}),ee=x({__name:"ZoneIngressDetailView",setup(f){const{t}=O();return(o,h)=>(r(),i(F,{name:"zone-ingress-detail-view","data-testid":"zone-ingress-detail-view"},{default:e(({route:u})=>[s(E,{breadcrumbs:[{to:{name:"zone-ingress-list-view"},text:n(t)("zone-ingresses.routes.item.breadcrumbs")}]},{title:e(()=>[b("h1",null,[s(L,{title:n(t)("zone-ingresses.routes.item.title",{name:u.params.zoneIngress}),render:!0},null,8,["title"])])]),default:e(()=>[a(),s(V,{src:`/zone-ingresses/${u.params.zoneIngress}`},{default:e(({data:m,isLoading:p,error:c})=>[p?(r(),i(Z,{key:0})):c!==void 0?(r(),i(C,{key:1,error:c},null,8,["error"])):m===void 0?(r(),i(T,{key:2})):(r(),i(R,{key:3,"zone-ingress-overview":m,"data-testid":"detail-view-details"},null,8,["zone-ingress-overview"]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{ee as default};
