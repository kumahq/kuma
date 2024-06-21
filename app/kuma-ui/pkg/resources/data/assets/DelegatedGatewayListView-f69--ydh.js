import{d as v,i,o as s,a as p,w as a,j as o,a1 as z,a3 as y,k as c,t as m,b as r,H as _,A as k,K as C,e as b,_ as V}from"./index-CyAtMQ3G.js";import{p as D}from"./kong-icons.es249-vgKX97Et.js";import{A as S}from"./AppCollection-DSdmYcz_.js";import{S as A}from"./StatusBadge-yau4rj5h.js";import"./kong-icons.es245-BjB891cP.js";const B=v({__name:"DelegatedGatewayListView",setup(L){return(N,P)=>{const g=i("RouterLink"),f=i("KCard"),w=i("AppView"),u=i("DataSource"),h=i("RouteView");return s(),p(u,{src:"/me"},{default:a(({data:x})=>[x?(s(),p(h,{key:0,name:"delegated-gateway-list-view",params:{page:1,size:10,mesh:""}},{default:a(({route:t,t:l})=>[o(u,{src:`/meshes/${t.params.mesh}/service-insights/of/gateway_delegated?page=${t.params.page}&size=${t.params.size}`},{default:a(({data:n,error:d})=>[o(w,null,{default:a(()=>[o(f,null,{default:a(()=>[d!==void 0?(s(),p(z,{key:0,error:d},null,8,["error"])):(s(),p(S,{key:1,class:"delegated-gateway-collection","data-testid":"delegated-gateway-collection","empty-state-message":l("common.emptyState.message",{type:"Delegated Gateways"}),"empty-state-cta-to":l("delegated-gateways.href.docs"),"empty-state-cta-text":l("common.documentation"),headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"DP proxies (online / total)",key:"dataplanes"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":t.params.page,"page-size":t.params.size,total:n==null?void 0:n.total,items:n==null?void 0:n.items,error:d,onChange:t.update},{name:a(({row:e})=>[o(y,{text:e.name},{default:a(()=>[o(g,{to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:t.params.page,size:t.params.size}}},{default:a(()=>[c(m(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),addressPort:a(({row:e})=>[e.addressPort?(s(),p(y,{key:0,text:e.addressPort},null,8,["text"])):(s(),r(_,{key:1},[c(m(l("common.collection.none")),1)],64))]),dataplanes:a(({row:e})=>[e.dataplanes?(s(),r(_,{key:0},[c(m(e.dataplanes.online||0)+" / "+m(e.dataplanes.total||0),1)],64)):(s(),r(_,{key:1},[c(m(l("common.collection.none")),1)],64))]),status:a(({row:e})=>[o(A,{status:e.status},null,8,["status"])]),details:a(({row:e})=>[o(g,{class:"details-link","data-testid":"details-link",to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:a(()=>[c(m(l("common.collection.details_link"))+" ",1),o(k(D),{decorative:"",size:k(C)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","page-number","page-size","total","items","error","onChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1})):b("",!0)]),_:1})}}}),G=V(B,[["__scopeId","data-v-640087f7"]]);export{G as default};
