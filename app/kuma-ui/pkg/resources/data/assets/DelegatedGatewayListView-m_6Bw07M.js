import{d as C,a as i,o as s,b as d,w as a,E as b,W as u,t as m,f as c,e as o,F as r,c as _,U as v,q as k,K as z,p as V,_ as D}from"./index-pAyRVwwQ.js";import{A as S}from"./AppCollection-WrwfvnZG.js";import{S as B}from"./StatusBadge-8xZKochX.js";const A=C({__name:"DelegatedGatewayListView",setup(L){return(N,P)=>{const g=i("RouterLink"),f=i("KCard"),w=i("AppView"),y=i("DataSource"),h=i("RouteView");return s(),d(y,{src:"/me"},{default:a(({data:x})=>[x?(s(),d(h,{key:0,name:"delegated-gateway-list-view",params:{page:1,size:10,mesh:""}},{default:a(({route:t,t:l})=>[o(y,{src:`/meshes/${t.params.mesh}/service-insights/of/gateway_delegated?page=${t.params.page}&size=${t.params.size}`},{default:a(({data:n,error:p})=>[o(w,null,{default:a(()=>[o(f,null,{default:a(()=>[p!==void 0?(s(),d(b,{key:0,error:p},null,8,["error"])):(s(),d(S,{key:1,class:"delegated-gateway-collection","data-testid":"delegated-gateway-collection","empty-state-message":l("common.emptyState.message",{type:"Delegated Gateways"}),"empty-state-cta-to":l("delegated-gateways.href.docs"),"empty-state-cta-text":l("common.documentation"),headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"DP proxies (online / total)",key:"dataplanes"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":t.params.page,"page-size":t.params.size,total:n==null?void 0:n.total,items:n==null?void 0:n.items,error:p,onChange:t.update},{name:a(({row:e})=>[o(u,{text:e.name},{default:a(()=>[o(g,{to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:t.params.page,size:t.params.size}}},{default:a(()=>[c(m(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),addressPort:a(({row:e})=>[e.addressPort?(s(),d(u,{key:0,text:e.addressPort},null,8,["text"])):(s(),_(r,{key:1},[c(m(l("common.collection.none")),1)],64))]),dataplanes:a(({row:e})=>[e.dataplanes?(s(),_(r,{key:0},[c(m(e.dataplanes.online||0)+" / "+m(e.dataplanes.total||0),1)],64)):(s(),_(r,{key:1},[c(m(l("common.collection.none")),1)],64))]),status:a(({row:e})=>[o(B,{status:e.status},null,8,["status"])]),details:a(({row:e})=>[o(g,{class:"details-link","data-testid":"details-link",to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:a(()=>[c(m(l("common.collection.details_link"))+" ",1),o(k(v),{display:"inline-block",decorative:"",size:k(z)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error","onChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1})):V("",!0)]),_:1})}}}),R=D(A,[["__scopeId","data-v-881adeca"]]);export{R as default};
