import{d as V,a as l,o as t,b as i,w as a,E as b,W as g,t as r,f as m,e as o,F as d,c as _,V as x,p as k,K as D,D as B,q as h,_ as R}from"./index-RCz0LtdB.js";import{A}from"./AppCollection-BtqZNFPF.js";import{S as L}from"./StatusBadge-3fOP4roF.js";import{S as N}from"./SummaryView-0mnCJCQI.js";const P=V({__name:"ServiceListView",setup(E){return(I,K)=>{const u=l("RouterLink"),w=l("KCard"),C=l("RouterView"),S=l("AppView"),v=l("DataSource"),z=l("RouteView");return t(),i(v,{src:"/me"},{default:a(({data:f})=>[f?(t(),i(z,{key:0,name:"service-list-view",params:{page:1,size:f.pageSize,mesh:"",service:""}},{default:a(({route:s,t:c})=>[o(v,{src:`/meshes/${s.params.mesh}/service-insights/of/internal?page=${s.params.page}&size=${s.params.size}`},{default:a(({data:n,error:p})=>[o(S,null,{default:a(()=>[o(w,null,{default:a(()=>[p!==void 0?(t(),i(b,{key:0,error:p},null,8,["error"])):(t(),i(A,{key:1,class:"service-collection","data-testid":"service-collection","empty-state-message":c("common.emptyState.message",{type:"Services"}),headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"DP proxies (online / total)",key:"online"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":s.params.page,"page-size":s.params.size,total:n==null?void 0:n.total,items:n==null?void 0:n.items,error:p,"is-selected-row":e=>e.name===s.params.service,onChange:s.update},{name:a(({row:e})=>[o(g,{text:e.name},{default:a(()=>[o(u,{to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:s.params.page,size:s.params.size}}},{default:a(()=>[m(r(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),addressPort:a(({row:e})=>[e.addressPort?(t(),i(g,{key:0,text:e.addressPort},null,8,["text"])):(t(),_(d,{key:1},[m(r(c("common.collection.none")),1)],64))]),online:a(({row:e})=>[e.dataplanes?(t(),_(d,{key:0},[m(r(e.dataplanes.online||0)+" / "+r(e.dataplanes.total||0),1)],64)):(t(),_(d,{key:1},[m(r(c("common.collection.none")),1)],64))]),status:a(({row:e})=>[o(L,{status:e.status},null,8,["status"])]),details:a(({row:e})=>[o(u,{class:"details-link","data-testid":"details-link",to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:a(()=>[m(r(c("common.collection.details_link"))+" ",1),o(k(x),{decorative:"",size:k(D)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","headers","page-number","page-size","total","items","error","is-selected-row","onChange"]))]),_:2},1024),m(),s.params.service?(t(),i(C,{key:0},{default:a(e=>[o(N,{onClose:y=>s.replace({name:"service-list-view",params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size}})},{default:a(()=>[(t(),i(B(e.Component),{name:s.params.service,service:n==null?void 0:n.items.find(y=>y.name===s.params.service)},null,8,["name","service"]))]),_:2},1032,["onClose"])]),_:2},1024)):h("",!0)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):h("",!0)]),_:1})}}}),W=R(P,[["__scopeId","data-v-fb21dbd5"]]);export{W as default};
