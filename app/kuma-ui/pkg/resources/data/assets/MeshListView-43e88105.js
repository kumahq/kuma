import{K as z}from"./index-fce48c05.js";import{d as b,a as n,o as i,b as l,w as s,e as o,m as y,f as p,t as r,l as g,R,C as x,p as f,_ as S}from"./index-117d39c8.js";import{A as B}from"./AppCollection-49f203c7.js";import{E as L}from"./ErrorBlock-0f6b9bc9.js";import{S as D}from"./SummaryView-e1fca82a.js";import"./TextWithCopyButton-9fd2ca95.js";import"./CopyButton-3c068074.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-729e4f97.js";const N={class:"stack"},T=b({__name:"MeshListView",setup(A){return(I,K)=>{const w=n("RouteTitle"),_=n("RouterLink"),k=n("KCard"),v=n("RouterView"),C=n("AppView"),d=n("DataSource"),V=n("RouteView");return i(),l(d,{src:"/me"},{default:s(({data:h})=>[h?(i(),l(V,{key:0,name:"mesh-list-view",params:{page:1,size:h.pageSize,mesh:""}},{default:s(({route:e,t:m})=>[o(d,{src:`/mesh-insights?page=${e.params.page}&size=${e.params.size}`},{default:s(({data:t,error:c})=>[o(C,null,{title:s(()=>[y("h1",null,[o(w,{title:m("meshes.routes.items.title")},null,8,["title"])])]),default:s(()=>[p(),y("div",N,[o(k,null,{default:s(()=>[c!==void 0?(i(),l(L,{key:0,error:c},null,8,["error"])):(i(),l(B,{key:1,class:"mesh-collection","data-testid":"mesh-collection",headers:[{label:m("meshes.common.name"),key:"name"},{label:m("meshes.routes.items.collection.services"),key:"services"},{label:m("meshes.routes.items.collection.dataplanes"),key:"dataplanes"},{label:"Details",key:"details",hideLabel:!0}],"page-number":e.params.page,"page-size":e.params.size,total:t==null?void 0:t.total,items:t==null?void 0:t.items,error:c,"empty-state-message":m("common.emptyState.message",{type:"Meshes"}),"empty-state-cta-to":m("meshes.href.docs"),"empty-state-cta-text":m("common.documentation"),"is-selected-row":a=>a.name===e.params.mesh,onChange:e.update},{name:s(({row:a})=>[o(_,{to:{name:"mesh-detail-view",params:{mesh:a.name},query:{page:e.params.page,size:e.params.size}}},{default:s(()=>[p(r(a.name),1)]),_:2},1032,["to"])]),services:s(({row:a})=>[p(r(a.services.internal),1)]),dataplanes:s(({row:a})=>[p(r(a.dataplanesByType.standard.online)+" / "+r(a.dataplanesByType.standard.total),1)]),details:s(({row:a})=>[o(_,{class:"details-link","data-testid":"details-link",to:{name:"mesh-detail-view",params:{mesh:a.name}}},{default:s(()=>[p(r(m("common.collection.details_link"))+" ",1),o(g(R),{display:"inline-block",decorative:"",size:g(z)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["headers","page-number","page-size","total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"]))]),_:2},1024),p(),e.params.mesh?(i(),l(v,{key:0},{default:s(a=>[o(D,{onClose:u=>e.replace({name:"mesh-list-view",params:{mesh:e.params.mesh},query:{page:e.params.page,size:e.params.size}})},{default:s(()=>[(i(),l(x(a.Component),{name:e.params.mesh,"mesh-insight":t==null?void 0:t.items.find(u=>u.name===e.params.mesh)},null,8,["name","mesh-insight"]))]),_:2},1032,["onClose"])]),_:2},1024)):f("",!0)])]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):f("",!0)]),_:1})}}});const F=S(T,[["__scopeId","data-v-a8e593e7"]]);export{F as default};
