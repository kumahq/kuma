import{K as C}from"./index-fce48c05.js";import{d as V,a as r,o as a,b as l,w as e,e as n,p as x,f as i,t as p,c as h,F as S,q as g,T as A,D as b,s as z,_ as R}from"./index-7a0947c2.js";import{A as B}from"./AppCollection-1fb4918a.js";import{E as D}from"./ErrorBlock-78880c60.js";import{S as L}from"./StatusBadge-c02c8868.js";import{S as N}from"./SummaryView-4b76de54.js";import{T}from"./TextWithCopyButton-3aa03737.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-4198f65d.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-1c689249.js";import"./CopyButton-a5c25cdd.js";const Z=V({__name:"ZoneEgressListView",setup(I){return(K,$)=>{const y=r("RouteTitle"),_=r("RouterLink"),f=r("KCard"),k=r("RouterView"),d=r("DataSource"),w=r("AppView"),E=r("RouteView");return a(),l(d,{src:"/me"},{default:e(({data:v})=>[v?(a(),l(E,{key:0,name:"zone-egress-list-view",params:{zone:"",zoneEgress:""}},{default:e(({route:t,t:m})=>[n(w,null,{title:e(()=>[x("h2",null,[n(y,{title:m("zone-egresses.routes.items.title")},null,8,["title"])])]),default:e(()=>[i(),n(d,{src:`/zone-cps/${t.params.zone||"*"}/egresses?page=1&size=100`},{default:e(({data:o,error:c})=>[n(f,null,{default:e(()=>[c!==void 0?(a(),l(D,{key:0,error:c},null,8,["error"])):(a(),l(B,{key:1,class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{label:"Name",key:"name"},{label:"Address",key:"socketAddress"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":1,"page-size":100,total:o==null?void 0:o.total,items:o==null?void 0:o.items,error:c,"empty-state-message":m("common.emptyState.message",{type:"Zone Egresses"}),"empty-state-cta-to":m("zone-egresses.href.docs"),"empty-state-cta-text":m("common.documentation"),"is-selected-row":s=>s.name===t.params.zoneEgress,onChange:t.update},{name:e(({row:s})=>[n(_,{to:{name:"zone-egress-summary-view",params:{zone:t.params.zone,zoneEgress:s.name},query:{page:1,size:100}}},{default:e(()=>[i(p(s.name),1)]),_:2},1032,["to"])]),socketAddress:e(({row:s})=>[s.zoneEgress.socketAddress.length>0?(a(),l(T,{key:0,text:s.zoneEgress.socketAddress},null,8,["text"])):(a(),h(S,{key:1},[i(p(m("common.collection.none")),1)],64))]),status:e(({row:s})=>[n(L,{status:s.state},null,8,["status"])]),details:e(({row:s})=>[n(_,{class:"details-link","data-testid":"details-link",to:{name:"zone-egress-detail-view",params:{zoneEgress:s.name}}},{default:e(()=>[i(p(m("common.collection.details_link"))+" ",1),n(g(A),{display:"inline-block",decorative:"",size:g(C)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"]))]),_:2},1024),i(),t.params.zoneEgress?(a(),l(k,{key:0},{default:e(s=>[n(N,{onClose:u=>t.replace({name:"zone-egress-list-view",params:{zone:t.params.zone},query:{page:1,size:100}})},{default:e(()=>[(a(),l(b(s.Component),{name:t.params.zoneEgress,"zone-egress-overview":o==null?void 0:o.items.find(u=>u.name===t.params.zoneEgress)},null,8,["name","zone-egress-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):z("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["params"])):z("",!0)]),_:1})}}});const P=R(Z,[["__scopeId","data-v-c16dd184"]]);export{P as default};
