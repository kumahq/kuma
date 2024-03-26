import{d as S,a as r,o as i,b as m,w as e,e as c,a2 as $,f as t,t as l,c as h,F as k,G as K,m as d,q as u,p as v,V as N,K as I,D as A,_ as q}from"./index-CasyeFB_.js";import{D as E,A as M}from"./AppCollection-BOBuExbI.js";import{P as F}from"./PolicyTypeTag-1tRJ0gQK.js";import{S as H}from"./SummaryView-ESI-BYLO.js";const O={class:"stack"},Z={class:"visually-hidden"},G=["innerHTML"],U={key:0},j=S({__name:"PolicyListView",props:{policyTypes:{}},setup(C){const R=C;return(J,Q)=>{const f=r("KBadge"),z=r("KCard"),g=r("RouterLink"),V=r("RouterView"),D=r("DataLoader"),L=r("AppView"),P=r("DataCollection"),w=r("RouteView"),B=r("DataSource");return i(),m(B,{src:"/me"},{default:e(({data:b})=>[b?(i(),m(w,{key:0,name:"policy-list-view",params:{page:1,size:b.pageSize,mesh:"",policyPath:"",policy:""}},{default:e(({route:s,t:o,can:x})=>[c(P,{predicate:y=>typeof y<"u"&&y.path===s.params.policyPath,find:!0,items:R.policyTypes??[void 0]},{empty:e(()=>[c($,null,{message:e(()=>[t(l(o("policies.routes.items.empty")),1)]),_:2},1024)]),default:e(({items:y})=>[(i(!0),h(k,null,K([y[0]],n=>(i(),m(L,{key:n},{default:e(()=>[d("div",O,[c(z,null,{default:e(()=>[d("header",null,[d("div",null,[n.isExperimental?(i(),m(f,{key:0,appearance:"warning"},{default:e(()=>[t(l(o("policies.collection.beta")),1)]),_:2},1024)):u("",!0),t(),n.isInbound?(i(),m(f,{key:1,appearance:"neutral"},{default:e(()=>[t(l(o("policies.collection.inbound")),1)]),_:2},1024)):u("",!0),t(),n.isOutbound?(i(),m(f,{key:2,appearance:"neutral"},{default:e(()=>[t(l(o("policies.collection.outbound")),1)]),_:2},1024)):u("",!0),t(),c(E,{href:o("policies.href.docs",{name:n.name}),"data-testid":"policy-documentation-link"},{default:e(()=>[d("span",Z,l(o("common.documentation")),1)]),_:2},1032,["href"])]),t(),d("h3",null,[c(F,{"policy-type":n.name},{default:e(()=>[t(l(o("policies.collection.title",{name:n.name})),1)]),_:2},1032,["policy-type"])])]),t(),d("div",{innerHTML:o(`policies.type.${n.name}.description`,void 0,{defaultMessage:o("policies.collection.description")})},null,8,G)]),_:2},1024),t(),c(z,null,{default:e(()=>[c(D,{src:`/meshes/${s.params.mesh}/policy-path/${s.params.policyPath}?page=${s.params.page}&size=${s.params.size}`,loader:!1},{default:e(({data:p})=>[c(M,{class:"policy-collection","data-testid":"policy-collection","empty-state-message":o("common.emptyState.message",{type:`${n.name} policies`}),"empty-state-cta-to":o("policies.href.docs",{name:n.name}),"empty-state-cta-text":o("common.documentation"),headers:[{label:"Name",key:"name"},...x("use zones")&&n.isTargetRefBased?[{label:"Zone",key:"zone"}]:[],...n.isTargetRefBased?[{label:"Target ref",key:"targetRef"}]:[],{label:"Details",key:"details",hideLabel:!0}],"page-number":s.params.page,"page-size":s.params.size,total:p==null?void 0:p.total,items:p==null?void 0:p.items,"is-selected-row":a=>a.name===s.params.policy,onChange:s.update},{name:e(({row:a})=>[c(g,{to:{name:"policy-summary-view",params:{mesh:a.mesh,policyPath:n.path,policy:a.name},query:{page:s.params.page,size:s.params.size}}},{default:e(()=>[t(l(a.name),1)]),_:2},1032,["to"])]),targetRef:e(({row:a})=>{var _;return[n.isTargetRefBased&&typeof((_=a.spec)==null?void 0:_.targetRef)<"u"?(i(),m(f,{key:0,appearance:"neutral"},{default:e(()=>[t(l(a.spec.targetRef.kind),1),a.spec.targetRef.name?(i(),h("span",U,[t(":"),d("b",null,l(a.spec.targetRef.name),1)])):u("",!0)]),_:2},1024)):(i(),h(k,{key:1},[t(l(o("common.detail.none")),1)],64))]}),zone:e(({row:a})=>[a.labels&&a.labels["kuma.io/origin"]==="zone"&&a.labels["kuma.io/zone"]?(i(),m(g,{key:0,to:{name:"zone-cp-detail-view",params:{zone:a.labels["kuma.io/zone"]}}},{default:e(()=>[t(l(a.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):(i(),h(k,{key:1},[t(l(o("common.detail.none")),1)],64))]),details:e(({row:a})=>[c(g,{class:"details-link","data-testid":"details-link",to:{name:"policy-detail-view",params:{mesh:a.mesh,policyPath:n.path,policy:a.name}}},{default:e(()=>[t(l(o("common.collection.details_link"))+" ",1),c(v(N),{decorative:"",size:v(I)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","is-selected-row","onChange"]),t(),s.params.policy?(i(),m(V,{key:0},{default:e(({Component:a})=>[c(H,{onClose:_=>s.replace({name:"policy-list-view",params:{mesh:s.params.mesh,policyPath:s.params.policyPath},query:{page:s.params.page,size:s.params.size}})},{default:e(()=>[(i(),m(A(a),{policy:p==null?void 0:p.items.find(_=>_.name===s.params.policy),"policy-type":n},null,8,["policy","policy-type"]))]),_:2},1032,["onClose"])]),_:2},1024)):u("",!0)]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024))),128))]),_:2},1032,["predicate","items"])]),_:2},1032,["params"])):u("",!0)]),_:1})}}}),ee=q(j,[["__scopeId","data-v-c519cdd6"]]);export{ee as default};
