import{g as E,y as x,E as $,q as L,r as N,K as I,f as B,A,i as T,h as S,t as O,_ as V}from"./RouteView.vue_vue_type_script_setup_true_lang-9ff59c45.js";import{d as w,u as K,r as U,o as s,e as f,h as c,w as a,F as z,v as F,n as Y,b as e,g as o,t as n,i as p,k as R,a as r,Y as b,f as v,R as q,H as G,x as H,J}from"./index-6c44d272.js";import{P as X}from"./PolicyTypeTag-d757992d.js";import{n as Z}from"./notEmpty-7f452b20.js";import{_ as j}from"./RouteTitle.vue_vue_type_script_setup_true_lang-ce3600d0.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-7307ec83.js";const M={class:"policy-list-content"},Q={class:"policy-count"},W={class:"policy-list"},D={class:"stack"},ee={class:"description"},te={class:"description-content"},ae={class:"description-actions"},oe={class:"visually-hidden"},se={key:0},ie=w({__name:"PolicyList",props:{pageNumber:{},pageSize:{},policyTypes:{},currentPolicyType:{},policyCollection:{},policyError:{},meshInsight:{}},emits:["change"],setup(P,{emit:u}){const t=P,{t:l}=E(),g=K();return(d,y)=>{const h=U("RouterLink");return s(),f("div",M,[c(e(R),{class:"policy-type-list","data-testid":"policy-type-list"},{body:a(()=>[(s(!0),f(z,null,F(t.policyTypes,(m,_)=>{var i,k,C;return s(),f("div",{key:_,class:Y(["policy-type-link-wrapper",{"policy-type-link-wrapper--is-active":m.path===t.currentPolicyType.path}])},[c(h,{class:"policy-type-link",to:{name:"policies-list-view",params:{mesh:e(g).params.mesh,policyPath:m.path}},"data-testid":`policy-type-link-${m.name}`},{default:a(()=>[o(n(m.name),1)]),_:2},1032,["to","data-testid"]),o(),p("div",Q,n(((C=(k=(i=t.meshInsight)==null?void 0:i.policies)==null?void 0:k[m.name])==null?void 0:C.total)??0),1)],2)}),128))]),_:1}),o(),p("div",W,[p("div",D,[c(e(R),null,{body:a(()=>[p("div",ee,[p("div",te,[p("h3",null,[c(X,{"policy-type":t.currentPolicyType.name},{default:a(()=>[o(n(e(l)("policies.collection.title",{name:t.currentPolicyType.name})),1)]),_:1},8,["policy-type"])]),o(),p("p",null,n(e(l)("policies.collection.description")),1)]),o(),p("div",ae,[t.currentPolicyType.isExperimental?(s(),r(e(b),{key:0,appearance:"warning"},{default:a(()=>[o(n(e(l)("policies.collection.beta")),1)]),_:1})):v("",!0),o(),t.currentPolicyType.isInbound?(s(),r(e(b),{key:1,appearance:"neutral"},{default:a(()=>[o(n(e(l)("policies.collection.inbound")),1)]),_:1})):v("",!0),o(),t.currentPolicyType.isOutbound?(s(),r(e(b),{key:2,appearance:"neutral"},{default:a(()=>[o(n(e(l)("policies.collection.outbound")),1)]),_:1})):v("",!0),o(),c(x,{href:e(l)("policies.href.docs",{name:t.currentPolicyType.name}),"data-testid":"policy-documentation-link"},{default:a(()=>[p("span",oe,n(e(l)("common.documentation")),1)]),_:1},8,["href"])])])]),_:1}),o(),c(e(R),null,{body:a(()=>{var m,_;return[t.policyError!==void 0?(s(),r($,{key:0,error:t.policyError},null,8,["error"])):(s(),r(L,{key:1,class:"policy-collection","data-testid":"policy-collection","empty-state-message":e(l)("common.emptyState.message",{type:`${t.currentPolicyType.name} policies`}),"empty-state-cta-to":e(l)("policies.href.docs",{name:t.currentPolicyType.name}),"empty-state-cta-text":e(l)("common.documentation"),headers:[{label:"Name",key:"name"},t.currentPolicyType.isTargetRefBased?{label:"Target ref",key:"targetRef"}:void 0,{label:"Actions",key:"actions",hideLabel:!0}].filter(e(Z)),"page-number":t.pageNumber,"page-size":t.pageSize,total:(m=t.policyCollection)==null?void 0:m.total,items:(_=t.policyCollection)==null?void 0:_.items,error:t.policyError,onChange:y[0]||(y[0]=i=>u("change",i))},{name:a(({rowValue:i})=>[c(h,{to:{name:"policy-detail-view",params:{mesh:e(g).params.mesh,policyPath:t.currentPolicyType.path,policy:i}}},{default:a(()=>[o(n(i),1)]),_:2},1032,["to"])]),targetRef:a(({row:i})=>[t.currentPolicyType.isTargetRefBased?(s(),r(e(b),{key:0,appearance:"neutral"},{default:a(()=>[o(n(i.spec.targetRef.kind),1),i.spec.targetRef.name?(s(),f("span",se,[o(":"),p("b",null,n(i.spec.targetRef.name),1)])):v("",!0)]),_:2},1024)):(s(),f(z,{key:1},[o(n(e(l)("common.detail.none")),1)],64))]),actions:a(({row:i})=>[c(e(q),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:a(()=>[c(e(G),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:a(()=>[c(e(H),{color:e(N),icon:"more",size:e(I)},null,8,["color","size"])]),_:1})]),items:a(()=>[c(e(J),{item:{to:{name:"policy-detail-view",params:{mesh:e(g).params.mesh,policyPath:t.currentPolicyType.path,policy:i.name}},label:e(l)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:1},8,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error"]))]}),_:1})])])])}}});const ce=B(ie,[["__scopeId","data-v-4f05fbc6"]]),de=w({__name:"PolicyListView",props:{page:{},size:{}},setup(P){const u=P,{t}=E();return(l,g)=>(s(),r(V,{name:"policies-list-view"},{default:a(({route:d})=>[c(A,null,{title:a(()=>[p("h2",null,[c(j,{title:e(t)("policies.routes.items.title"),render:!0},null,8,["title"])])]),default:a(()=>[o(),c(T,{src:"/*/policy-types"},{default:a(({data:y,error:h})=>[h?(s(),r($,{key:0,error:h},null,8,["error"])):y===void 0?(s(),r(S,{key:1})):y.policies.length===0?(s(),r(O,{key:2})):(s(),r(T,{key:3,src:`/meshes/${d.params.mesh}/policy-path/${d.params.policyPath}?page=${u.page}&size=${u.size}`},{default:a(({data:m,error:_})=>[c(T,{src:`/mesh-insights/${d.params.mesh}`},{default:a(({data:i})=>[(s(),r(ce,{key:d.params.policyPath,"page-number":u.page,"page-size":u.size,"current-policy-type":y.policies.find(k=>k.path===d.params.policyPath)??y.policies[0],"policy-types":y.policies,"mesh-insight":i,"policy-collection":m,"policy-error":_,onChange:d.update},null,8,["page-number","page-size","current-policy-type","policy-types","mesh-insight","policy-collection","policy-error","onChange"]))]),_:2},1032,["src"])]),_:2},1032,["src"]))]),_:2},1024)]),_:2},1024)]),_:1}))}});export{de as default};
