import{d as z,l as k,o as n,c as _,e as t,w as e,p as u,t as a,f as s,Z as v,W as h,b as d,F as x,a as l,m as c,a2 as V,_ as O}from"./index-i9Uphcre.js";import{S as B}from"./StatusBadge-TCGLRU6a.js";const R={class:"stack"},S=z({__name:"ZoneEgressSummary",props:{zoneEgressOverview:{}},setup(m){const{t:o}=k(),r=m;return(w,g)=>(n(),_("div",R,[t(v,null,{title:e(()=>[s(a(u(o)("http.api.property.status")),1)]),body:e(()=>[t(B,{status:r.zoneEgressOverview.state},null,8,["status"])]),_:1}),s(),t(v,null,{title:e(()=>[s(a(u(o)("http.api.property.address")),1)]),body:e(()=>[r.zoneEgressOverview.zoneEgress.socketAddress.length>0?(n(),d(h,{key:0,text:r.zoneEgressOverview.zoneEgress.socketAddress},null,8,["text"])):(n(),_(x,{key:1},[s(a(u(o)("common.detail.none")),1)],64))]),_:1})]))}}),Z={key:1,class:"stack"},C=z({__name:"ZoneEgressSummaryView",props:{zoneEgressOverview:{default:void 0}},setup(m){const o=m;return(r,w)=>{const g=l("RouteTitle"),y=l("RouterLink"),E=l("AppView"),f=l("RouteView");return n(),d(f,{name:"zone-egress-summary-view",params:{zone:"",zoneEgress:""}},{default:e(({route:p,t:i})=>[t(E,null,{title:e(()=>[c("h2",null,[t(y,{to:{name:"zone-egress-detail-view",params:{zone:p.params.zone,zoneEgress:p.params.zoneEgress}}},{default:e(()=>[t(g,{title:i("zone-egresses.routes.item.title",{name:p.params.zoneEgress})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[s(),o.zoneEgressOverview===void 0?(n(),d(V,{key:0},{title:e(()=>[s(a(i("common.collection.summary.empty_title",{type:"ZoneEgress"})),1)]),default:e(()=>[s(),c("p",null,a(i("common.collection.summary.empty_message",{type:"ZoneEgress"})),1)]),_:2},1024)):(n(),_("div",Z,[c("div",null,[c("h3",null,a(i("zone-egresses.routes.item.overview")),1),s(),t(S,{class:"mt-4","zone-egress-overview":o.zoneEgressOverview},null,8,["zone-egress-overview"])])]))]),_:2},1024)]),_:1})}}}),b=O(C,[["__scopeId","data-v-97456a15"]]);export{b as default};
