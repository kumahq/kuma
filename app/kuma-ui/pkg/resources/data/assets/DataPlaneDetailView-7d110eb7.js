import{d as x,L as R,M as V,f as _,ab as B,o,j as h,g as d,k as w,l as a,h as l,w as t,m as p,ac as u,D as n,i as e,Y as S,ad as L,H as z,K,F as N,a3 as k,ae as $,q as j,A as F,_ as U}from"./index-18fd9432.js";import{_ as Z}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-fcbabf84.js";import{T as A}from"./TagList-01d4a4cd.js";import{_ as q}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-5b861cbc.js";import{a as G,d as P,b as W,p as H,c as Y,C as J,I as Q,e as X}from"./dataplane-30467516.js";import"./AccordionList-0f5fb797.js";const aa={class:"stack"},ea={class:"columns",style:{"--columns":"3"}},ta={class:"status-with-reason"},sa=["href"],na={class:"columns",style:{"--columns":"3"}},la={key:1},ia=x({__name:"DataPlaneDetails",props:{dataplaneOverview:{type:Object,required:!0}},setup(y){const i=y,{t:s,formatIsoDate:D}=R(),E=V(),b=_(()=>G(i.dataplaneOverview.dataplane,i.dataplaneOverview.dataplaneInsight)),O=_(()=>P(i.dataplaneOverview.dataplane)),I=_(()=>W(i.dataplaneOverview.dataplaneInsight)),m=_(()=>H(i.dataplaneOverview,D)),C=_(()=>{var T;const v=((T=i.dataplaneOverview.dataplaneInsight)==null?void 0:T.subscriptions)??[];if(v.length===0)return[];const g=v[v.length-1];if(!("version"in g)||!g.version)return[];const c=[],r=g.version;if(r.kumaDp&&r.envoy){const f=Y(r);f.kind!==J&&f.kind!==Q&&c.push(f)}return E("use zones")&&P(i.dataplaneOverview.dataplane).find(M=>M.label===B)&&typeof r.kumaDp.kumaCpCompatible=="boolean"&&!r.kumaDp.kumaCpCompatible&&c.push({kind:X,payload:{kumaDp:r.kumaDp.version}}),c});return(v,g)=>{var c;return o(),h("div",aa,[C.value.length>0?(o(),d(q,{key:0,warnings:C.value,"data-testid":"data-plane-warnings"},null,8,["warnings"])):w("",!0),a(),l(e(k),null,{body:t(()=>[p("div",ea,[l(u,null,{title:t(()=>[a(n(e(s)("http.api.property.status")),1)]),body:t(()=>[p("div",ta,[l(S,{status:b.value.status},null,8,["status"]),a(),b.value.reason.length>0?(o(),d(e(L),{key:0,label:b.value.reason.join(", "),class:"reason-tooltip"},{default:t(()=>[l(e(z),{icon:"info",size:e(K),"hide-title":""},null,8,["size"])]),_:1},8,["label"])):w("",!0)])]),_:1}),a(),l(u,null,{title:t(()=>[a(n(e(s)("http.api.property.tags")),1)]),body:t(()=>[O.value.length>0?(o(),d(A,{key:0,tags:O.value},null,8,["tags"])):(o(),h(N,{key:1},[a(n(e(s)("common.detail.none")),1)],64))]),_:1}),a(),l(u,null,{title:t(()=>[a(n(e(s)("http.api.property.dependencies")),1)]),body:t(()=>[I.value!==null?(o(),d(A,{key:0,tags:I.value},null,8,["tags"])):(o(),h(N,{key:1},[a(n(e(s)("common.detail.none")),1)],64))]),_:1})])]),_:1}),a(),p("div",null,[p("h2",null,n(e(s)("data-planes.detail.mtls")),1),a(),m.value===null?(o(),d(e($),{key:0,class:"mt-4",appearance:"warning"},{alertMessage:t(()=>[a(n(e(s)("data-planes.detail.no_mtls"))+` —
          `,1),p("a",{href:e(s)("data-planes.href.docs.mutual-tls"),class:"external-link",target:"_blank"},n(e(s)("data-planes.detail.no_mtls_learn_more",{product:e(s)("common.product.name")})),9,sa)]),_:1})):(o(),d(e(k),{key:1,class:"mt-4"},{body:t(()=>[p("div",na,[l(u,null,{title:t(()=>[a(n(e(s)("http.api.property.certificateExpirationTime")),1)]),body:t(()=>[a(n(m.value.certificateExpirationTime),1)]),_:1}),a(),l(u,null,{title:t(()=>[a(n(e(s)("http.api.property.lastCertificateRegeneration")),1)]),body:t(()=>[a(n(m.value.lastCertificateRegeneration),1)]),_:1}),a(),l(u,null,{title:t(()=>[a(n(e(s)("http.api.property.certificateRegenerations")),1)]),body:t(()=>[a(n(m.value.certificateRegenerations),1)]),_:1})])]),_:1}))]),a(),(((c=i.dataplaneOverview.dataplaneInsight)==null?void 0:c.subscriptions)??[]).length>0?(o(),h("div",la,[p("h2",null,n(e(s)("data-planes.detail.subscriptions")),1),a(),l(e(k),{class:"mt-4"},{body:t(()=>{var r;return[l(Z,{subscriptions:((r=i.dataplaneOverview.dataplaneInsight)==null?void 0:r.subscriptions)??[]},null,8,["subscriptions"])]}),_:1})])):w("",!0)])}}});const oa=j(ia,[["__scopeId","data-v-e0adad19"]]),va=x({__name:"DataPlaneDetailView",props:{data:{}},setup(y){const i=y;return(s,D)=>(o(),d(U,{name:"data-plane-detail-view","data-testid":"data-plane-detail-view"},{default:t(()=>[l(F,null,{default:t(()=>[l(oa,{"dataplane-overview":i.data,"data-testid":"detail-view-details"},null,8,["dataplane-overview"])]),_:1})]),_:1}))}});export{va as default};
