import{d as w,c as u,U as N,o as p,e as V,a as d,f as I,g as e,h as a,w as t,i as v,t as r,b as o,l as y,s as A}from"./index-f1b8ae6a.js";import{h as B,w as D,v as E,D as m,S as x,A as T,_ as Z}from"./RouteView.vue_vue_type_script_setup_true_lang-4a32e1ca.js";import{_ as $}from"./CodeBlock.vue_vue_type_style_index_0_lang-d1d1c408.js";import{_ as J}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-6f9420ad.js";import{f as L}from"./dataplane-e7ae9fed.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-14dd845b.js";const M={class:"stack"},P={class:"columns",style:{"--columns":"3"}},j=w({__name:"ZoneDetails",props:{zoneOverview:{type:Object,required:!0}},setup(_){const i=_,{t:l}=B(),f=D(),z=u(()=>{var n;for(const s of((n=i.zoneOverview.zoneInsight)==null?void 0:n.subscriptions)??[])if(s.config)return JSON.parse(s.config).environment;return"kubernetes"}),O=u(()=>E(i.zoneOverview)),k=u(()=>N(i.zoneOverview)),g=u(()=>{var c;const n=[],s=((c=i.zoneOverview.zoneInsight)==null?void 0:c.subscriptions)??[];if(s.length>0){const b=s[s.length-1],C=b.version.kumaCp.version||"-",{kumaCpGlobalCompatible:S=!0}=b.version.kumaCp;S||n.push({kind:L,payload:{zoneCpVersion:C,globalCpVersion:f("KUMA_VERSION")}})}return n}),h=u(()=>{var s;const n=((s=i.zoneOverview.zoneInsight)==null?void 0:s.subscriptions)??[];if(n.length>0){const c=n[n.length-1];if(c.config)return JSON.stringify(JSON.parse(c.config),null,2)}return null});return(n,s)=>(p(),V("div",M,[g.value.length>0?(p(),d(J,{key:0,warnings:g.value},null,8,["warnings"])):I("",!0),e(),a(o(y),null,{body:t(()=>[v("div",P,[a(m,null,{title:t(()=>[e(r(o(l)("http.api.property.status")),1)]),body:t(()=>[a(x,{status:O.value},null,8,["status"])]),_:1}),e(),a(m,null,{title:t(()=>[e(r(o(l)("http.api.property.type")),1)]),body:t(()=>[e(r(z.value),1)]),_:1}),e(),a(m,null,{title:t(()=>[e(r(o(l)("http.api.property.authenticationType")),1)]),body:t(()=>[e(r(k.value),1)]),_:1})])]),_:1}),e(),v("div",null,[v("h2",null,r(o(l)("zone-cps.detail.configuration_title")),1),e(),a(o(y),{class:"mt-4"},{body:t(()=>[h.value!==null?(p(),d($,{key:0,id:"code-block-zone-config",language:"json",code:h.value,"is-searchable":"","query-key":"zone-config"},null,8,["code"])):(p(),d(o(A),{key:1,class:"mt-4","data-testid":"warning-no-subscriptions",appearance:"warning"},{alertMessage:t(()=>[e(r(o(l)("zone-cps.detail.no_subscriptions")),1)]),_:1}))]),_:1})])]))}}),H=w({__name:"ZoneDetailView",props:{data:{}},setup(_){const i=_;return(l,f)=>(p(),d(Z,{name:"zone-cp-detail-view","data-testid":"zone-cp-detail-view"},{default:t(()=>[a(T,null,{default:t(()=>[a(j,{"zone-overview":i.data,"data-testid":"detail-view-details"},null,8,["zone-overview"])]),_:1})]),_:1}))}});export{H as default};
