import{A as G,a as M}from"./AccordionList-2af6ecc9.js";import{d as S,a as f,o as e,c as t,m as o,f as s,F as r,C as b,t as i,e as y,w as a,p as B,b as n,W as D,s as H,v as W,_ as j,l as R,k as Y,n as U}from"./index-a04e4171.js";import{C as q}from"./CodeBlock-1370eb60.js";import{P as J}from"./PolicyTypeTag-ad3d6e7d.js";import{T as z}from"./TagList-74d9e162.js";import{t as Q}from"./toYaml-4e00099e.js";import{_ as A}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-89970193.js";import{E as K}from"./ErrorBlock-43034db1.js";import{_ as V}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-ea2db3d2.js";import"./uniqueId-90cc9b93.js";import"./index-fce48c05.js";import"./TextWithCopyButton-7e4c909e.js";import"./CopyButton-882edec4.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-9aa5665b.js";const N=h=>(H("data-v-f3e7afbb"),h=h(),W(),h),X={class:"policies-list"},Z={class:"mesh-gateway-policy-list"},E=N(()=>o("h3",{class:"mb-2"},`
        Gateway policies
      `,-1)),ee={key:0},te=N(()=>o("h3",{class:"mt-6 mb-2"},`
        Listeners
      `,-1)),se=N(()=>o("b",null,"Host",-1)),ae=N(()=>o("h4",{class:"mt-2 mb-2"},`
                Routes
              `,-1)),le={class:"dataplane-policy-header"},ne=N(()=>o("b",null,"Route",-1)),oe=N(()=>o("b",null,"Service",-1)),ie={key:0,class:"badge-list"},ce={class:"mt-1"},re=S({__name:"BuiltinGatewayPolicies",props:{gatewayDataplane:{},policyTypesByName:{}},setup(h){const d=h;return(c,P)=>{const v=f("RouterLink"),k=f("KBadge");return e(),t("div",X,[o("div",Z,[E,s(),c.gatewayDataplane.routePolicies.length>0?(e(),t("ul",ee,[(e(!0),t(r,null,b(c.gatewayDataplane.routePolicies,(u,m)=>(e(),t("li",{key:m},[o("span",null,i(u.type),1),s(`:

          `),y(v,{to:{name:"policy-detail-view",params:{mesh:u.mesh,policyPath:d.policyTypesByName[u.type].path,policy:u.name}}},{default:a(()=>[s(i(u.name),1)]),_:2},1032,["to"])]))),128))])):B("",!0),s(),te,s(),o("div",null,[(e(!0),t(r,null,b(c.gatewayDataplane.listenerEntries,(u,m)=>(e(),t("div",{key:m},[o("div",null,[o("div",null,[se,s(": "+i(u.hostName)+":"+i(u.port)+" ("+i(u.protocol)+`)
            `,1)]),s(),u.routeEntries.length>0?(e(),t(r,{key:0},[ae,s(),y(M,{"initially-open":[],"multiple-open":""},{default:a(()=>[(e(!0),t(r,null,b(u.routeEntries,(_,p)=>(e(),n(G,{key:p},D({"accordion-header":a(()=>[o("div",le,[o("div",null,[o("div",null,[ne,s(": "),y(v,{to:{name:"policy-detail-view",params:{mesh:_.route.mesh,policyPath:d.policyTypesByName[_.route.type].path,policy:_.route.name}}},{default:a(()=>[s(i(_.route.name),1)]),_:2},1032,["to"])]),s(),o("div",null,[oe,s(": "+i(_.service),1)])]),s(),_.origins.length>0?(e(),t("div",ie,[(e(!0),t(r,null,b(_.origins,(l,g)=>(e(),n(k,{key:`${m}-${g}`},{default:a(()=>[s(i(l.type),1)]),_:2},1024))),128))])):B("",!0)])]),_:2},[_.origins.length>0?{name:"accordion-content",fn:a(()=>[o("ul",ce,[(e(!0),t(r,null,b(_.origins,(l,g)=>(e(),t("li",{key:`${m}-${g}`},[s(i(l.type)+`:

                        `,1),y(v,{to:{name:"policy-detail-view",params:{mesh:l.mesh,policyPath:d.policyTypesByName[l.type].path,policy:l.name}}},{default:a(()=>[s(i(l.name),1)]),_:2},1032,["to"])]))),128))])]),key:"0"}:void 0]),1024))),128))]),_:2},1024)],64)):B("",!0)])]))),128))])])])}}});const pe=j(re,[["__scopeId","data-v-f3e7afbb"]]),ue={class:"policy-type-heading"},de={class:"policy-list"},_e={key:0},ye=S({__name:"PolicyTypeEntryList",props:{id:{},policyTypeEntries:{},policyTypesByName:{}},setup(h){const d=h;function c({headerKey:P}){return{class:`cell-${P}`}}return(P,v)=>{const k=f("RouterLink"),u=f("KTable");return e(),n(M,{"initially-open":[],"multiple-open":""},{default:a(()=>[(e(!0),t(r,null,b(d.policyTypeEntries,(m,_)=>(e(),n(G,{key:_},{"accordion-header":a(()=>[o("h3",ue,[y(J,{"policy-type":m.type},{default:a(()=>[s(i(m.type)+" ("+i(m.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":a(()=>[o("div",de,[y(u,{class:"policy-type-table",fetcher:()=>({data:m.connections,total:m.connections.length}),headers:[{label:"From",key:"sourceTags"},{label:"To",key:"destinationTags"},{label:"On",key:"name"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}],"cell-attrs":c,"disable-pagination":"","is-clickable":""},{sourceTags:a(({row:p})=>[p.sourceTags.length>0?(e(),n(z,{key:0,class:"tag-list","should-truncate":"",tags:p.sourceTags},null,8,["tags"])):(e(),t(r,{key:1},[s(`
                —
              `)],64))]),destinationTags:a(({row:p})=>[p.destinationTags.length>0?(e(),n(z,{key:0,class:"tag-list","should-truncate":"",tags:p.destinationTags},null,8,["tags"])):(e(),t(r,{key:1},[s(`
                —
              `)],64))]),name:a(({row:p})=>[p.name!==null?(e(),t(r,{key:0},[s(i(p.name),1)],64)):(e(),t(r,{key:1},[s(`
                —
              `)],64))]),origins:a(({row:p})=>[p.origins.length>0?(e(),t("ul",_e,[(e(!0),t(r,null,b(p.origins,(l,g)=>(e(),t("li",{key:`${_}-${g}`},[y(k,{to:{name:"policy-detail-view",params:{mesh:l.mesh,policyPath:d.policyTypesByName[l.type].path,policy:l.name}}},{default:a(()=>[s(i(l.name),1)]),_:2},1032,["to"])]))),128))])):(e(),t(r,{key:1},[s(`
                —
              `)],64))]),config:a(({row:p,rowKey:l})=>[p.config?(e(),n(q,{key:0,id:`${d.id}-${_}-${l}-code-block`,code:R(Q)(p.config),language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),t(r,{key:1},[s(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const me=j(ye,[["__scopeId","data-v-b4ad75df"]]),he=h=>(H("data-v-2b6b806e"),h=h(),W(),h),ge={class:"policy-type-heading"},ke={class:"policy-list"},fe={key:0,class:"matcher"},be={key:0,class:"matcher__and"},ve=he(()=>o("br",null,null,-1)),$e={key:1,class:"matcher__not"},Te={class:"matcher__term"},Re={key:1},Be={key:0},Pe=S({__name:"RuleEntryList",props:{id:{},ruleEntries:{},policyTypesByName:{},showMatchers:{type:Boolean,default:!0}},setup(h){const{t:d}=Y(),c=h;function P({headerKey:v}){return{class:`cell-${v}`}}return(v,k)=>{const u=f("RouterLink"),m=f("KTable");return e(),n(M,{"initially-open":[],"multiple-open":""},{default:a(()=>[(e(!0),t(r,null,b(c.ruleEntries,(_,p)=>(e(),n(G,{key:p},{"accordion-header":a(()=>[o("h3",ge,[y(J,{"policy-type":_.type},{default:a(()=>[s(i(_.type),1)]),_:2},1032,["policy-type"])])]),"accordion-content":a(()=>[o("div",ke,[y(m,{class:U(["policy-type-table",{"has-matchers":c.showMatchers}]),fetcher:()=>({data:_.rules,total:_.rules.length}),headers:[...c.showMatchers?[{label:"Matchers",key:"matchers"}]:[],{label:"Origin policies",key:"origins"},{label:"Conf",key:"config"}],"cell-attrs":P,"disable-pagination":""},D({origins:a(({row:l})=>[l.origins.length>0?(e(),t("ul",Be,[(e(!0),t(r,null,b(l.origins,(g,T)=>(e(),t("li",{key:`${p}-${T}`},[y(u,{to:{name:"policy-detail-view",params:{mesh:g.mesh,policyPath:c.policyTypesByName[g.type].path,policy:g.name}}},{default:a(()=>[s(i(g.name),1)]),_:2},1032,["to"])]))),128))])):(e(),t(r,{key:1},[s(i(R(d)("common.collection.none")),1)],64))]),config:a(({row:l,rowKey:g})=>[l.config?(e(),n(q,{key:0,id:`${c.id}-${p}-${g}-code-block`,code:R(Q)(l.config),language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),t(r,{key:1},[s(i(R(d)("common.collection.none")),1)],64))]),_:2},[c.showMatchers?{name:"matchers",fn:a(({row:l})=>[l.matchers&&l.matchers.length>0?(e(),t("span",fe,[(e(!0),t(r,null,b(l.matchers,({key:g,value:T,not:w},$)=>(e(),t(r,{key:$},[$>0?(e(),t("span",be,[s(" and"),ve])):B("",!0),w?(e(),t("span",$e,"!")):B("",!0),o("span",Te,i(`${g}:${T}`),1)],64))),128))])):(e(),t("i",Re,i(R(d)("data-planes.routes.item.matches_everything")),1))]),key:"0"}:void 0]),1032,["class","fetcher","headers"])])]),_:2},1024))),128))]),_:1})}}});const O=j(Pe,[["__scopeId","data-v-2b6b806e"]]),we={class:"stack"},Le={class:"mb-2"},Ne=S({__name:"StandardDataplanePolicies",props:{inspectRulesForDataplane:{},policyTypesByName:{}},setup(h){const{t:d}=Y(),c=h;return(P,v)=>{const k=f("KCard");return e(),t("div",we,[c.inspectRulesForDataplane.proxyRules.length>0?(e(),n(k,{key:0},{default:a(()=>[o("h3",null,i(R(d)("data-planes.routes.item.proxy_rule")),1),s(),y(O,{id:"proxy-rules",class:"mt-2","rule-entries":c.inspectRulesForDataplane.proxyRules,"policy-types-by-name":c.policyTypesByName,"show-matchers":!1,"data-testid":"proxy-rule-list"},null,8,["rule-entries","policy-types-by-name"])]),_:1})):B("",!0),s(),c.inspectRulesForDataplane.toRules.length>0?(e(),n(k,{key:1},{default:a(()=>[o("h3",null,i(R(d)("data-planes.routes.item.to_rules")),1),s(),y(O,{id:"to-rules",class:"mt-2","rule-entries":c.inspectRulesForDataplane.toRules,"policy-types-by-name":c.policyTypesByName,"data-testid":"to-rule-list"},null,8,["rule-entries","policy-types-by-name"])]),_:1})):B("",!0),s(),c.inspectRulesForDataplane.fromRuleInbounds.length>0?(e(),n(k,{key:2},{default:a(()=>[o("h3",Le,i(R(d)("data-planes.routes.item.from_rules")),1),s(),(e(!0),t(r,null,b(c.inspectRulesForDataplane.fromRuleInbounds,(u,m)=>(e(),t("div",{key:m},[o("h4",null,i(R(d)("data-planes.routes.item.port",{port:u.port})),1),s(),y(O,{id:`from-rules-${m}`,class:"mt-2","rule-entries":u.ruleEntries,"policy-types-by-name":c.policyTypesByName,"data-testid":`from-rule-list-${m}`},null,8,["id","rule-entries","policy-types-by-name","data-testid"])]))),128))]),_:1})):B("",!0)])}}}),Ce={class:"stack"},We=S({__name:"DataPlanePoliciesView",props:{data:{}},setup(h){const d=h;return(c,P)=>{const v=f("RouteTitle"),k=f("DataSource"),u=f("KCard"),m=f("AppView"),_=f("RouteView");return e(),n(_,{name:"data-plane-policies-view",params:{mesh:"",dataPlane:""}},{default:a(({can:p,route:l,t:g})=>[y(m,null,{title:a(()=>[o("h2",null,[y(v,{title:g("data-planes.routes.item.navigation.data-plane-policies-view")},null,8,["title"])])]),default:a(()=>[s(),o("div",Ce,[y(k,{src:"/*/policy-types"},{default:a(({data:T,error:w})=>[y(k,{src:`/meshes/${l.params.mesh}/dataplanes/${l.params.dataPlane}/rules`},{default:a(({data:$,error:C})=>[w?(e(),n(K,{key:0,error:w},null,8,["error"])):C?(e(),n(K,{key:1,error:C},null,8,["error"])):T===void 0||$===void 0?(e(),n(V,{key:2})):$.rules.length===0?(e(),n(A,{key:3})):(e(),n(Ne,{key:4,"policy-types-by-name":T.policies.reduce((L,F)=>Object.assign(L,{[F.name]:F}),{}),"inspect-rules-for-dataplane":$,"data-testid":"rules-based-policies"},null,8,["policy-types-by-name","inspect-rules-for-dataplane"]))]),_:2},1032,["src"]),s(),p("use zones")?B("",!0):(e(),n(k,{key:0,src:d.data.dataplaneType!=="builtin"?`/meshes/${l.params.mesh}/dataplanes/${l.params.dataPlane}/sidecar-dataplane-policies`:""},{default:a(({data:$,error:C})=>[y(k,{src:d.data.dataplaneType==="builtin"?`/meshes/${l.params.mesh}/dataplanes/${l.params.dataPlane}/gateway-dataplane-policies`:""},{default:a(({data:L,error:F})=>[o("div",null,[o("h3",null,i(g("data-planes.routes.item.legacy_policies")),1),s(),w?(e(),n(K,{key:0,error:w},null,8,["error"])):C?(e(),n(K,{key:1,error:C},null,8,["error"])):F?(e(),n(K,{key:2,error:F},null,8,["error"])):T===void 0?(e(),n(V,{key:3})):d.data.dataplaneType==="builtin"?(e(),t(r,{key:4},[L===void 0?(e(),n(V,{key:0})):L.routePolicies.length===0&&L.listenerEntries.length===0?(e(),n(A,{key:1})):(e(),n(u,{key:2,class:"mt-4"},{default:a(()=>[y(pe,{"policy-types-by-name":T.policies.reduce((x,I)=>Object.assign(x,{[I.name]:I}),{}),"gateway-dataplane":L,"data-testid":"builtin-gateway-dataplane-policies"},null,8,["policy-types-by-name","gateway-dataplane"])]),_:2},1024))],64)):(e(),t(r,{key:5},[$===void 0?(e(),n(V,{key:0})):$.policyTypeEntries.length===0?(e(),n(A,{key:1})):(e(),n(u,{key:2,class:"mt-4"},{default:a(()=>[y(me,{id:"policies","policy-type-entries":$.policyTypeEntries,"policy-types-by-name":T.policies.reduce((x,I)=>Object.assign(x,{[I.name]:I}),{}),"data-testid":"sidecar-dataplane-policies"},null,8,["policy-type-entries","policy-types-by-name"])]),_:2},1024))],64))])]),_:2},1032,["src"])]),_:2},1032,["src"]))]),_:2},1024)])]),_:2},1024)]),_:1})}}});export{We as default};
