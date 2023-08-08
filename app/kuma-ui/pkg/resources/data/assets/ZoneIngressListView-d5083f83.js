import{d as w,r as v,o as n,a as l,w as e,h as t,q as y,b as s,g as p,G as h,t as d,e as V,F as I,V as C,D as S,v as x,H as A}from"./index-7e71fe76.js";import{_ as L}from"./MultizoneInfo.vue_vue_type_script_setup_true_lang-e1bc3284.js";import{A as N}from"./AppCollection-8d01782e.js";import{g as $,e as B,A as R,p as T,S as q,_ as D,n as F,f as Z}from"./RouteView.vue_vue_type_script_setup_true_lang-159ad8a0.js";import{_ as E}from"./RouteTitle.vue_vue_type_script_setup_true_lang-3c1a3272.js";const G=w({__name:"ZoneIngressListView",props:{page:{type:Number,required:!0},size:{type:Number,required:!0}},setup(g){const i=g,{t:c}=$(),f=B();function b(_){return _.map(u=>{const{name:r}=u,m={name:"zone-ingress-detail-view",params:{zoneIngress:r}},o=F(u.zoneIngressInsight??{});return{detailViewRoute:m,name:r,status:o}})}return(_,u)=>{const r=v("RouterLink");return n(),l(D,{name:"zone-ingress-list-view"},{default:e(({route:m})=>[t(R,null,{title:e(()=>[y("h2",null,[t(E,{title:s(c)("zone-ingresses.routes.items.title"),render:!0},null,8,["title"])])]),default:e(()=>[p(),s(f).getters["config/getMulticlusterStatus"]===!1?(n(),l(L,{key:0})):(n(),l(T,{key:1,src:`/zone-ingresses?size=${i.size}&page=${i.page}`},{default:e(({data:o,error:z})=>[t(s(h),null,{body:e(()=>[t(N,{class:"zone-ingress-collection","data-testid":"zone-ingress-collection",headers:[{label:"Name",key:"name"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":i.page,"page-size":i.size,total:o==null?void 0:o.total,items:o?b(o.items):void 0,error:z,onChange:m.update},{name:e(({row:a,rowValue:k})=>[t(r,{to:a.detailViewRoute,"data-testid":"detail-view-link"},{default:e(()=>[p(d(k),1)]),_:2},1032,["to"])]),status:e(({rowValue:a})=>[a?(n(),l(q,{key:0,status:a},null,8,["status"])):(n(),V(I,{key:1},[p(d(s(c)("common.collection.none")),1)],64))]),actions:e(({row:a})=>[t(s(C),{class:"actions-dropdown","data-testid":"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[t(s(S),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:e(()=>[t(s(x),{color:"var(--black-400)",icon:"more",size:"16"})]),_:1})]),items:e(()=>[t(s(A),{item:{to:a.detailViewRoute,label:s(c)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["page-number","page-size","total","items","error","onChange"])]),_:2},1024)]),_:2},1032,["src"]))]),_:2},1024)]),_:1})}}});const O=Z(G,[["__scopeId","data-v-fee36c84"]]);export{O as default};
