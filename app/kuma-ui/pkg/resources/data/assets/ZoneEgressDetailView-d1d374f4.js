import{a as D,A,_ as B,S as O}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-b6b92d2d.js";import{D as C,a as S}from"./DefinitionListItem-fba6c8a3.js";import{E as v}from"./EnvoyData-12e05dcb.js";import{T}from"./TabsWidget-dda31314.js";import{T as w}from"./TextWithCopyButton-494dc184.js";import{g as k,A as V,p as q,r as I,E as Z,s as L,_ as N}from"./RouteView.vue_vue_type_script_setup_true_lang-159ad8a0.js";import{d as x,c as p,r as F,o as t,a as n,w as e,q as R,g as m,h as s,t as E,e as d,F as g,s as b,b as f}from"./index-7e71fe76.js";import{_ as W}from"./RouteTitle.vue_vue_type_script_setup_true_lang-3c1a3272.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-0b9f82eb.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-e15754a4.js";const j={class:"entity-heading"},H=x({__name:"ZoneEgressDetails",props:{zoneEgressOverview:{type:Object,required:!0}},setup(h){const o=h,{t:y}=k(),z=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Egress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],u=p(()=>({name:"zone-egress-detail-view",params:{zoneEgress:o.zoneEgressOverview.name}})),a=p(()=>{const{type:r,name:i}=o.zoneEgressOverview;return{type:r,name:i}}),_=p(()=>{var i;const r=((i=o.zoneEgressOverview.zoneEgressInsight)==null?void 0:i.subscriptions)??[];return Array.from(r).reverse()});return(r,i)=>{const $=F("router-link");return t(),n(T,{tabs:z},{tabHeader:e(()=>[R("h1",j,[m(`
        Zone Egress:

        `),s(w,{text:a.value.name},{default:e(()=>[s($,{to:u.value},{default:e(()=>[m(E(a.value.name),1)]),_:1},8,["to"])]),_:1},8,["text"])])]),overview:e(()=>[s(C,null,{default:e(()=>[(t(!0),d(g,null,b(a.value,(l,c)=>(t(),n(S,{key:c,term:f(y)(`http.api.property.${c}`)},{default:e(()=>[c==="name"?(t(),n(w,{key:0,text:l},null,8,["text"])):(t(),d(g,{key:1},[m(E(l),1)],64))]),_:2},1032,["term"]))),128))]),_:1})]),insights:e(()=>[s(D,{"initially-open":0},{default:e(()=>[(t(!0),d(g,null,b(_.value,(l,c)=>(t(),n(A,{key:c},{"accordion-header":e(()=>[s(B,{details:l},null,8,["details"])]),"accordion-content":e(()=>[s(O,{details:l,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),"xds-configuration":e(()=>[s(v,{"data-path":"xds","zone-egress-name":a.value.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),"envoy-stats":e(()=>[s(v,{"data-path":"stats","zone-egress-name":a.value.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),"envoy-clusters":e(()=>[s(v,{"data-path":"clusters","zone-egress-name":a.value.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),_:1})}}}),X={key:3,class:"kcard-border","data-testid":"detail-view-details"},te=x({__name:"ZoneEgressDetailView",setup(h){const{t:o}=k();return(y,z)=>(t(),n(N,{name:"zone-egress-detail-view"},{default:e(({route:u})=>[s(W,{title:f(o)("zone-egresses.routes.item.title",{name:u.params.zoneEgress})},null,8,["title"]),m(),s(V,{breadcrumbs:[{to:{name:"zone-egress-list-view"},text:f(o)("zone-egresses.routes.item.breadcrumbs")}]},{default:e(()=>[s(q,{src:`/zone-egresses/${u.params.zoneEgress}`},{default:e(({data:a,isLoading:_,error:r})=>[_?(t(),n(I,{key:0})):r!==void 0?(t(),n(Z,{key:1,error:r},null,8,["error"])):a===void 0?(t(),n(L,{key:2})):(t(),d("div",X,[s(H,{"zone-egress-overview":a},null,8,["zone-egress-overview"])]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{te as default};
