import{d as z,h as i,o as s,a as c,w as a,j as o,T as C,C as v,X as y,k as d,t as m,b as r,F as _,S as b,z as k,M as V,e as D,A as S}from"./index-9gITI0JG.js";import{e as A}from"./kong-icons.es249-VUMzy5J0.js";const B=z({__name:"DelegatedGatewayListView",setup(L){return(N,P)=>{const g=i("RouterLink"),f=i("KCard"),w=i("AppView"),u=i("DataSource"),h=i("RouteView");return s(),c(u,{src:"/me"},{default:a(({data:x})=>[x?(s(),c(h,{key:0,name:"delegated-gateway-list-view",params:{page:1,size:10,mesh:""}},{default:a(({route:t,t:l})=>[o(u,{src:`/meshes/${t.params.mesh}/service-insights/of/gateway_delegated?page=${t.params.page}&size=${t.params.size}`},{default:a(({data:n,error:p})=>[o(w,null,{default:a(()=>[o(f,null,{default:a(()=>[p!==void 0?(s(),c(C,{key:0,error:p},null,8,["error"])):(s(),c(v,{key:1,class:"delegated-gateway-collection","data-testid":"delegated-gateway-collection","empty-state-message":l("common.emptyState.message",{type:"Delegated Gateways"}),"empty-state-cta-to":l("delegated-gateways.href.docs"),"empty-state-cta-text":l("common.documentation"),headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"DP proxies (online / total)",key:"dataplanes"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":t.params.page,"page-size":t.params.size,total:n==null?void 0:n.total,items:n==null?void 0:n.items,error:p,onChange:t.update},{name:a(({row:e})=>[o(y,{text:e.name},{default:a(()=>[o(g,{to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:t.params.page,size:t.params.size}}},{default:a(()=>[d(m(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),addressPort:a(({row:e})=>[e.addressPort?(s(),c(y,{key:0,text:e.addressPort},null,8,["text"])):(s(),r(_,{key:1},[d(m(l("common.collection.none")),1)],64))]),dataplanes:a(({row:e})=>[e.dataplanes?(s(),r(_,{key:0},[d(m(e.dataplanes.online||0)+" / "+m(e.dataplanes.total||0),1)],64)):(s(),r(_,{key:1},[d(m(l("common.collection.none")),1)],64))]),status:a(({row:e})=>[o(b,{status:e.status},null,8,["status"])]),details:a(({row:e})=>[o(g,{class:"details-link","data-testid":"details-link",to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:a(()=>[d(m(l("common.collection.details_link"))+" ",1),o(k(A),{decorative:"",size:k(V)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","page-number","page-size","total","items","error","onChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1})):D("",!0)]),_:1})}}}),$=S(B,[["__scopeId","data-v-640087f7"]]);export{$ as default};
