import{d as w,r as v,o as m,a as _,w as e,h as t,q as k,b as s,g as u,G as h,t as d,e as y,F as V,V as C,D as E,v as x,H as A}from"./index-7e71fe76.js";import{A as L}from"./AppCollection-8d01782e.js";import{g as N,A as B,p as I,S as R,_ as S,n as $,f as T}from"./RouteView.vue_vue_type_script_setup_true_lang-159ad8a0.js";import{_ as q}from"./RouteTitle.vue_vue_type_script_setup_true_lang-3c1a3272.js";const D=w({__name:"ZoneEgressListView",props:{page:{type:Number,required:!0},size:{type:Number,required:!0}},setup(g){const n=g,{t:r}=N();function f(p){return p.map(l=>{const{name:i}=l,c={name:"zone-egress-detail-view",params:{zoneEgress:i}},o=$(l.zoneEgressInsight??{});return{detailViewRoute:c,name:i,status:o}})}return(p,l)=>{const i=v("RouterLink");return m(),_(S,{name:"zone-egress-list-view"},{default:e(({route:c})=>[t(B,null,{title:e(()=>[k("h2",null,[t(q,{title:s(r)("zone-egresses.routes.items.title"),render:!0},null,8,["title"])])]),default:e(()=>[u(),t(I,{src:`/zone-egresses?size=${n.size}&page=${n.page}`},{default:e(({data:o,error:b})=>[t(s(h),null,{body:e(()=>[t(L,{class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{label:"Name",key:"name"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":n.page,"page-size":n.size,total:o==null?void 0:o.total,items:o?f(o.items):void 0,error:b,onChange:c.update},{name:e(({row:a,rowValue:z})=>[t(i,{to:a.detailViewRoute,"data-testid":"detail-view-link"},{default:e(()=>[u(d(z),1)]),_:2},1032,["to"])]),status:e(({rowValue:a})=>[a?(m(),_(R,{key:0,status:a},null,8,["status"])):(m(),y(V,{key:1},[u(d(s(r)("common.collection.none")),1)],64))]),actions:e(({row:a})=>[t(s(C),{class:"actions-dropdown","data-testid":"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[t(s(E),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:e(()=>[t(s(x),{color:"var(--black-400)",icon:"more",size:"16"})]),_:1})]),items:e(()=>[t(s(A),{item:{to:a.detailViewRoute,label:s(r)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["page-number","page-size","total","items","error","onChange"])]),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});const j=T(D,[["__scopeId","data-v-0850253b"]]);export{j as default};
