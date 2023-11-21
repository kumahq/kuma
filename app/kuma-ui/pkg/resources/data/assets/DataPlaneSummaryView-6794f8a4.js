import{d as I,l as P,O as b,a as m,o as i,c as _,e as n,w as t,f as e,t as o,q as a,p as l,b as v,a2 as $,v as f,a1 as u,F as k,C as B,_ as V,G as R,A as T,B as C,a4 as K}from"./index-784d2bbf.js";import{_ as A}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-f6a2a033.js";import{K as N}from"./index-9dd3e7d3.js";import{g as L}from"./data-207af457.js";import{S as F}from"./StatusBadge-a6acfbee.js";import{_ as S}from"./TagList.vue_vue_type_script_setup_true_lang-115cdf7e.js";import{T as x}from"./TextWithCopyButton-7ef74197.js";import{a as U}from"./dataplane-dcd0858b.js";import"./CopyButton-9c00109a.js";const z={class:"stack"},E={class:"status-with-reason"},W={key:0},j={class:"mt-4"},q={class:"stack"},G={key:1},Z={class:"mt-4"},H={class:"inbound-list"},J={class:"mt-2 stack"},M=I({__name:"DataPlaneSummary",props:{dataplaneOverview:{}},setup(c){const{t:s}=P(),r=c,p=b(()=>U(r.dataplaneOverview.dataplane,r.dataplaneOverview.dataplaneInsight)),w=b(()=>{var y;return L(((y=r.dataplaneOverview.dataplaneInsight)==null?void 0:y.subscriptions)??[])});return(y,O)=>{const g=m("KTooltip"),h=m("KBadge");return i(),_("div",z,[n(u,null,{title:t(()=>[e(o(a(s)("http.api.property.status")),1)]),body:t(()=>[l("div",E,[n(F,{status:p.value.status},null,8,["status"]),e(),p.value.reason.length>0?(i(),v(g,{key:0,label:p.value.reason.join(", "),class:"reason-tooltip","position-fixed":""},{default:t(()=>[n(a($),{size:a(N),"hide-title":""},null,8,["size"])]),_:1},8,["label"])):f("",!0)])]),_:1}),e(),n(u,null,{title:t(()=>[e(o(a(s)("data-planes.routes.item.last_updated")),1)]),body:t(()=>[w.value?(i(),_(k,{key:0},[e(o(w.value),1)],64)):(i(),_(k,{key:1},[e(o(a(s)("common.detail.none")),1)],64))]),_:1}),e(),r.dataplaneOverview.dataplane.networking.gateway?(i(),_("div",W,[l("h3",null,o(a(s)("data-planes.routes.item.gateway")),1),e(),l("div",j,[l("div",q,[n(u,null,{title:t(()=>[e(o(a(s)("http.api.property.address")),1)]),body:t(()=>[n(x,{text:`${r.dataplaneOverview.dataplane.networking.address}`},null,8,["text"])]),_:1}),e(),n(u,null,{title:t(()=>[e(o(a(s)("http.api.property.tags")),1)]),body:t(()=>[n(S,{tags:r.dataplaneOverview.dataplane.networking.gateway.tags},null,8,["tags"])]),_:1})])])])):f("",!0),e(),r.dataplaneOverview.dataplane.networking.inbound?(i(),_("div",G,[l("h3",null,o(a(s)("data-planes.routes.item.inbounds")),1),e(),l("div",Z,[l("div",H,[(i(!0),_(k,null,B(r.dataplaneOverview.dataplane.networking.inbound,(d,D)=>(i(),_("div",{key:D,class:"inbound"},[l("h4",null,[n(x,{text:d.tags["kuma.io/service"]},{default:t(()=>[e(o(a(s)("data-planes.routes.item.inbound_name",{service:d.tags["kuma.io/service"]})),1)]),_:2},1032,["text"])]),e(),l("div",J,[n(u,null,{title:t(()=>[e(o(a(s)("http.api.property.status")),1)]),body:t(()=>[!d.health||d.health.ready?(i(),v(h,{key:0,appearance:"success"},{default:t(()=>[e(o(a(s)("data-planes.routes.item.health.ready")),1)]),_:1})):(i(),v(h,{key:1,appearance:"danger"},{default:t(()=>[e(o(a(s)("data-planes.routes.item.health.not_ready")),1)]),_:1}))]),_:2},1024),e(),n(u,null,{title:t(()=>[e(o(a(s)("http.api.property.address")),1)]),body:t(()=>[n(x,{text:`${d.address??r.dataplaneOverview.dataplane.networking.advertisedAddress??r.dataplaneOverview.dataplane.networking.address}:${d.port}`},null,8,["text"])]),_:2},1024),e(),n(u,null,{title:t(()=>[e(o(a(s)("http.api.property.tags")),1)]),body:t(()=>[n(S,{tags:d.tags},null,8,["tags"])]),_:2},1024)])]))),128))])])])):f("",!0)])}}});const Q=V(M,[["__scopeId","data-v-a0c567ee"]]),X=c=>(T("data-v-b50681c1"),c=c(),C(),c),Y={class:"summary-title-wrapper"},ee=X(()=>l("img",{"aria-hidden":"true",src:K},null,-1)),te={class:"summary-title"},ae={key:1,class:"stack"},se=I({__name:"DataPlaneSummaryView",props:{name:{},dataplaneOverview:{default:void 0}},setup(c){const{t:s}=P(),r=R(),p=c;return(w,y)=>{const O=m("RouteTitle"),g=m("RouterLink"),h=m("AppView"),d=m("RouteView");return i(),v(d,{name:a(r).name},{default:t(()=>[n(h,null,{title:t(()=>[l("div",Y,[ee,e(),l("h2",te,[n(g,{to:{name:"data-plane-detail-view",params:{dataPlane:p.name}}},{default:t(()=>[n(O,{title:a(s)("data-planes.routes.item.title",{name:p.name})},null,8,["title"])]),_:1},8,["to"])])])]),default:t(()=>[e(),p.dataplaneOverview===void 0?(i(),v(A,{key:0},{message:t(()=>[l("p",null,o(a(s)("common.collection.summary.empty_message",{type:"Data Plane Proxy"})),1)]),default:t(()=>[e(o(a(s)("common.collection.summary.empty_title",{type:"Data Plane Proxy"}))+" ",1)]),_:1})):(i(),_("div",ae,[l("div",null,[l("h3",null,o(a(s)("data-planes.routes.item.overview")),1),e(),n(Q,{class:"mt-4","dataplane-overview":p.dataplaneOverview},null,8,["dataplane-overview"])])]))]),_:1})]),_:1},8,["name"])}}});const ue=V(se,[["__scopeId","data-v-b50681c1"]]);export{ue as default};
