import{d as $,a as m,o as i,b as p,w as e,e as c,ab as N,f as t,t as l,c as y,F as k,G as A,m as d,q as u,p as v,Q as K,K as I,E,_ as q}from"./index-T0BkiAMa.js";import{A as M}from"./AppCollection--hpbcKQi.js";import{P as F}from"./PolicyTypeTag-DEQcAbFf.js";import{S as H}from"./SummaryView-DIa7F5ow.js";const O={class:"stack"},X={class:"visually-hidden"},Z=["innerHTML"],G={key:0},Q=$({__name:"PolicyListView",props:{policyTypes:{}},setup(C){const R=C;return(U,j)=>{const _=m("KBadge"),h=m("XAction"),b=m("KCard"),V=m("RouterLink"),P=m("RouterView"),w=m("DataLoader"),L=m("AppView"),x=m("DataCollection"),B=m("RouteView"),D=m("DataSource");return i(),p(D,{src:"/me"},{default:e(({data:z})=>[z?(i(),p(B,{key:0,name:"policy-list-view",params:{page:1,size:z.pageSize,mesh:"",policyPath:"",policy:""}},{default:e(({route:s,t:n,can:S})=>[c(x,{predicate:f=>typeof f<"u"&&f.path===s.params.policyPath,find:!0,items:R.policyTypes??[void 0]},{empty:e(()=>[c(N,null,{message:e(()=>[t(l(n("policies.routes.items.empty")),1)]),_:2},1024)]),default:e(({items:f})=>[(i(!0),y(k,null,A([f[0]],o=>(i(),p(L,{key:o},{default:e(()=>[d("div",O,[c(b,null,{default:e(()=>[d("header",null,[d("div",null,[o.isExperimental?(i(),p(_,{key:0,appearance:"warning"},{default:e(()=>[t(l(n("policies.collection.beta")),1)]),_:2},1024)):u("",!0),t(),o.isInbound?(i(),p(_,{key:1,appearance:"neutral"},{default:e(()=>[t(l(n("policies.collection.inbound")),1)]),_:2},1024)):u("",!0),t(),o.isOutbound?(i(),p(_,{key:2,appearance:"neutral"},{default:e(()=>[t(l(n("policies.collection.outbound")),1)]),_:2},1024)):u("",!0),t(),c(h,{type:"docs",href:n("policies.href.docs",{name:o.name}),"data-testid":"policy-documentation-link"},{default:e(()=>[d("span",X,l(n("common.documentation")),1)]),_:2},1032,["href"])]),t(),d("h3",null,[c(F,{"policy-type":o.name},{default:e(()=>[t(l(n("policies.collection.title",{name:o.name})),1)]),_:2},1032,["policy-type"])])]),t(),d("div",{innerHTML:n(`policies.type.${o.name}.description`,void 0,{defaultMessage:n("policies.collection.description")})},null,8,Z)]),_:2},1024),t(),c(b,null,{default:e(()=>[c(w,{src:`/meshes/${s.params.mesh}/policy-path/${s.params.policyPath}?page=${s.params.page}&size=${s.params.size}`,loader:!1},{default:e(({data:r})=>[c(M,{class:"policy-collection","data-testid":"policy-collection","empty-state-message":n("common.emptyState.message",{type:`${o.name} policies`}),"empty-state-cta-to":n("policies.href.docs",{name:o.name}),"empty-state-cta-text":n("common.documentation"),headers:[{label:"Name",key:"name"},{label:"Namespace",key:"namespace"},...S("use zones")&&o.isTargetRefBased?[{label:"Zone",key:"zone"}]:[],...o.isTargetRefBased?[{label:"Target ref",key:"targetRef"}]:[],{label:"Details",key:"details",hideLabel:!0}],"page-number":s.params.page,"page-size":s.params.size,total:r==null?void 0:r.total,items:r==null?void 0:r.items,"is-selected-row":a=>a.id===s.params.policy,onChange:s.update},{name:e(({row:a})=>[c(h,{to:{name:"policy-summary-view",params:{mesh:a.mesh,policyPath:o.path,policy:a.id},query:{page:s.params.page,size:s.params.size}}},{default:e(()=>[t(l(a.name),1)]),_:2},1032,["to"])]),namespace:e(({row:a})=>[t(l(a.namespace.length>0?a.namespace:n("common.detail.none")),1)]),targetRef:e(({row:a})=>{var g;return[o.isTargetRefBased&&typeof((g=a.spec)==null?void 0:g.targetRef)<"u"?(i(),p(_,{key:0,appearance:"neutral"},{default:e(()=>[t(l(a.spec.targetRef.kind),1),a.spec.targetRef.name?(i(),y("span",G,[t(":"),d("b",null,l(a.spec.targetRef.name),1)])):u("",!0)]),_:2},1024)):(i(),y(k,{key:1},[t(l(n("common.detail.none")),1)],64))]}),zone:e(({row:a})=>[a.labels&&a.labels["kuma.io/origin"]==="zone"&&a.labels["kuma.io/zone"]?(i(),p(V,{key:0,to:{name:"zone-cp-detail-view",params:{zone:a.labels["kuma.io/zone"]}}},{default:e(()=>[t(l(a.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):(i(),y(k,{key:1},[t(l(n("common.detail.none")),1)],64))]),details:e(({row:a})=>[c(h,{class:"details-link","data-testid":"details-link",to:{name:"policy-detail-view",params:{mesh:a.mesh,policyPath:o.path,policy:a.id}}},{default:e(()=>[t(l(n("common.collection.details_link"))+" ",1),c(v(K),{decorative:"",size:v(I)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","is-selected-row","onChange"]),t(),s.params.policy?(i(),p(P,{key:0},{default:e(({Component:a})=>[c(H,{onClose:g=>s.replace({name:"policy-list-view",params:{mesh:s.params.mesh,policyPath:s.params.policyPath},query:{page:s.params.page,size:s.params.size}})},{default:e(()=>[typeof r<"u"?(i(),p(E(a),{key:0,items:r.items,"policy-type":o},null,8,["items","policy-type"])):u("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):u("",!0)]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024))),128))]),_:2},1032,["predicate","items"])]),_:2},1032,["params"])):u("",!0)]),_:1})}}}),ee=q(Q,[["__scopeId","data-v-114c58b9"]]);export{ee as default};
