import{h as B,l as V,e as $,D as d,S as K,k as L,E as z,n as F,i as U,f as Z,A as j,_ as q}from"./RouteView.vue_vue_type_script_setup_true_lang-4a32e1ca.js";import{d as x,u as G,c as _,K as W,o as l,e as y,a as p,f as A,g as e,h as i,w as t,i as u,t as n,b as a,q as H,m as J,F as w,l as E,s as Q}from"./index-f1b8ae6a.js";import{_ as X}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-be27b925.js";import{T as N}from"./TagList-386ee57e.js";import{_ as Y}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-6f9420ad.js";import{g as aa,d as P,a as ea,p as ta,c as sa,C as na,I as la,b as ia}from"./dataplane-e7ae9fed.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-14dd845b.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-d1d1c408.js";import"./toYaml-4e00099e.js";const oa={class:"stack"},ra={class:"columns",style:{"--columns":"3"}},pa={class:"status-with-reason"},ca=["href"],ua={class:"columns",style:{"--columns":"3"}},da=x({__name:"DataPlaneDetails",props:{dataplaneOverview:{type:Object,required:!0}},setup(k){const o=k,{t:s,formatIsoDate:D}=B(),M=V(),O=G(),S=$(),b=_(()=>aa(o.dataplaneOverview.dataplane,o.dataplaneOverview.dataplaneInsight)),C=_(()=>P(o.dataplaneOverview.dataplane)),I=_(()=>ea(o.dataplaneOverview.dataplaneInsight)),m=_(()=>ta(o.dataplaneOverview,D)),T=_(()=>{var g;const v=((g=o.dataplaneOverview.dataplaneInsight)==null?void 0:g.subscriptions)??[];if(v.length===0)return[];const f=v[v.length-1];if(!("version"in f)||!f.version)return[];const c=[],r=f.version;if(r.kumaDp&&r.envoy){const h=sa(r);h.kind!==na&&h.kind!==la&&c.push(h)}return S.getters["config/getMulticlusterStatus"]&&P(o.dataplaneOverview.dataplane).find(R=>R.label===W)&&typeof r.kumaDp.kumaCpCompatible=="boolean"&&!r.kumaDp.kumaCpCompatible&&c.push({kind:ia,payload:{kumaDp:r.kumaDp.version}}),c});return(v,f)=>(l(),y("div",oa,[T.value.length>0?(l(),p(Y,{key:0,warnings:T.value,"data-testid":"data-plane-warnings"},null,8,["warnings"])):A("",!0),e(),i(a(E),null,{body:t(()=>[u("div",ra,[i(d,null,{title:t(()=>[e(n(a(s)("http.api.property.status")),1)]),body:t(()=>[u("div",pa,[i(K,{status:b.value.status},null,8,["status"]),e(),b.value.reason.length>0?(l(),p(a(H),{key:0,label:b.value.reason.join(", "),class:"reason-tooltip"},{default:t(()=>[i(a(J),{icon:"info",size:a(L),"hide-title":""},null,8,["size"])]),_:1},8,["label"])):A("",!0)])]),_:1}),e(),i(d,null,{title:t(()=>[e(n(a(s)("http.api.property.tags")),1)]),body:t(()=>[C.value.length>0?(l(),p(N,{key:0,tags:C.value},null,8,["tags"])):(l(),y(w,{key:1},[e(n(a(s)("common.detail.none")),1)],64))]),_:1}),e(),i(d,null,{title:t(()=>[e(n(a(s)("http.api.property.dependencies")),1)]),body:t(()=>[I.value!==null?(l(),p(N,{key:0,tags:I.value},null,8,["tags"])):(l(),y(w,{key:1},[e(n(a(s)("common.detail.none")),1)],64))]),_:1})])]),_:1}),e(),u("div",null,[u("h3",null,n(a(s)("data-planes.detail.mtls")),1),e(),m.value===null?(l(),p(a(Q),{key:0,class:"mt-4",appearance:"danger"},{alertMessage:t(()=>[e(n(a(s)("data-planes.detail.no_mtls"))+` —
          `,1),u("a",{href:a(s)("data-planes.href.docs.mutual-tls"),class:"external-link",target:"_blank"},n(a(s)("data-planes.detail.no_mtls_learn_more",{product:a(s)("common.product.name")})),9,ca)]),_:1})):(l(),p(a(E),{key:1,class:"mt-4"},{body:t(()=>[u("div",ua,[i(d,null,{title:t(()=>[e(n(a(s)("http.api.property.certificateExpirationTime")),1)]),body:t(()=>[e(n(m.value.certificateExpirationTime),1)]),_:1}),e(),i(d,null,{title:t(()=>[e(n(a(s)("http.api.property.lastCertificateRegeneration")),1)]),body:t(()=>[e(n(m.value.lastCertificateRegeneration),1)]),_:1}),e(),i(d,null,{title:t(()=>[e(n(a(s)("http.api.property.certificateRegenerations")),1)]),body:t(()=>[e(n(m.value.certificateRegenerations),1)]),_:1})])]),_:1}))]),e(),u("div",null,[i(U,{src:`/meshes/${a(O).params.mesh}/dataplanes/${a(O).params.dataPlane}`},{default:t(({data:c,error:r})=>[r?(l(),p(z,{key:0,error:r},null,8,["error"])):c===void 0?(l(),p(F,{key:1})):(l(),y(w,{key:2},[u("h3",null,n(a(s)("data-planes.detail.configuration")),1),e(),i(X,{id:"code-block-data-plane",class:"mt-4",resource:c,"resource-fetcher":g=>a(M).getDataplaneFromMesh({mesh:c.mesh,name:c.name},g),"is-searchable":""},null,8,["resource","resource-fetcher"])],64))]),_:1},8,["src"])])]))}});const _a=Z(da,[["__scopeId","data-v-48b6484f"]]),Oa=x({__name:"DataPlaneDetailView",props:{data:{}},setup(k){const o=k;return(s,D)=>(l(),p(q,{name:"data-plane-detail-view","data-testid":"data-plane-detail-view"},{default:t(()=>[i(j,null,{default:t(()=>[i(_a,{"dataplane-overview":o.data,"data-testid":"detail-view-details"},null,8,["dataplane-overview"])]),_:1})]),_:1}))}});export{Oa as default};
