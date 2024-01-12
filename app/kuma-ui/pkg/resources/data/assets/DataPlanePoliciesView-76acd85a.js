import{A as V,a as x}from"./AccordionList-6c3fcb31.js";import{d as I,a as k,o as e,c as t,m as n,f as a,F as u,D as b,t as c,e as y,w as s,p as R,b as i,W as z,v as D,x as j,_ as A,l as T,k as H,n as J}from"./index-cf10d15e.js";import{_ as W}from"./CodeBlock.vue_vue_type_style_index_0_lang-4bf3aea6.js";import{P as Y}from"./PolicyTypeTag-83e2890f.js";import{T as M}from"./TagList-a050d243.js";import{t as q}from"./toYaml-4e00099e.js";import{_ as Q}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-dae16a2a.js";import{E as F}from"./ErrorBlock-ce60392d.js";import{_ as O}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-df4338bd.js";import"./uniqueId-90cc9b93.js";import"./index-fce48c05.js";import"./TextWithCopyButton-b8bd594c.js";import"./CopyButton-0aa5d830.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-990b7d32.js";const L=h=>(D("data-v-ab96b2ac"),h=h(),j(),h),U={class:"policies-list","data-testid":"builtin-gateway-dataplane-policies"},X={class:"mesh-gateway-policy-list"},Z=L(()=>n("h3",{class:"mb-2"},`
        Gateway policies
      `,-1)),ee={key:0},te=L(()=>n("h3",{class:"mt-6 mb-2"},`
        Listeners
      `,-1)),se=L(()=>n("b",null,"Host",-1)),ae=L(()=>n("h4",{class:"mt-2 mb-2"},`
                Routes
              `,-1)),le={class:"dataplane-policy-header"},oe=L(()=>n("b",null,"Route",-1)),ne=L(()=>n("b",null,"Service",-1)),ie={key:0,class:"badge-list"},ce={class:"mt-1"},pe=I({__name:"BuiltinGatewayPolicies",props:{gatewayDataplane:{},policyTypesByName:{}},setup(h){const d=h;return(o,P)=>{const v=k("RouterLink"),f=k("KBadge");return e(),t("div",U,[n("div",X,[Z,a(),o.gatewayDataplane.routePolicies.length>0?(e(),t("ul",ee,[(e(!0),t(u,null,b(o.gatewayDataplane.routePolicies,(p,m)=>(e(),t("li",{key:m},[n("span",null,c(p.type),1),a(`:

          `),y(v,{to:{name:"policy-detail-view",params:{mesh:p.mesh,policyPath:d.policyTypesByName[p.type].path,policy:p.name}}},{default:s(()=>[a(c(p.name),1)]),_:2},1032,["to"])]))),128))])):R("",!0),a(),te,a(),n("div",null,[(e(!0),t(u,null,b(o.gatewayDataplane.listenerEntries,(p,m)=>(e(),t("div",{key:m},[n("div",null,[n("div",null,[se,a(": "+c(p.hostName)+":"+c(p.port)+" ("+c(p.protocol)+`)
            `,1)]),a(),p.routeEntries.length>0?(e(),t(u,{key:0},[ae,a(),y(x,{"initially-open":[],"multiple-open":""},{default:s(()=>[(e(!0),t(u,null,b(p.routeEntries,(_,r)=>(e(),i(V,{key:r},z({"accordion-header":s(()=>[n("div",le,[n("div",null,[n("div",null,[oe,a(": "),y(v,{to:{name:"policy-detail-view",params:{mesh:_.route.mesh,policyPath:d.policyTypesByName[_.route.type].path,policy:_.route.name}}},{default:s(()=>[a(c(_.route.name),1)]),_:2},1032,["to"])]),a(),n("div",null,[ne,a(": "+c(_.service),1)])]),a(),_.origins.length>0?(e(),t("div",ie,[(e(!0),t(u,null,b(_.origins,(l,g)=>(e(),i(f,{key:`${m}-${g}`},{default:s(()=>[a(c(l.type),1)]),_:2},1024))),128))])):R("",!0)])]),_:2},[_.origins.length>0?{name:"accordion-content",fn:s(()=>[n("ul",ce,[(e(!0),t(u,null,b(_.origins,(l,g)=>(e(),t("li",{key:`${m}-${g}`},[a(c(l.type)+`:

                        `,1),y(v,{to:{name:"policy-detail-view",params:{mesh:l.mesh,policyPath:d.policyTypesByName[l.type].path,policy:l.name}}},{default:s(()=>[a(c(l.name),1)]),_:2},1032,["to"])]))),128))])]),key:"0"}:void 0]),1024))),128))]),_:2},1024)],64)):R("",!0)])]))),128))])])])}}});const re=A(pe,[["__scopeId","data-v-ab96b2ac"]]),ue={class:"policy-type-heading"},ye={class:"policy-list"},_e={key:0},de=I({__name:"PolicyTypeEntryList",props:{id:{},policyTypeEntries:{},policyTypesByName:{}},setup(h){const d=h;function o({headerKey:P}){return{class:`cell-${P}`}}return(P,v)=>{const f=k("RouterLink"),p=k("KTable");return e(),i(x,{"initially-open":[],"multiple-open":""},{default:s(()=>[(e(!0),t(u,null,b(d.policyTypeEntries,(m,_)=>(e(),i(V,{key:_},{"accordion-header":s(()=>[n("h3",ue,[y(Y,{"policy-type":m.type},{default:s(()=>[a(c(m.type)+" ("+c(m.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":s(()=>[n("div",ye,[y(p,{class:"policy-type-table",fetcher:()=>({data:m.connections,total:m.connections.length}),headers:[{label:"From",key:"sourceTags"},{label:"To",key:"destinationTags"},{label:"On",key:"name"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}],"cell-attrs":o,"disable-pagination":"","is-clickable":""},{sourceTags:s(({row:r})=>[r.sourceTags.length>0?(e(),i(M,{key:0,class:"tag-list","should-truncate":"",tags:r.sourceTags},null,8,["tags"])):(e(),t(u,{key:1},[a(`
                —
              `)],64))]),destinationTags:s(({row:r})=>[r.destinationTags.length>0?(e(),i(M,{key:0,class:"tag-list","should-truncate":"",tags:r.destinationTags},null,8,["tags"])):(e(),t(u,{key:1},[a(`
                —
              `)],64))]),name:s(({row:r})=>[r.name!==null?(e(),t(u,{key:0},[a(c(r.name),1)],64)):(e(),t(u,{key:1},[a(`
                —
              `)],64))]),origins:s(({row:r})=>[r.origins.length>0?(e(),t("ul",_e,[(e(!0),t(u,null,b(r.origins,(l,g)=>(e(),t("li",{key:`${_}-${g}`},[y(f,{to:{name:"policy-detail-view",params:{mesh:l.mesh,policyPath:d.policyTypesByName[l.type].path,policy:l.name}}},{default:s(()=>[a(c(l.name),1)]),_:2},1032,["to"])]))),128))])):(e(),t(u,{key:1},[a(`
                —
              `)],64))]),config:s(({row:r,rowKey:l})=>[r.config?(e(),i(W,{key:0,id:`${d.id}-${_}-${l}-code-block`,code:T(q)(r.config),language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),t(u,{key:1},[a(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const me=A(de,[["__scopeId","data-v-3e719841"]]),he=h=>(D("data-v-7942f6c8"),h=h(),j(),h),ge={class:"policy-type-heading"},fe={class:"policy-list"},ke={key:0,class:"matcher"},be={key:0,class:"matcher__and"},ve=he(()=>n("br",null,null,-1)),$e={key:1,class:"matcher__not"},Te={class:"matcher__term"},we={key:1},Re={key:0},Pe=I({__name:"RuleEntryList",props:{id:{},ruleEntries:{},policyTypesByName:{},showMatchers:{type:Boolean,default:!0}},setup(h){const{t:d}=H(),o=h;function P({headerKey:v}){return{class:`cell-${v}`}}return(v,f)=>{const p=k("RouterLink"),m=k("KTable");return e(),i(x,{"initially-open":[],"multiple-open":""},{default:s(()=>[(e(!0),t(u,null,b(o.ruleEntries,(_,r)=>(e(),i(V,{key:r},{"accordion-header":s(()=>[n("h3",ge,[y(Y,{"policy-type":_.type},{default:s(()=>[a(c(_.type),1)]),_:2},1032,["policy-type"])])]),"accordion-content":s(()=>[n("div",fe,[y(m,{class:J(["policy-type-table",{"has-matchers":o.showMatchers}]),fetcher:()=>({data:_.rules,total:_.rules.length}),headers:[...o.showMatchers?[{label:"Matchers",key:"matchers"}]:[],{label:"Origin policies",key:"origins"},{label:"Conf",key:"config"}],"cell-attrs":P,"disable-pagination":""},z({origins:s(({row:l})=>[l.origins.length>0?(e(),t("ul",Re,[(e(!0),t(u,null,b(l.origins,(g,w)=>(e(),t("li",{key:`${r}-${w}`},[y(p,{to:{name:"policy-detail-view",params:{mesh:g.mesh,policyPath:o.policyTypesByName[g.type].path,policy:g.name}}},{default:s(()=>[a(c(g.name),1)]),_:2},1032,["to"])]))),128))])):(e(),t(u,{key:1},[a(c(T(d)("common.collection.none")),1)],64))]),config:s(({row:l,rowKey:g})=>[l.config?(e(),i(W,{key:0,id:`${o.id}-${r}-${g}-code-block`,code:T(q)(l.config),language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),t(u,{key:1},[a(c(T(d)("common.collection.none")),1)],64))]),_:2},[o.showMatchers?{name:"matchers",fn:s(({row:l})=>[l.matchers&&l.matchers.length>0?(e(),t("span",ke,[(e(!0),t(u,null,b(l.matchers,({key:g,value:w,not:B},$)=>(e(),t(u,{key:$},[$>0?(e(),t("span",be,[a(" and"),ve])):R("",!0),B?(e(),t("span",$e,"!")):R("",!0),n("span",Te,c(`${g}:${w}`),1)],64))),128))])):(e(),t("i",we,c(T(d)("data-planes.routes.item.matches_everything")),1))]),key:"0"}:void 0]),1032,["class","fetcher","headers"])])]),_:2},1024))),128))]),_:1})}}});const S=A(Pe,[["__scopeId","data-v-7942f6c8"]]),Be={"data-testid":"standard-dataplane-policies",class:"stack"},Le={class:"mb-2"},Ne={key:3},Ce=I({__name:"StandardDataplanePolicies",props:{policyTypeEntries:{},inspectRulesForDataplane:{},policyTypesByName:{},showLegacyPolicies:{type:Boolean}},setup(h){const{t:d}=H(),o=h;return(P,v)=>{const f=k("KCard");return e(),t("div",Be,[(!o.showLegacyPolicies||o.policyTypeEntries.length===0)&&o.inspectRulesForDataplane.rules.length===0?(e(),i(f,{key:0},{default:s(()=>[y(Q)]),_:1})):(e(),t(u,{key:1},[o.inspectRulesForDataplane.proxyRule?(e(),i(f,{key:0},{default:s(()=>[n("h3",null,c(T(d)("data-planes.routes.item.proxy_rule")),1),a(),y(S,{id:"proxy-rules",class:"mt-2","rule-entries":[o.inspectRulesForDataplane.proxyRule],"policy-types-by-name":o.policyTypesByName,"show-matchers":!1,"data-testid":"proxy-rule-list"},null,8,["rule-entries","policy-types-by-name"])]),_:1})):R("",!0),a(),o.inspectRulesForDataplane.toRules.length>0?(e(),i(f,{key:1},{default:s(()=>[n("h3",null,c(T(d)("data-planes.routes.item.to_rules")),1),a(),y(S,{id:"to-rules",class:"mt-2","rule-entries":o.inspectRulesForDataplane.toRules,"policy-types-by-name":o.policyTypesByName,"data-testid":"to-rule-list"},null,8,["rule-entries","policy-types-by-name"])]),_:1})):R("",!0),a(),o.inspectRulesForDataplane.fromRuleInbounds.length>0?(e(),i(f,{key:2},{default:s(()=>[n("h3",Le,c(T(d)("data-planes.routes.item.from_rules")),1),a(),(e(!0),t(u,null,b(o.inspectRulesForDataplane.fromRuleInbounds,(p,m)=>(e(),t("div",{key:m},[n("h4",null,c(T(d)("data-planes.routes.item.port",{port:p.port})),1),a(),y(S,{id:`from-rules-${m}`,class:"mt-2","rule-entries":p.ruleEntries,"policy-types-by-name":o.policyTypesByName,"data-testid":`from-rule-list-${m}`},null,8,["id","rule-entries","policy-types-by-name","data-testid"])]))),128))]),_:1})):R("",!0),a(),o.showLegacyPolicies?(e(),t("div",Ne,[n("h3",null,c(T(d)("data-planes.routes.item.legacy_policies")),1),a(),y(f,{class:"mt-4"},{default:s(()=>[y(me,{id:"policies","policy-type-entries":o.policyTypeEntries,"policy-types-by-name":o.policyTypesByName,"data-testid":"policy-list"},null,8,["policy-type-entries","policy-types-by-name"])]),_:1})])):R("",!0)],64))])}}}),We=I({__name:"DataPlanePoliciesView",props:{data:{}},setup(h){const d=h;return(o,P)=>{const v=k("RouteTitle"),f=k("KCard"),p=k("DataSource"),m=k("AppView"),_=k("RouteView");return e(),i(_,{name:"data-plane-policies-view",params:{mesh:"",dataPlane:""}},{default:s(({can:r,route:l,t:g})=>[y(m,null,{title:s(()=>[n("h2",null,[y(v,{title:g("data-planes.routes.item.navigation.data-plane-policies-view")},null,8,["title"])])]),default:s(()=>[a(),d.data.dataplaneType==="builtin"?(e(),i(p,{key:0,src:"/*/policy-types"},{default:s(({data:w,error:B})=>[y(p,{src:`/meshes/${l.params.mesh}/dataplanes/${l.params.dataPlane}/gateway-dataplane-policies`},{default:s(({data:$,error:N})=>[B?(e(),i(F,{key:0,error:B},null,8,["error"])):N?(e(),i(F,{key:1,error:N},null,8,["error"])):$===void 0||w===void 0?(e(),i(O,{key:2})):(e(),i(f,{key:3},{default:s(()=>[y(re,{"policy-types-by-name":w.policies.reduce((K,C)=>Object.assign(K,{[C.name]:C}),{}),"gateway-dataplane":$},null,8,["policy-types-by-name","gateway-dataplane"])]),_:2},1024))]),_:2},1032,["src"])]),_:2},1024)):(e(),i(p,{key:1,src:"/*/policy-types"},{default:s(({data:w,error:B})=>[y(p,{src:r("use zones")?"":`/meshes/${l.params.mesh}/dataplanes/${l.params.dataPlane}/sidecar-dataplane-policies`},{default:s(({data:$,error:N})=>[y(p,{src:`/meshes/${l.params.mesh}/dataplanes/${l.params.dataPlane}/rules`},{default:s(({data:K,error:C})=>[B?(e(),i(F,{key:0,error:B},null,8,["error"])):N?(e(),i(F,{key:1,error:N},null,8,["error"])):C?(e(),i(F,{key:2,error:C},null,8,["error"])):w===void 0||!r("use zones")&&$===void 0||K===void 0?(e(),i(O,{key:3})):(e(),i(Ce,{key:4,"policy-types-by-name":w.policies.reduce((E,G)=>Object.assign(E,{[G.name]:G}),{}),"policy-type-entries":($==null?void 0:$.policyTypeEntries)??[],"inspect-rules-for-dataplane":K,"show-legacy-policies":!r("use zones")},null,8,["policy-types-by-name","policy-type-entries","inspect-rules-for-dataplane","show-legacy-policies"]))]),_:2},1032,["src"])]),_:2},1032,["src"])]),_:2},1024))]),_:2},1024)]),_:1})}}});export{We as default};
