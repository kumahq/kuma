import{g as b,A as w,o as v,p as h,S as k,q as E,K as V,_ as C,f as I}from"./RouteView.vue_vue_type_script_setup_true_lang-fbd72064.js";import{d as N,r as R,o as m,a as _,w as t,h as s,q as x,b as e,g as p,G as L,t as d,e as S,F as T,V as A,D as B,v as Z,H as $}from"./index-73ac0e73.js";import{_ as q}from"./RouteTitle.vue_vue_type_script_setup_true_lang-0d00e209.js";import{g as O}from"./dataplane-30467516.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-367da69d.js";const D=N({__name:"ZoneEgressListView",props:{page:{type:Number,required:!0},size:{type:Number,required:!0}},setup(g){const i=g,{t:a}=b();function f(u){return u.map(c=>{const{name:r}=c,l={name:"zone-egress-detail-view",params:{zoneEgress:r}},o=O(c.zoneEgressInsight??{});return{detailViewRoute:l,name:r,status:o}})}return(u,c)=>{const r=R("RouterLink");return m(),_(C,{name:"zone-egress-list-view"},{default:t(({route:l})=>[s(w,null,{title:t(()=>[x("h1",null,[s(q,{title:e(a)("zone-egresses.routes.items.title"),render:!0},null,8,["title"])])]),default:t(()=>[p(),s(v,{src:`/zone-egresses?size=${i.size}&page=${i.page}`},{default:t(({data:o,error:y})=>[s(e(L),null,{body:t(()=>[s(h,{class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{label:"Name",key:"name"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":i.page,"page-size":i.size,total:o==null?void 0:o.total,items:o?f(o.items):void 0,error:y,"empty-state-message":e(a)("common.emptyState.message",{type:"Zone Egresses"}),"empty-state-cta-to":e(a)("zone-egresses.href.docs"),"empty-state-cta-text":e(a)("common.documentation"),onChange:l.update},{name:t(({row:n,rowValue:z})=>[s(r,{to:n.detailViewRoute,"data-testid":"detail-view-link"},{default:t(()=>[p(d(z),1)]),_:2},1032,["to"])]),status:t(({rowValue:n})=>[n?(m(),_(k,{key:0,status:n},null,8,["status"])):(m(),S(T,{key:1},[p(d(e(a)("common.collection.none")),1)],64))]),actions:t(({row:n})=>[s(e(A),{class:"actions-dropdown","data-testid":"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:t(()=>[s(e(B),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:t(()=>[s(e(Z),{color:e(E),icon:"more",size:e(V)},null,8,["color","size"])]),_:1})]),items:t(()=>[s(e($),{item:{to:n.detailViewRoute,label:e(a)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["page-number","page-size","total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"])]),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});const X=I(D,[["__scopeId","data-v-cf4cec8c"]]);export{X as default};
