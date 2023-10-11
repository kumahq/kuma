import{d as R,a2 as L,e as N,r as g,o as c,j as h,h as n,w as a,F as C,G as S,y as V,i as e,l as i,D as r,m,Y as v,g as p,af as b,k as f,ay as B,E as $,Z as A,$ as K,W as F,a0 as O,K as Z,a1 as j,q,t as D,ae as G}from"./index-ba2f01fe.js";import{P as M}from"./PolicyTypeTag-512a384d.js";const U={class:"policy-list-content"},W={class:"policy-count"},Y={class:"policy-list"},H={class:"stack"},J={class:"description"},Q={class:"description-content"},X={class:"description-actions"},ee={class:"visually-hidden"},te={key:0},ae=R({__name:"PolicyList",props:{pageNumber:{},pageSize:{},policyTypes:{},currentPolicyType:{},policyCollection:{},policyError:{},meshInsight:{}},emits:["change"],setup(T,{emit:z}){const t=T,{t:l}=L(),y=N();return(w,k)=>{const u=g("RouterLink");return c(),h("div",U,[n(e(v),{class:"policy-type-list","data-testid":"policy-type-list"},{body:a(()=>[(c(!0),h(C,null,S(t.policyTypes,(s,d)=>{var o,_,P;return c(),h("div",{key:d,class:V(["policy-type-link-wrapper",{"policy-type-link-wrapper--is-active":s.path===t.currentPolicyType.path}])},[n(u,{class:"policy-type-link",to:{name:"policy-list-view",params:{mesh:e(y).params.mesh,policyPath:s.path}},"data-testid":`policy-type-link-${s.name}`},{default:a(()=>[i(r(s.name),1)]),_:2},1032,["to","data-testid"]),i(),m("div",W,r(((P=(_=(o=t.meshInsight)==null?void 0:o.policies)==null?void 0:_[s.name])==null?void 0:P.total)??0),1)],2)}),128))]),_:1}),i(),m("div",Y,[m("div",H,[n(e(v),null,{body:a(()=>[m("div",J,[m("div",Q,[m("h3",null,[n(M,{"policy-type":t.currentPolicyType.name},{default:a(()=>[i(r(e(l)("policies.collection.title",{name:t.currentPolicyType.name})),1)]),_:1},8,["policy-type"])]),i(),m("p",null,r(e(l)(`policies.type.${t.currentPolicyType.name}.description`,void 0,{defaultMessage:e(l)("policies.collection.description")})),1)]),i(),m("div",X,[t.currentPolicyType.isExperimental?(c(),p(e(b),{key:0,appearance:"warning"},{default:a(()=>[i(r(e(l)("policies.collection.beta")),1)]),_:1})):f("",!0),i(),t.currentPolicyType.isInbound?(c(),p(e(b),{key:1,appearance:"neutral"},{default:a(()=>[i(r(e(l)("policies.collection.inbound")),1)]),_:1})):f("",!0),i(),t.currentPolicyType.isOutbound?(c(),p(e(b),{key:2,appearance:"neutral"},{default:a(()=>[i(r(e(l)("policies.collection.outbound")),1)]),_:1})):f("",!0),i(),n(B,{href:e(l)("policies.href.docs",{name:t.currentPolicyType.name}),"data-testid":"policy-documentation-link"},{default:a(()=>[m("span",ee,r(e(l)("common.documentation")),1)]),_:1},8,["href"])])])]),_:1}),i(),n(e(v),null,{body:a(()=>{var s,d;return[t.policyError!==void 0?(c(),p($,{key:0,error:t.policyError},null,8,["error"])):(c(),p(A,{key:1,class:"policy-collection","data-testid":"policy-collection","empty-state-message":e(l)("common.emptyState.message",{type:`${t.currentPolicyType.name} policies`}),"empty-state-cta-to":e(l)("policies.href.docs",{name:t.currentPolicyType.name}),"empty-state-cta-text":e(l)("common.documentation"),headers:[{label:"Name",key:"name"},...t.currentPolicyType.isTargetRefBased?[{label:"Target ref",key:"targetRef"}]:[],{label:"Actions",key:"actions",hideLabel:!0}],"page-number":t.pageNumber,"page-size":t.pageSize,total:(s=t.policyCollection)==null?void 0:s.total,items:(d=t.policyCollection)==null?void 0:d.items,error:t.policyError,onChange:k[0]||(k[0]=o=>z("change",o))},{name:a(({rowValue:o})=>[n(u,{to:{name:"policy-detail-view",params:{mesh:e(y).params.mesh,policyPath:t.currentPolicyType.path,policy:o}}},{default:a(()=>[i(r(o),1)]),_:2},1032,["to"])]),targetRef:a(({row:o})=>[t.currentPolicyType.isTargetRefBased?(c(),p(e(b),{key:0,appearance:"neutral"},{default:a(()=>[i(r(o.spec.targetRef.kind),1),o.spec.targetRef.name?(c(),h("span",te,[i(":"),m("b",null,r(o.spec.targetRef.name),1)])):f("",!0)]),_:2},1024)):(c(),h(C,{key:1},[i(r(e(l)("common.detail.none")),1)],64))]),actions:a(({row:o})=>[n(e(K),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:a(()=>[n(e(F),{class:"non-visual-button",appearance:"secondary",size:"small"},{default:a(()=>[n(e(O),{size:e(Z)},null,8,["size"])]),_:1})]),items:a(()=>[n(e(j),{item:{to:{name:"policy-detail-view",params:{mesh:e(y).params.mesh,policyPath:t.currentPolicyType.path,policy:o.name}},label:e(l)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:1},8,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error"]))]}),_:1})])])])}}});const se=q(ae,[["__scopeId","data-v-9ebcab5f"]]),ce=R({__name:"PolicyListView",setup(T){return(z,t)=>{const l=g("RouteTitle"),y=g("DataSource"),w=g("AppView"),k=g("RouteView");return c(),p(y,{src:"/me"},{default:a(({data:u})=>[u?(c(),p(k,{key:0,name:"policy-list-view",params:{page:1,size:u.pageSize,mesh:"",policyPath:""}},{default:a(({route:s,t:d})=>[n(w,null,{title:a(()=>[m("h2",null,[n(l,{title:d("policies.routes.items.title"),render:!0},null,8,["title"])])]),default:a(()=>[i(),n(y,{src:"/*/policy-types"},{default:a(({data:o,error:_})=>[_?(c(),p($,{key:0,error:_},null,8,["error"])):o===void 0?(c(),p(D,{key:1})):o.policies.length===0?(c(),p(G,{key:2})):(c(),p(y,{key:3,src:`/meshes/${s.params.mesh}/policy-path/${s.params.policyPath}?page=${s.params.page}&size=${s.params.size}`},{default:a(({data:P,error:x})=>[n(y,{src:`/mesh-insights/${s.params.mesh}`},{default:a(({data:I})=>[(c(),p(se,{key:s.params.policyPath,"page-number":parseInt(s.params.page),"page-size":parseInt(s.params.size),"current-policy-type":o.policies.find(E=>E.path===s.params.policyPath)??o.policies[0],"policy-types":o.policies,"mesh-insight":I,"policy-collection":P,"policy-error":x,onChange:s.update},null,8,["page-number","page-size","current-policy-type","policy-types","mesh-insight","policy-collection","policy-error","onChange"]))]),_:2},1032,["src"])]),_:2},1032,["src"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["params"])):f("",!0)]),_:1})}}});export{ce as default};
