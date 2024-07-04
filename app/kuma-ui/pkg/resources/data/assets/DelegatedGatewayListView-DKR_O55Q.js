import{d as b,r as n,o as s,m as p,w as a,b as o,P as v,A as C,U as u,e as i,t as m,c as d,F as g,S as z,p as V}from"./index-BRBWbknO.js";const G=b({__name:"DelegatedGatewayListView",setup(S){return(D,P)=>{const y=n("RouterLink"),w=n("XAction"),k=n("XActionGroup"),f=n("KCard"),h=n("AppView"),_=n("DataSource"),x=n("RouteView");return s(),p(_,{src:"/me"},{default:a(({data:A})=>[A?(s(),p(x,{key:0,name:"delegated-gateway-list-view",params:{page:1,size:10,mesh:""}},{default:a(({route:t,t:l})=>[o(_,{src:`/meshes/${t.params.mesh}/service-insights/of/gateway_delegated?page=${t.params.page}&size=${t.params.size}`},{default:a(({data:c,error:r})=>[o(h,null,{default:a(()=>[o(f,null,{default:a(()=>[r!==void 0?(s(),p(v,{key:0,error:r},null,8,["error"])):(s(),p(C,{key:1,class:"delegated-gateway-collection","data-testid":"delegated-gateway-collection","empty-state-message":l("common.emptyState.message",{type:"Delegated Gateways"}),"empty-state-cta-to":l("delegated-gateways.href.docs"),"empty-state-cta-text":l("common.documentation"),headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"DP proxies (online / total)",key:"dataplanes"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":t.params.page,"page-size":t.params.size,total:c==null?void 0:c.total,items:c==null?void 0:c.items,error:r,onChange:t.update},{name:a(({row:e})=>[o(u,{text:e.name},{default:a(()=>[o(y,{to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:t.params.page,size:t.params.size}}},{default:a(()=>[i(m(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),addressPort:a(({row:e})=>[e.addressPort?(s(),p(u,{key:0,text:e.addressPort},null,8,["text"])):(s(),d(g,{key:1},[i(m(l("common.collection.none")),1)],64))]),dataplanes:a(({row:e})=>[e.dataplanes?(s(),d(g,{key:0},[i(m(e.dataplanes.online||0)+" / "+m(e.dataplanes.total||0),1)],64)):(s(),d(g,{key:1},[i(m(l("common.collection.none")),1)],64))]),status:a(({row:e})=>[o(z,{status:e.status},null,8,["status"])]),actions:a(({row:e})=>[o(k,null,{default:a(()=>[o(w,{to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:a(()=>[i(m(l("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","page-number","page-size","total","items","error","onChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1})):V("",!0)]),_:1})}}});export{G as default};
