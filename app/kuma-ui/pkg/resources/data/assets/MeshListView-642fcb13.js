import{d as u,r as d,o as f,a as h,w as e,h as s,q as m,b as t,g as l,G as g,t as v,V as w,D as y,v as b,H as k}from"./index-7e71fe76.js";import{A as V}from"./AppCollection-8d01782e.js";import{g as z,p as C,A as x,_ as A,f as L}from"./RouteView.vue_vue_type_script_setup_true_lang-159ad8a0.js";import{_ as $}from"./RouteTitle.vue_vue_type_script_setup_true_lang-3c1a3272.js";const M={class:"stack"},N=u({__name:"MeshListView",props:{page:{},size:{}},setup(c){const o=c,{t:n}=z();return(B,S)=>{const r=d("RouterLink");return f(),h(A,{name:"mesh-list-view"},{default:e(({route:p})=>[s(C,{src:`/meshes?page=${o.page}&size=${o.size}`},{default:e(({data:a,error:_})=>[s(x,null,{title:e(()=>[m("h1",null,[s($,{title:t(n)("meshes.routes.items.title"),render:!0},null,8,["title"])])]),default:e(()=>[l(),m("div",M,[s(t(g),null,{body:e(()=>[s(V,{class:"mesh-collection","data-testid":"mesh-collection","empty-state-title":t(n)("common.emptyState.title"),"empty-state-message":t(n)("common.emptyState.message",{type:"Meshes"}),headers:[{label:"Name",key:"name"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":o.page,"page-size":o.size,total:a==null?void 0:a.total,items:a==null?void 0:a.items,error:_,onChange:p.update},{name:e(({row:i})=>[s(r,{to:{name:"mesh-detail-view",params:{mesh:i.name}}},{default:e(()=>[l(v(i.name),1)]),_:2},1032,["to"])]),actions:e(({row:i})=>[s(t(w),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[s(t(y),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:e(()=>[s(t(b),{color:"var(--black-400)",icon:"more",size:"16"})]),_:1})]),items:e(()=>[s(t(k),{item:{to:{name:"mesh-detail-view",params:{mesh:i.name}},label:t(n)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["empty-state-title","empty-state-message","page-number","page-size","total","items","error","onChange"])]),_:2},1024)])]),_:2},1024)]),_:2},1032,["src"])]),_:1})}}});const q=L(N,[["__scopeId","data-v-daff90a5"]]);export{q as default};
