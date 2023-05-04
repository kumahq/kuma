import{A as z,a as I}from"./AccordionList-68fd7c69.js";import{D as b,a as w}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-d6e052e1.js";import{E as l}from"./EnvoyData-3e492041.js";import{_ as D,S as k}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-22eb5fc3.js";import{T as x}from"./TabsWidget-59c9beec.js";import{d as O,c as m,e as u,w as e,o as r,i as S,t as g,b as n,g as s,j as p,q as h,h as q,F as v}from"./index-bd38c154.js";const A={class:"entity-heading"},N=O({__name:"ZoneIngressDetails",props:{zoneIngressOverview:{type:Object,required:!0}},setup(y){const d=y,f=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Ingress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],t=m(()=>{const{type:o,name:a}=d.zoneIngressOverview;return{type:o,name:a}}),_=m(()=>{var a;const o=((a=d.zoneIngressOverview.zoneIngressInsight)==null?void 0:a.subscriptions)??[];return Array.from(o).reverse()});return(o,a)=>(r(),u(x,{tabs:f},{tabHeader:e(()=>[S("h1",A,`
        Zone Ingress: `+g(n(t).name),1)]),overview:e(()=>[s(w,null,{default:e(()=>[(r(!0),p(v,null,h(n(t),(i,c)=>(r(),u(b,{key:c,term:c},{default:e(()=>[q(g(i),1)]),_:2},1032,["term"]))),128))]),_:1})]),insights:e(()=>[s(I,{"initially-open":0},{default:e(()=>[(r(!0),p(v,null,h(n(_),(i,c)=>(r(),u(z,{key:c},{"accordion-header":e(()=>[s(D,{details:i},null,8,["details"])]),"accordion-content":e(()=>[s(k,{details:i,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),"xds-configuration":e(()=>[s(l,{"data-path":"xds","zone-ingress-name":n(t).name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-stats":e(()=>[s(l,{"data-path":"stats","zone-ingress-name":n(t).name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-clusters":e(()=>[s(l,{"data-path":"clusters","zone-ingress-name":n(t).name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),_:1}))}});export{N as _};
