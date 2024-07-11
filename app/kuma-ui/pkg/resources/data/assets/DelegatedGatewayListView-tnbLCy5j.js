import{d as A,r as o,o as n,m as r,w as a,b as s,P as b,A as v,U as u,e as i,t as p,c as g,F as _,S as C}from"./index-Be4yFAuI.js";const R=A({__name:"DelegatedGatewayListView",setup(P){return(S,V)=>{const y=o("RouterLink"),h=o("XAction"),w=o("XActionGroup"),k=o("KCard"),f=o("AppView"),x=o("DataSource"),z=o("RouteView");return n(),r(z,{name:"delegated-gateway-list-view",params:{page:1,size:50,mesh:""}},{default:a(({route:t,t:l,me:c})=>[s(x,{src:`/meshes/${t.params.mesh}/service-insights/of/gateway_delegated?page=${t.params.page}&size=${t.params.size}`},{default:a(({data:m,error:d})=>[s(f,null,{default:a(()=>[s(k,null,{default:a(()=>[d!==void 0?(n(),r(b,{key:0,error:d},null,8,["error"])):(n(),r(v,{key:1,class:"delegated-gateway-collection","data-testid":"delegated-gateway-collection","empty-state-message":l("common.emptyState.message",{type:"Delegated Gateways"}),"empty-state-cta-to":l("delegated-gateways.href.docs"),"empty-state-cta-text":l("common.documentation"),headers:[{...c.get("headers.name"),label:"Name",key:"name"},{...c.get("headers.addressPort"),label:"Address",key:"addressPort"},{...c.get("headers.dataplanes"),label:"DP proxies (online / total)",key:"dataplanes"},{...c.get("headers.status"),label:"Status",key:"status"},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],"page-number":t.params.page,"page-size":t.params.size,total:m==null?void 0:m.total,items:m==null?void 0:m.items,error:d,onChange:t.update,onResize:c.set},{name:a(({row:e})=>[s(u,{text:e.name},{default:a(()=>[s(y,{to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:t.params.page,size:t.params.size}}},{default:a(()=>[i(p(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),addressPort:a(({row:e})=>[e.addressPort?(n(),r(u,{key:0,text:e.addressPort},null,8,["text"])):(n(),g(_,{key:1},[i(p(l("common.collection.none")),1)],64))]),dataplanes:a(({row:e})=>[e.dataplanes?(n(),g(_,{key:0},[i(p(e.dataplanes.online||0)+" / "+p(e.dataplanes.total||0),1)],64)):(n(),g(_,{key:1},[i(p(l("common.collection.none")),1)],64))]),status:a(({row:e})=>[s(C,{status:e.status},null,8,["status"])]),actions:a(({row:e})=>[s(w,null,{default:a(()=>[s(h,{to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:a(()=>[i(p(l("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error","onChange","onResize"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1})}}});export{R as default};