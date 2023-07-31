import{d as z,r as w,o as m,a as _,w as e,h as s,b as t,s as v,g as u,j as k,t as y,e as h,F as V,$ as x,E,x as C,H as L}from"./index-2a9ba339.js";import{_ as $,A}from"./DataSource.vue_vue_type_script_setup_true_lang-c18fdd22.js";import{g as N,A as B,_ as I,o as R,f as S}from"./RouteView.vue_vue_type_script_setup_true_lang-5d6806ed.js";import{_ as T}from"./RouteTitle.vue_vue_type_script_setup_true_lang-4859e7c4.js";import{S as Z}from"./StatusBadge-4403951c.js";const F=z({__name:"ZoneEgressListView",props:{page:{type:Number,required:!0},size:{type:Number,required:!0}},setup(d){const n=d,{t:i}=N();function g(p){return p.map(l=>{const{name:r}=l,c={name:"zone-egress-detail-view",params:{zoneEgress:r}},a=R(l.zoneEgressInsight??{});return{detailViewRoute:c,name:r,status:a}})}return(p,l)=>{const r=w("RouterLink");return m(),_(I,{name:"zone-egress-list-view"},{default:e(({route:c})=>[s(B,{breadcrumbs:[{to:{name:"zone-egress-list-view"},text:t(i)("zone-egresses.routes.items.breadcrumbs")}]},{title:e(()=>[v("h2",null,[s(T,{title:t(i)("zone-egresses.routes.items.title"),render:!0},null,8,["title"])])]),default:e(()=>[u(),s($,{src:`/zone-egresses?size=${n.size}&page=${n.page}`},{default:e(({data:a,error:b})=>[s(t(k),null,{body:e(()=>[s(A,{class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{label:"Name",key:"name"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":n.page,"page-size":n.size,total:a==null?void 0:a.total,items:a?g(a.items):void 0,error:b,onChange:c.update},{name:e(({row:o,rowValue:f})=>[s(r,{to:o.detailViewRoute,"data-testid":"detail-view-link"},{default:e(()=>[u(y(f),1)]),_:2},1032,["to"])]),status:e(({rowValue:o})=>[o?(m(),_(Z,{key:0,status:o},null,8,["status"])):(m(),h(V,{key:1},[u(`
                  —
                `)],64))]),actions:e(({row:o})=>[s(t(x),{class:"actions-dropdown","data-testid":"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[s(t(E),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:e(()=>[s(t(C),{color:"var(--black-400)",icon:"more",size:"16"})]),_:1})]),items:e(()=>[s(t(L),{item:{to:o.detailViewRoute,label:t(i)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["page-number","page-size","total","items","error","onChange"])]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1})}}});const J=S(F,[["__scopeId","data-v-853de2d9"]]);export{J as default};
