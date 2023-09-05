import{g as R,D as u,S,K as V,f as B,A as L,_ as z}from"./RouteView.vue_vue_type_script_setup_true_lang-da83f5a8.js";import{d as P,y as K,c as _,L as $,o,e as y,a as d,f as w,g as a,h as i,w as t,i as p,t as n,b as e,M as F,x as U,F as N,k,N as Z}from"./index-9a3d231d.js";import{_ as j}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-11478281.js";import{T as x}from"./TagList-dc4b1e54.js";import{_ as G}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-b0b1c765.js";import{a as W,d as A,b as q,p as H,c as J,C as Q,I as X,e as Y}from"./dataplane-30467516.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-fe937ad6.js";import"./AccordionList-aa0b3f52.js";const aa={class:"stack"},ea={class:"columns",style:{"--columns":"3"}},ta={class:"status-with-reason"},sa=["href"],na={class:"columns",style:{"--columns":"3"}},ia={key:1},la=P({__name:"DataPlaneDetails",props:{dataplaneOverview:{type:Object,required:!0}},setup(h){const l=h,{t:s,formatIsoDate:D}=R(),E=K(),b=_(()=>W(l.dataplaneOverview.dataplane,l.dataplaneOverview.dataplaneInsight)),O=_(()=>A(l.dataplaneOverview.dataplane)),I=_(()=>q(l.dataplaneOverview.dataplaneInsight)),m=_(()=>H(l.dataplaneOverview,D)),C=_(()=>{var T;const v=((T=l.dataplaneOverview.dataplaneInsight)==null?void 0:T.subscriptions)??[];if(v.length===0)return[];const g=v[v.length-1];if(!("version"in g)||!g.version)return[];const c=[],r=g.version;if(r.kumaDp&&r.envoy){const f=J(r);f.kind!==Q&&f.kind!==X&&c.push(f)}return E("use zones")&&A(l.dataplaneOverview.dataplane).find(M=>M.label===$)&&typeof r.kumaDp.kumaCpCompatible=="boolean"&&!r.kumaDp.kumaCpCompatible&&c.push({kind:Y,payload:{kumaDp:r.kumaDp.version}}),c});return(v,g)=>{var c;return o(),y("div",aa,[C.value.length>0?(o(),d(G,{key:0,warnings:C.value,"data-testid":"data-plane-warnings"},null,8,["warnings"])):w("",!0),a(),i(e(k),null,{body:t(()=>[p("div",ea,[i(u,null,{title:t(()=>[a(n(e(s)("http.api.property.status")),1)]),body:t(()=>[p("div",ta,[i(S,{status:b.value.status},null,8,["status"]),a(),b.value.reason.length>0?(o(),d(e(F),{key:0,label:b.value.reason.join(", "),class:"reason-tooltip"},{default:t(()=>[i(e(U),{icon:"info",size:e(V),"hide-title":""},null,8,["size"])]),_:1},8,["label"])):w("",!0)])]),_:1}),a(),i(u,null,{title:t(()=>[a(n(e(s)("http.api.property.tags")),1)]),body:t(()=>[O.value.length>0?(o(),d(x,{key:0,tags:O.value},null,8,["tags"])):(o(),y(N,{key:1},[a(n(e(s)("common.detail.none")),1)],64))]),_:1}),a(),i(u,null,{title:t(()=>[a(n(e(s)("http.api.property.dependencies")),1)]),body:t(()=>[I.value!==null?(o(),d(x,{key:0,tags:I.value},null,8,["tags"])):(o(),y(N,{key:1},[a(n(e(s)("common.detail.none")),1)],64))]),_:1})])]),_:1}),a(),p("div",null,[p("h2",null,n(e(s)("data-planes.detail.mtls")),1),a(),m.value===null?(o(),d(e(Z),{key:0,class:"mt-4",appearance:"warning"},{alertMessage:t(()=>[a(n(e(s)("data-planes.detail.no_mtls"))+` —
          `,1),p("a",{href:e(s)("data-planes.href.docs.mutual-tls"),class:"external-link",target:"_blank"},n(e(s)("data-planes.detail.no_mtls_learn_more",{product:e(s)("common.product.name")})),9,sa)]),_:1})):(o(),d(e(k),{key:1,class:"mt-4"},{body:t(()=>[p("div",na,[i(u,null,{title:t(()=>[a(n(e(s)("http.api.property.certificateExpirationTime")),1)]),body:t(()=>[a(n(m.value.certificateExpirationTime),1)]),_:1}),a(),i(u,null,{title:t(()=>[a(n(e(s)("http.api.property.lastCertificateRegeneration")),1)]),body:t(()=>[a(n(m.value.lastCertificateRegeneration),1)]),_:1}),a(),i(u,null,{title:t(()=>[a(n(e(s)("http.api.property.certificateRegenerations")),1)]),body:t(()=>[a(n(m.value.certificateRegenerations),1)]),_:1})])]),_:1}))]),a(),(((c=l.dataplaneOverview.dataplaneInsight)==null?void 0:c.subscriptions)??[]).length>0?(o(),y("div",ia,[p("h2",null,n(e(s)("data-planes.detail.subscriptions")),1),a(),i(e(k),{class:"mt-4"},{body:t(()=>{var r;return[i(j,{subscriptions:((r=l.dataplaneOverview.dataplaneInsight)==null?void 0:r.subscriptions)??[]},null,8,["subscriptions"])]}),_:1})])):w("",!0)])}}});const oa=B(la,[["__scopeId","data-v-e0adad19"]]),fa=P({__name:"DataPlaneDetailView",props:{data:{}},setup(h){const l=h;return(s,D)=>(o(),d(z,{name:"data-plane-detail-view","data-testid":"data-plane-detail-view"},{default:t(()=>[i(L,null,{default:t(()=>[i(oa,{"dataplane-overview":l.data,"data-testid":"detail-view-details"},null,8,["dataplane-overview"])]),_:1})]),_:1}))}});export{fa as default};
