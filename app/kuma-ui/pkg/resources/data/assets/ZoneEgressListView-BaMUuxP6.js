import{d as E,a as r,o as a,b as l,w as e,e as t,m as V,f as i,a1 as h,t as p,T as v,c as x,F as S,p as g,Q as b,K as A,B,q as u,_ as R}from"./index-IotYe1KN.js";import{A as N}from"./AppCollection-atohepkv.js";import{S as D}from"./StatusBadge-CbARFMK9.js";import{S as L}from"./SummaryView-ChaOsfK2.js";const T=E({__name:"ZoneEgressListView",setup(I){return(K,Z)=>{const z=r("RouteTitle"),d=r("RouterLink"),y=r("KCard"),k=r("RouterView"),_=r("DataSource"),f=r("AppView"),w=r("RouteView");return a(),l(_,{src:"/me"},{default:e(({data:C})=>[C?(a(),l(w,{key:0,name:"zone-egress-list-view",params:{zone:"",zoneEgress:""}},{default:e(({route:n,t:m})=>[t(f,null,{title:e(()=>[V("h2",null,[t(z,{title:m("zone-egresses.routes.items.title")},null,8,["title"])])]),default:e(()=>[i(),t(_,{src:`/zone-cps/${n.params.zone||"*"}/egresses?page=1&size=100`},{default:e(({data:o,error:c})=>[t(y,null,{default:e(()=>[c!==void 0?(a(),l(h,{key:0,error:c},null,8,["error"])):(a(),l(N,{key:1,class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{label:"Name",key:"name"},{label:"Address",key:"socketAddress"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":1,"page-size":100,total:o==null?void 0:o.total,items:o==null?void 0:o.items,error:c,"empty-state-message":m("common.emptyState.message",{type:"Zone Egresses"}),"empty-state-cta-to":m("zone-egresses.href.docs"),"empty-state-cta-text":m("common.documentation"),"is-selected-row":s=>s.name===n.params.zoneEgress,onChange:n.update},{name:e(({row:s})=>[t(d,{to:{name:"zone-egress-summary-view",params:{zone:n.params.zone,zoneEgress:s.id},query:{page:1,size:100}}},{default:e(()=>[i(p(s.name),1)]),_:2},1032,["to"])]),socketAddress:e(({row:s})=>[s.zoneEgress.socketAddress.length>0?(a(),l(v,{key:0,text:s.zoneEgress.socketAddress},null,8,["text"])):(a(),x(S,{key:1},[i(p(m("common.collection.none")),1)],64))]),status:e(({row:s})=>[t(D,{status:s.state},null,8,["status"])]),details:e(({row:s})=>[t(d,{class:"details-link","data-testid":"details-link",to:{name:"zone-egress-detail-view",params:{zoneEgress:s.id}}},{default:e(()=>[i(p(m("common.collection.details_link"))+" ",1),t(g(b),{decorative:"",size:g(A)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"]))]),_:2},1024),i(),n.params.zoneEgress?(a(),l(k,{key:0},{default:e(s=>[t(L,{onClose:$=>n.replace({name:"zone-egress-list-view",params:{zone:n.params.zone},query:{page:1,size:100}})},{default:e(()=>[typeof o<"u"?(a(),l(B(s.Component),{key:0,items:o.items},null,8,["items"])):u("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):u("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["params"])):u("",!0)]),_:1})}}}),U=R(T,[["__scopeId","data-v-5e29b2b4"]]);export{U as default};
