import{d as D,i as n,o as i,a as m,w as e,j as t,g as b,t as p,k as l,A as d,aw as v,b as P,H as N,K as S,r as R,e as _,_ as x}from"./index-B4OAi35c.js";import{p as A}from"./kong-icons.es249-CTBvwfdU.js";import{A as L}from"./AppCollection-D_06DYjA.js";import{S as B}from"./SummaryView-BnRgi_m1.js";import"./kong-icons.es245-uMhO6AtE.js";const I=D({__name:"PolicyDetailView",setup(K){return(E,Z)=>{const c=n("RouterLink"),y=n("DataCollection"),f=n("RouterView"),h=n("DataLoader"),w=n("KCard"),z=n("AppView"),k=n("RouteView"),g=n("DataSource");return i(),m(g,{src:"/me"},{default:e(({data:u})=>[u?(i(),m(k,{key:0,name:"policy-detail-view",params:{mesh:"",policy:"",policyPath:"",dataPlane:"",s:"",page:1,size:u.pageSize}},{default:e(({route:a,t:r,uri:C,can:V})=>[t(z,null,{title:e(()=>[b("h2",null,p(r("policies.detail.affected_dpps")),1)]),default:e(()=>[l(),t(w,null,{default:e(()=>[t(h,{src:C(d(v),"/meshes/:mesh/policy-path/:path/policy/:name/dataplanes",{mesh:a.params.mesh,path:a.params.policyPath,name:a.params.policy},{page:a.params.page,size:a.params.size})},{loadable:e(({data:o})=>[t(y,{type:"data-planes",items:(o==null?void 0:o.items)??[void 0]},{default:e(()=>[t(L,{"page-number":a.params.page,"page-size":a.params.size,headers:[{label:"Name",key:"name"},{label:"Namespace",key:"namespace"},...V("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"Details",key:"details",hideLabel:!0}],items:o==null?void 0:o.items,total:o==null?void 0:o.total,"is-selected-row":s=>s.id===a.params.dataPlane,onChange:a.update},{name:e(({row:s})=>[t(c,{"data-action":"",to:{name:"data-plane-detail-view",params:{dataPlane:s.id}}},{default:e(()=>[l(p(s.name),1)]),_:2},1032,["to"])]),namespace:e(({row:s})=>[l(p(s.namespace),1)]),zone:e(({row:s})=>[s.zone?(i(),m(c,{key:0,to:{name:"zone-cp-detail-view",params:{zone:s.zone}}},{default:e(()=>[l(p(s.zone),1)]),_:2},1032,["to"])):(i(),P(N,{key:1},[l(p(r("common.collection.none")),1)],64))]),details:e(({row:s})=>[t(c,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:s.id}}},{default:e(()=>[l(p(r("common.collection.details_link"))+" ",1),t(d(A),{decorative:"",size:d(S)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["page-number","page-size","headers","items","total","is-selected-row","onChange"])]),_:2},1032,["items"]),l(),t(f,null,{default:e(({Component:s})=>[a.child()?(i(),m(B,{key:0,onClose:j=>a.replace({params:{mesh:a.params.mesh},query:{page:a.params.page,size:a.params.size,s:a.params.s}})},{default:e(()=>[typeof o<"u"?(i(),m(R(s),{key:0,items:o.items},null,8,["items"])):_("",!0)]),_:2},1032,["onClose"])):_("",!0)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:2},1032,["params"])):_("",!0)]),_:1})}}}),U=x(I,[["__scopeId","data-v-4db5e959"]]);export{U as default};
