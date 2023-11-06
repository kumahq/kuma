import{d as z,r as m,o as i,i as l,w as s,j as o,p as y,n as p,E as b,H as r,k as g,a0 as x,K as R,q as S,m as w,t as B}from"./index-c6bd05ee.js";import{A as I}from"./AppCollection-6aabd095.js";import{S as L}from"./SummaryView-227f39d3.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-fb6e10fa.js";const A={class:"stack"},D=z({__name:"MeshListView",setup(N){return(T,K)=>{const k=m("RouteTitle"),_=m("RouterLink"),f=m("KCard"),v=m("RouterView"),C=m("AppView"),d=m("DataSource"),V=m("RouteView");return i(),l(d,{src:"/me"},{default:s(({data:h})=>[h?(i(),l(V,{key:0,name:"mesh-list-view",params:{page:1,size:h.pageSize,mesh:""}},{default:s(({route:e,t:n})=>[o(d,{src:`/meshes?page=${e.params.page}&size=${e.params.size}`},{default:s(({data:t,error:c})=>[o(C,null,{title:s(()=>[y("h1",null,[o(k,{title:n("meshes.routes.items.title"),render:!0},null,8,["title"])])]),default:s(()=>[p(),y("div",A,[o(f,null,{body:s(()=>[c!==void 0?(i(),l(b,{key:0,error:c},null,8,["error"])):(i(),l(I,{key:1,class:"mesh-collection","data-testid":"mesh-collection",headers:[{label:n("meshes.common.name"),key:"name"},{label:n("meshes.routes.items.collection.services"),key:"services"},{label:n("meshes.routes.items.collection.dataplanes"),key:"dataplanes"},{label:"Details",key:"details",hideLabel:!0}],"page-number":parseInt(e.params.page),"page-size":parseInt(e.params.size),total:t==null?void 0:t.total,items:t==null?void 0:t.items,error:c,"empty-state-message":n("common.emptyState.message",{type:"Meshes"}),"empty-state-cta-to":n("meshes.href.docs"),"empty-state-cta-text":n("common.documentation"),"is-selected-row":a=>a.name===e.params.mesh,onChange:e.update},{name:s(({row:a})=>[o(_,{to:{name:"mesh-detail-view",params:{mesh:a.name},query:{page:e.params.page,size:e.params.size}}},{default:s(()=>[p(r(a.name),1)]),_:2},1032,["to"])]),services:s(({row:a})=>[p(r(a.services.internal??"0"),1)]),dataplanes:s(({row:a})=>[p(r(a.dataplanesByType.standard.online??"0")+" / "+r(a.dataplanesByType.standard.total??"0"),1)]),details:s(({row:a})=>[o(_,{class:"details-link","data-testid":"details-link",to:{name:"mesh-detail-view",params:{mesh:a.name}}},{default:s(()=>[p(r(n("common.collection.details_link"))+" ",1),o(g(x),{display:"inline-block",decorative:"",size:g(R)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["headers","page-number","page-size","total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"]))]),_:2},1024),p(),e.params.mesh?(i(),l(v,{key:0},{default:s(a=>[o(L,{onClose:u=>e.replace({name:"mesh-list-view",params:{mesh:e.params.mesh},query:{page:e.params.page,size:e.params.size}})},{default:s(()=>[(i(),l(S(a.Component),{name:e.params.mesh,"mesh-insight":t==null?void 0:t.items.find(u=>u.name===e.params.mesh)},null,8,["name","mesh-insight"]))]),_:2},1032,["onClose"])]),_:2},1024)):w("",!0)])]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):w("",!0)]),_:1})}}});const j=B(D,[["__scopeId","data-v-924b1309"]]);export{j as default};
