import{d as z,a as n,o as r,b as d,w as e,e as o,m as _,f as i,t as m,U as C,q as h,K as V,p as b,_ as x}from"./index-KOnKkPpw.js";import{A as L}from"./AppCollection-1OnRtgTt.js";const D={class:"stack"},R=z({__name:"MeshListView",setup(S){return(B,N)=>{const u=n("RouteTitle"),c=n("RouterLink"),y=n("DataLoader"),g=n("KCard"),f=n("AppView"),w=n("RouteView"),k=n("DataSource");return r(),d(k,{src:"/me"},{default:e(({data:p})=>[p?(r(),d(w,{key:0,name:"mesh-list-view",params:{page:1,size:p.pageSize,mesh:""}},{default:e(({route:a,t})=>[o(f,null,{title:e(()=>[_("h1",null,[o(u,{title:t("meshes.routes.items.title")},null,8,["title"])])]),default:e(()=>[i(),_("div",D,[o(g,null,{default:e(()=>[o(y,{src:`/mesh-insights?page=${a.params.page}&size=${a.params.size}`,loader:!1},{default:e(({data:l,error:v})=>[o(L,{class:"mesh-collection","data-testid":"mesh-collection",headers:[{label:t("meshes.common.name"),key:"name"},{label:t("meshes.routes.items.collection.services"),key:"services"},{label:t("meshes.routes.items.collection.dataplanes"),key:"dataplanes"},{label:"Details",key:"details",hideLabel:!0}],"page-number":a.params.page,"page-size":a.params.size,total:l==null?void 0:l.total,items:l==null?void 0:l.items,error:v,"empty-state-message":t("common.emptyState.message",{type:"Meshes"}),"empty-state-cta-to":t("meshes.href.docs"),"empty-state-cta-text":t("common.documentation"),"is-selected-row":s=>s.name===a.params.mesh,onChange:a.update},{name:e(({row:s})=>[o(c,{to:{name:"mesh-detail-view",params:{mesh:s.name},query:{page:a.params.page,size:a.params.size}}},{default:e(()=>[i(m(s.name),1)]),_:2},1032,["to"])]),services:e(({row:s})=>[i(m(s.services.internal),1)]),dataplanes:e(({row:s})=>[i(m(s.dataplanesByType.standard.online)+" / "+m(s.dataplanesByType.standard.total),1)]),details:e(({row:s})=>[o(c,{class:"details-link","data-testid":"details-link",to:{name:"mesh-detail-view",params:{mesh:s.name}}},{default:e(()=>[i(m(t("common.collection.details_link"))+" ",1),o(h(C),{display:"inline-block",decorative:"",size:h(V)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["headers","page-number","page-size","total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"])]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024)]),_:2},1032,["params"])):b("",!0)]),_:1})}}}),I=x(R,[["__scopeId","data-v-3ecdead6"]]);export{I as default};
