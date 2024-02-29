import{d as x,k as b,a as l,o as t,b as i,w as a,q as d,E as D,W as g,t as r,f as m,e as o,F as _,c as u,U as B,K as R,D as A,p as w,_ as L}from"./index-KOnKkPpw.js";import{A as N}from"./AppCollection-1OnRtgTt.js";import{S as P}from"./StatusBadge-w0GWtH_d.js";import{S as $}from"./SummaryView-SS0he8cU.js";const E=x({__name:"ServiceListView",setup(I){const h=b();return(K,q)=>{const v=l("RouterLink"),C=l("KCard"),S=l("RouterView"),z=l("AppView"),f=l("DataSource"),V=l("RouteView");return t(),i(f,{src:"/me"},{default:a(({data:y})=>[y?(t(),i(V,{key:0,name:"service-list-view",params:{page:1,size:y.pageSize,mesh:"",service:""}},{default:a(({route:s,t:c})=>[o(f,{src:`/meshes/${s.params.mesh}/service-insights/of/${d(h)("use gateways ui")?"internal":"all"}?page=${s.params.page}&size=${s.params.size}`},{default:a(({data:n,error:p})=>[o(z,null,{default:a(()=>[o(C,null,{default:a(()=>[p!==void 0?(t(),i(D,{key:0,error:p},null,8,["error"])):(t(),i(N,{key:1,class:"service-collection","data-testid":"service-collection","empty-state-message":c("common.emptyState.message",{type:"Services"}),headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"DP proxies (online / total)",key:"online"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":s.params.page,"page-size":s.params.size,total:n==null?void 0:n.total,items:n==null?void 0:n.items,error:p,"is-selected-row":e=>e.name===s.params.service,onChange:s.update},{name:a(({row:e})=>[o(g,{text:e.name},{default:a(()=>[o(v,{to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:s.params.page,size:s.params.size}}},{default:a(()=>[m(r(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),addressPort:a(({row:e})=>[e.addressPort?(t(),i(g,{key:0,text:e.addressPort},null,8,["text"])):(t(),u(_,{key:1},[m(r(c("common.collection.none")),1)],64))]),online:a(({row:e})=>[e.dataplanes?(t(),u(_,{key:0},[m(r(e.dataplanes.online||0)+" / "+r(e.dataplanes.total||0),1)],64)):(t(),u(_,{key:1},[m(r(c("common.collection.none")),1)],64))]),status:a(({row:e})=>[o(P,{status:e.status},null,8,["status"])]),details:a(({row:e})=>[o(v,{class:"details-link","data-testid":"details-link",to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:a(()=>[m(r(c("common.collection.details_link"))+" ",1),o(d(B),{display:"inline-block",decorative:"",size:d(R)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","headers","page-number","page-size","total","items","error","is-selected-row","onChange"]))]),_:2},1024),m(),s.params.service?(t(),i(S,{key:0},{default:a(e=>[o($,{onClose:k=>s.replace({name:"service-list-view",params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size}})},{default:a(()=>[(t(),i(A(e.Component),{name:s.params.service,service:n==null?void 0:n.items.find(k=>k.name===s.params.service)},null,8,["name","service"]))]),_:2},1032,["onClose"])]),_:2},1024)):w("",!0)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):w("",!0)]),_:1})}}}),O=L(E,[["__scopeId","data-v-ff2d5f60"]]);export{O as default};
