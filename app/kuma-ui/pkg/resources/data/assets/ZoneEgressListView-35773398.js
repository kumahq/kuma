import{K as C}from"./index-fce48c05.js";import{d as V,a as r,o as a,b as l,w as e,e as o,m as x,f as i,t as p,c as h,F as S,l as g,Z as A,B as b,p as z,_ as B}from"./index-fb2eded6.js";import{A as R}from"./AppCollection-53886c36.js";import{E as Z}from"./ErrorBlock-9db074cb.js";import{S as L}from"./StatusBadge-45121d03.js";import{S as N}from"./SummaryView-0fe22010.js";import{T as D}from"./TextWithCopyButton-e5ad58f3.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-66297082.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-2b371e68.js";import"./CopyButton-cc9318d8.js";const T=V({__name:"ZoneEgressListView",setup(I){return(K,$)=>{const y=r("RouteTitle"),_=r("RouterLink"),f=r("KCard"),k=r("RouterView"),u=r("DataSource"),w=r("AppView"),v=r("RouteView");return a(),l(u,{src:"/me"},{default:e(({data:E})=>[E?(a(),l(v,{key:0,name:"zone-egress-list-view",params:{zone:"",zoneEgress:""}},{default:e(({route:n,t:m})=>[o(w,null,{title:e(()=>[x("h2",null,[o(y,{title:m("zone-egresses.routes.items.title")},null,8,["title"])])]),default:e(()=>[i(),o(u,{src:`/zone-cps/${n.params.zone||"*"}/egresses?page=1&size=100`},{default:e(({data:t,error:c})=>[o(f,null,{default:e(()=>[c!==void 0?(a(),l(Z,{key:0,error:c},null,8,["error"])):(a(),l(R,{key:1,class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{label:"Name",key:"name"},{label:"Address",key:"socketAddress"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":1,"page-size":100,total:t==null?void 0:t.total,items:t==null?void 0:t.items,error:c,"empty-state-message":m("common.emptyState.message",{type:"Zone Egresses"}),"empty-state-cta-to":m("zone-egresses.href.docs"),"empty-state-cta-text":m("common.documentation"),"is-selected-row":s=>s.name===n.params.zoneEgress,onChange:n.update},{name:e(({row:s})=>[o(_,{to:{name:"zone-egress-summary-view",params:{zone:n.params.zone,zoneEgress:s.name},query:{page:1,size:100}}},{default:e(()=>[i(p(s.name),1)]),_:2},1032,["to"])]),socketAddress:e(({row:s})=>[s.zoneEgress.socketAddress.length>0?(a(),l(D,{key:0,text:s.zoneEgress.socketAddress},null,8,["text"])):(a(),h(S,{key:1},[i(p(m("common.collection.none")),1)],64))]),status:e(({row:s})=>[o(L,{status:s.state},null,8,["status"])]),details:e(({row:s})=>[o(_,{class:"details-link","data-testid":"details-link",to:{name:"zone-egress-detail-view",params:{zoneEgress:s.name}}},{default:e(()=>[i(p(m("common.collection.details_link"))+" ",1),o(g(A),{display:"inline-block",decorative:"",size:g(C)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"]))]),_:2},1024),i(),n.params.zoneEgress?(a(),l(k,{key:0},{default:e(s=>[o(N,{onClose:d=>n.replace({name:"zone-egress-list-view",params:{zone:n.params.zone},query:{page:1,size:100}})},{default:e(()=>[(a(),l(b(s.Component),{"zone-egress-overview":t==null?void 0:t.items.find(d=>d.name===n.params.zoneEgress)},null,8,["zone-egress-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):z("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["params"])):z("",!0)]),_:1})}}});const P=B(T,[["__scopeId","data-v-ecf2c484"]]);export{P as default};
