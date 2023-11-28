import{K as z}from"./index-52545d1d.js";import{d as b,a as n,o as i,b as l,w as s,e as m,p as y,f as p,t as r,q as f,V as x,G as R,s as g,_ as S}from"./index-079a3a85.js";import{A as B}from"./AppCollection-962d3121.js";import{E as I}from"./ErrorBlock-1fa583ae.js";import{S as L}from"./SummaryView-c87c8065.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-448d7af6.js";import"./TextWithCopyButton-f3080f30.js";import"./CopyButton-86a7f09c.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-b68734c9.js";const D={class:"stack"},N=b({__name:"MeshListView",setup(T){return(A,K)=>{const w=n("RouteTitle"),_=n("RouterLink"),k=n("KCard"),v=n("RouterView"),V=n("AppView"),d=n("DataSource"),C=n("RouteView");return i(),l(d,{src:"/me"},{default:s(({data:h})=>[h?(i(),l(C,{key:0,name:"mesh-list-view",params:{page:1,size:h.pageSize,mesh:""}},{default:s(({route:e,t:o})=>[m(d,{src:`/mesh-insights?page=${e.params.page}&size=${e.params.size}`},{default:s(({data:t,error:c})=>[m(V,null,{title:s(()=>[y("h1",null,[m(w,{title:o("meshes.routes.items.title")},null,8,["title"])])]),default:s(()=>[p(),y("div",D,[m(k,null,{body:s(()=>[c!==void 0?(i(),l(I,{key:0,error:c},null,8,["error"])):(i(),l(B,{key:1,class:"mesh-collection","data-testid":"mesh-collection",headers:[{label:o("meshes.common.name"),key:"name"},{label:o("meshes.routes.items.collection.services"),key:"services"},{label:o("meshes.routes.items.collection.dataplanes"),key:"dataplanes"},{label:"Details",key:"details",hideLabel:!0}],"page-number":parseInt(e.params.page),"page-size":parseInt(e.params.size),total:t==null?void 0:t.total,items:t==null?void 0:t.items,error:c,"empty-state-message":o("common.emptyState.message",{type:"Meshes"}),"empty-state-cta-to":o("meshes.href.docs"),"empty-state-cta-text":o("common.documentation"),"is-selected-row":a=>a.name===e.params.mesh,onChange:e.update},{name:s(({row:a})=>[m(_,{to:{name:"mesh-detail-view",params:{mesh:a.name},query:{page:e.params.page,size:e.params.size}}},{default:s(()=>[p(r(a.name),1)]),_:2},1032,["to"])]),services:s(({row:a})=>[p(r(a.services.internal??"0"),1)]),dataplanes:s(({row:a})=>[p(r(a.dataplanesByType.standard.online??"0")+" / "+r(a.dataplanesByType.standard.total??"0"),1)]),details:s(({row:a})=>[m(_,{class:"details-link","data-testid":"details-link",to:{name:"mesh-detail-view",params:{mesh:a.name}}},{default:s(()=>[p(r(o("common.collection.details_link"))+" ",1),m(f(x),{display:"inline-block",decorative:"",size:f(z)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["headers","page-number","page-size","total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"]))]),_:2},1024),p(),e.params.mesh?(i(),l(v,{key:0},{default:s(a=>[m(L,{onClose:u=>e.replace({name:"mesh-list-view",params:{mesh:e.params.mesh},query:{page:e.params.page,size:e.params.size}})},{default:s(()=>[(i(),l(R(a.Component),{name:e.params.mesh,"mesh-insight":t==null?void 0:t.items.find(u=>u.name===e.params.mesh)},null,8,["name","mesh-insight"]))]),_:2},1032,["onClose"])]),_:2},1024)):g("",!0)])]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):g("",!0)]),_:1})}}});const F=S(N,[["__scopeId","data-v-490caf3b"]]);export{F as default};
