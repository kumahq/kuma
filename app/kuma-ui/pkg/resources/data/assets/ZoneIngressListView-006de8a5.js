import{d as w,r as v,o as n,a as l,w as e,h as t,q as y,b as s,g as p,G as h,t as d,e as I,F as V,$,D as C,v as S,H as x}from"./index-a928d02c.js";import{_ as A}from"./MultizoneInfo.vue_vue_type_script_setup_true_lang-46c6a1c8.js";import{A as L}from"./AppCollection-1deef537.js";import{g as N,e as B,A as R,_ as T,n as q,f as D}from"./RouteView.vue_vue_type_script_setup_true_lang-f622f9ae.js";import{_ as F}from"./DataSource.vue_vue_type_script_setup_true_lang-8eeb0b78.js";import{_ as Z}from"./RouteTitle.vue_vue_type_script_setup_true_lang-a99b1649.js";import{S as E}from"./StatusBadge-de159c5b.js";const G=w({__name:"ZoneIngressListView",props:{page:{type:Number,required:!0},size:{type:Number,required:!0}},setup(g){const i=g,{t:c}=N(),f=B();function b(_){return _.map(m=>{const{name:r}=m,u={name:"zone-ingress-detail-view",params:{zoneIngress:r}},o=q(m.zoneIngressInsight??{});return{detailViewRoute:u,name:r,status:o}})}return(_,m)=>{const r=v("RouterLink");return n(),l(T,{name:"zone-ingress-list-view"},{default:e(({route:u})=>[t(R,null,{title:e(()=>[y("h2",null,[t(Z,{title:s(c)("zone-ingresses.routes.items.title"),render:!0},null,8,["title"])])]),default:e(()=>[p(),s(f).getters["config/getMulticlusterStatus"]===!1?(n(),l(A,{key:0})):(n(),l(F,{key:1,src:`/zone-ingresses?size=${i.size}&page=${i.page}`},{default:e(({data:o,error:z})=>[t(s(h),null,{body:e(()=>[t(L,{class:"zone-ingress-collection","data-testid":"zone-ingress-collection",headers:[{label:"Name",key:"name"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":i.page,"page-size":i.size,total:o==null?void 0:o.total,items:o?b(o.items):void 0,error:z,onChange:u.update},{name:e(({row:a,rowValue:k})=>[t(r,{to:a.detailViewRoute,"data-testid":"detail-view-link"},{default:e(()=>[p(d(k),1)]),_:2},1032,["to"])]),status:e(({rowValue:a})=>[a?(n(),l(E,{key:0,status:a},null,8,["status"])):(n(),I(V,{key:1},[p(d(s(c)("common.collection.none")),1)],64))]),actions:e(({row:a})=>[t(s($),{class:"actions-dropdown","data-testid":"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[t(s(C),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:e(()=>[t(s(S),{color:"var(--black-400)",icon:"more",size:"16"})]),_:1})]),items:e(()=>[t(s(x),{item:{to:a.detailViewRoute,label:s(c)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["page-number","page-size","total","items","error","onChange"])]),_:2},1024)]),_:2},1032,["src"]))]),_:2},1024)]),_:1})}}});const Q=D(G,[["__scopeId","data-v-fee36c84"]]);export{Q as default};
