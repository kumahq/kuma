import{d as h,l as w,M as y,o as l,c as d,e as r,w as t,f as a,t as m,q as i,a1 as f,b as v,F as x,a as u,p as c,y as S,z as V,aD as k,_ as I}from"./index-7fe6d41d.js";import{S as O}from"./StatusBadge-f4811259.js";import{T as b}from"./TextWithCopyButton-4896f22f.js";import{_ as T}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-6f5efa76.js";import"./CopyButton-3c266137.js";import"./index-fce48c05.js";function B(o={subscriptions:[]}){return(o.subscriptions??[]).some(s=>{var p;return((p=s.connectTime)==null?void 0:p.length)&&!s.disconnectTime})?"online":"offline"}const R={class:"stack"},Z=h({__name:"ZoneEgressSummary",props:{zoneEgressOverview:{}},setup(o){const{t:n}=w(),s=o,p=y(()=>B(s.zoneEgressOverview.zoneEgressInsight)),_=y(()=>{const{networking:e}=s.zoneEgressOverview.zoneEgress;return e!=null&&e.address&&(e!=null&&e.port)?`${e.address}:${e.port}`:null});return(e,g)=>(l(),d("div",R,[r(f,null,{title:t(()=>[a(m(i(n)("http.api.property.status")),1)]),body:t(()=>[r(O,{status:p.value},null,8,["status"])]),_:1}),a(),r(f,null,{title:t(()=>[a(m(i(n)("http.api.property.address")),1)]),body:t(()=>[_.value?(l(),v(b,{key:0,text:_.value},null,8,["text"])):(l(),d(x,{key:1},[a(m(i(n)("common.detail.none")),1)],64))]),_:1})]))}}),C=o=>(S("data-v-b8c65a14"),o=o(),V(),o),$={class:"summary-title-wrapper"},D=C(()=>c("img",{"aria-hidden":"true",src:k},null,-1)),F={class:"summary-title"},N={key:1,class:"stack"},A=h({__name:"ZoneEgressSummaryView",props:{name:{},zoneEgressOverview:{default:void 0}},setup(o){const{t:n}=w(),s=o;return(p,_)=>{const e=u("RouteTitle"),g=u("RouterLink"),z=u("AppView"),E=u("RouteView");return l(),v(E,{name:"zone-egress-summary-view"},{default:t(()=>[r(z,null,{title:t(()=>[c("div",$,[D,a(),c("h2",F,[r(g,{to:{name:"zone-egress-detail-view",params:{zone:s.name}}},{default:t(()=>[r(e,{title:i(n)("zone-egresses.routes.item.title",{name:s.name})},null,8,["title"])]),_:1},8,["to"])])])]),default:t(()=>[a(),s.zoneEgressOverview===void 0?(l(),v(T,{key:0},{message:t(()=>[c("p",null,m(i(n)("common.collection.summary.empty_message",{type:"ZoneEgress"})),1)]),default:t(()=>[a(m(i(n)("common.collection.summary.empty_title",{type:"ZoneEgress"}))+" ",1)]),_:1})):(l(),d("div",N,[c("div",null,[c("h3",null,m(i(n)("zone-egresses.routes.item.overview")),1),a(),r(Z,{class:"mt-4","zone-egress-overview":s.zoneEgressOverview},null,8,["zone-egress-overview"])])]))]),_:1})]),_:1})}}});const H=I(A,[["__scopeId","data-v-b8c65a14"]]);export{H as default};
