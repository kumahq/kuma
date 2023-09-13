import{d as g,L as k,r as b,o,g as p,w as s,h as t,p as f,A as h,m as w,C as x,i as a,l,a5 as C,E as S,a6 as z,D as i,ab as L,j as _,F as u,Y as T,a7 as E,a1 as B,H as I,a8 as N,K as A,a9 as R,_ as V,q as $}from"./index-cf0727dc.js";const O=g({__name:"ServiceListView",props:{page:{},size:{},mesh:{}},setup(d){const c=d,{t:r}=k();return(D,K)=>{const v=b("RouterLink");return o(),p(V,{name:"services-list-view"},{default:s(({route:y})=>[t(f,{src:`/meshes/${c.mesh}/service-insights?page=${c.page}&size=${c.size}`},{default:s(({data:n,error:m})=>[t(h,null,{title:s(()=>[w("h2",null,[t(x,{title:a(r)("services.routes.items.title"),render:!0},null,8,["title"])])]),default:s(()=>[l(),t(a(C),null,{body:s(()=>[m!==void 0?(o(),p(S,{key:0,error:m},null,8,["error"])):(o(),p(z,{key:1,class:"service-collection","data-testid":"service-collection","empty-state-message":a(r)("common.emptyState.message",{type:"Services"}),headers:[{label:"Name",key:"name"},{label:"Type",key:"serviceType"},{label:"Address",key:"addressPort"},{label:"DP proxies (online / total)",key:"online"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":c.page,"page-size":c.size,total:n==null?void 0:n.total,items:n==null?void 0:n.items,error:m,onChange:y.update},{name:s(({row:e})=>[t(v,{to:{name:"service-detail-view",params:{service:e.name}}},{default:s(()=>[l(i(e.name),1)]),_:2},1032,["to"])]),serviceType:s(({rowValue:e})=>[l(i(e||"internal"),1)]),addressPort:s(({rowValue:e})=>[e?(o(),p(L,{key:0,text:e},null,8,["text"])):(o(),_(u,{key:1},[l(i(a(r)("common.collection.none")),1)],64))]),online:s(({row:e})=>[e.dataplanes?(o(),_(u,{key:0},[l(i(e.dataplanes.online||0)+" / "+i(e.dataplanes.total||0),1)],64)):(o(),_(u,{key:1},[l(i(a(r)("common.collection.none")),1)],64))]),status:s(({row:e})=>[t(T,{status:e.status||"not_available"},null,8,["status"])]),actions:s(({row:e})=>[t(a(E),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:s(()=>[t(a(B),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:s(()=>[t(a(I),{color:a(N),icon:"more",size:a(A)},null,8,["color","size"])]),_:1})]),items:s(()=>[t(a(R),{item:{to:{name:"service-detail-view",params:{service:e.name}},label:a(r)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["empty-state-message","headers","page-number","page-size","total","items","error","onChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1})}}});const U=$(O,[["__scopeId","data-v-8b46a6c8"]]);export{U as default};
