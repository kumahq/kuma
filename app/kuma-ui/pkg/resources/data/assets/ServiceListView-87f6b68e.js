import{f,d as k,A as b,E as g,o as h,T as w,S as x,p as S,K as C,_ as T,e as z}from"./RouteView.vue_vue_type_script_setup_true_lang-c4937c78.js";import{d as E,r as L,o,a as m,w as s,b as a,i as R,e as t,h as i,k as A,t as l,f as _,F as u,R as B,H as N,x as I,J as V}from"./index-d6e04092.js";import{_ as $}from"./RouteTitle.vue_vue_type_script_setup_true_lang-11eb67ee.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-c618043c.js";const O=E({__name:"ServiceListView",props:{page:{},size:{},mesh:{}},setup(d){const r=d,{t:c}=f();return(D,K)=>{const v=L("RouterLink");return o(),m(T,{name:"services-list-view"},{default:s(({route:y})=>[a(k,{src:`/meshes/${r.mesh}/service-insights?page=${r.page}&size=${r.size}`},{default:s(({data:n,error:p})=>[a(b,null,{title:s(()=>[R("h2",null,[a($,{title:t(c)("services.routes.items.title"),render:!0},null,8,["title"])])]),default:s(()=>[i(),a(t(A),null,{body:s(()=>[p!==void 0?(o(),m(g,{key:0,error:p},null,8,["error"])):(o(),m(h,{key:1,class:"service-collection","data-testid":"service-collection","empty-state-message":t(c)("common.emptyState.message",{type:"Services"}),headers:[{label:"Name",key:"name"},{label:"Type",key:"serviceType"},{label:"Address",key:"addressPort"},{label:"DP proxies (online / total)",key:"online"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":r.page,"page-size":r.size,total:n==null?void 0:n.total,items:n==null?void 0:n.items,error:p,onChange:y.update},{name:s(({row:e})=>[a(v,{to:{name:"service-detail-view",params:{service:e.name}}},{default:s(()=>[i(l(e.name),1)]),_:2},1032,["to"])]),serviceType:s(({rowValue:e})=>[i(l(e||"internal"),1)]),addressPort:s(({rowValue:e})=>[e?(o(),m(w,{key:0,text:e},null,8,["text"])):(o(),_(u,{key:1},[i(l(t(c)("common.collection.none")),1)],64))]),online:s(({row:e})=>[e.dataplanes?(o(),_(u,{key:0},[i(l(e.dataplanes.online||0)+" / "+l(e.dataplanes.total||0),1)],64)):(o(),_(u,{key:1},[i(l(t(c)("common.collection.none")),1)],64))]),status:s(({row:e})=>[a(x,{status:e.status||"not_available"},null,8,["status"])]),actions:s(({row:e})=>[a(t(B),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:s(()=>[a(t(N),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:s(()=>[a(t(I),{color:t(S),icon:"more",size:t(C)},null,8,["color","size"])]),_:1})]),items:s(()=>[a(t(V),{item:{to:{name:"service-detail-view",params:{service:e.name}},label:t(c)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["empty-state-message","headers","page-number","page-size","total","items","error","onChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1})}}});const H=z(O,[["__scopeId","data-v-8b46a6c8"]]);export{H as default};
