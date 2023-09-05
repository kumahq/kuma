import{g as h,i as f,A as g,E as y,q as w,r as v,K as k,_ as b,f as x}from"./RouteView.vue_vue_type_script_setup_true_lang-da83f5a8.js";import{d as z,r as C,o as l,a as c,w as s,h as t,i as r,b as e,g as p,k as L,t as R,R as E,H as N,x as V,J as A}from"./index-9a3d231d.js";import{_ as I}from"./RouteTitle.vue_vue_type_script_setup_true_lang-3a51c48f.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-fe937ad6.js";const T={class:"stack"},$=z({__name:"MeshListView",props:{page:{},size:{}},setup(_){const n=_,{t:o}=h();return(B,M)=>{const u=C("RouterLink");return l(),c(b,{name:"mesh-list-view"},{default:s(({route:d})=>[t(f,{src:`/meshes?page=${n.page}&size=${n.size}`},{default:s(({data:a,error:m})=>[t(g,null,{title:s(()=>[r("h1",null,[t(I,{title:e(o)("meshes.routes.items.title"),render:!0},null,8,["title"])])]),default:s(()=>[p(),r("div",T,[t(e(L),null,{body:s(()=>[m!==void 0?(l(),c(y,{key:0,error:m},null,8,["error"])):(l(),c(w,{key:1,class:"mesh-collection","data-testid":"mesh-collection",headers:[{label:"Name",key:"name"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":n.page,"page-size":n.size,total:a==null?void 0:a.total,items:a==null?void 0:a.items,error:m,"empty-state-message":e(o)("common.emptyState.message",{type:"Meshes"}),"empty-state-cta-to":e(o)("meshes.href.docs"),"empty-state-cta-text":e(o)("common.documentation"),onChange:d.update},{name:s(({row:i})=>[t(u,{to:{name:"mesh-detail-view",params:{mesh:i.name}}},{default:s(()=>[p(R(i.name),1)]),_:2},1032,["to"])]),actions:s(({row:i})=>[t(e(E),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:s(()=>[t(e(N),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:s(()=>[t(e(V),{color:e(v),icon:"more",size:e(k)},null,8,["color","size"])]),_:1})]),items:s(()=>[t(e(A),{item:{to:{name:"mesh-detail-view",params:{mesh:i.name}},label:e(o)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["page-number","page-size","total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"]))]),_:2},1024)])]),_:2},1024)]),_:2},1032,["src"])]),_:1})}}});const D=x($,[["__scopeId","data-v-403a1a96"]]);export{D as default};
