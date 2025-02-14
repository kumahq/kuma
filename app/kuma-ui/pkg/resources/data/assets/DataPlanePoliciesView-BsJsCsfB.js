import{a as j,A as T}from"./AccordionList-BtF1TsR5.js";import{d as S,r as h,o as e,c as a,m as u,e as n,M as r,N as g,t as d,b as o,w as s,s as N,q as y,T as H,_ as G,p as L,a1 as K,l as Y,B as J,n as Q,af as U,a0 as W,I as Z}from"./index-DGFXSJ_T.js";import{S as E}from"./SummaryView-DDzw2InQ.js";import{P as M}from"./PolicyTypeTag-BcpvHEla.js";import{T as D}from"./TagList-Cp1FLwWy.js";import{R as ee}from"./RuleMatchers-C128ht7H.js";const te={class:"policies-list"},se={class:"mesh-gateway-policy-list"},ae={key:0},ne={class:"dataplane-policy-header"},le={key:0,class:"badge-list"},oe={class:"mt-1"},ie=S({__name:"BuiltinGatewayPolicies",props:{gatewayDataplane:{},types:{}},setup(R){const $=R;return(P,t)=>{const X=h("XAction"),w=h("XBadge");return e(),a("div",te,[u("div",se,[t[11]||(t[11]=u("h3",{class:"mb-2"},`
        Gateway policies
      `,-1)),t[12]||(t[12]=n()),P.gatewayDataplane.routePolicies.length>0?(e(),a("ul",ae,[(e(!0),a(r,null,g(P.gatewayDataplane.routePolicies,(i,m)=>{var p;return e(),a("li",{key:m},[u("span",null,d(i.type),1),t[0]||(t[0]=n(`:

          `)),o(X,{to:{name:"policy-detail-view",params:{mesh:i.mesh,policyPath:((p=$.types[i.type])==null?void 0:p.path)??"",policy:i.name}}},{default:s(()=>[n(d(i.name),1)]),_:2},1032,["to"])])}),128))])):N("",!0),t[13]||(t[13]=n()),t[14]||(t[14]=u("h3",{class:"mt-6 mb-2"},`
        Listeners
      `,-1)),t[15]||(t[15]=n()),u("div",null,[(e(!0),a(r,null,g(P.gatewayDataplane.listenerEntries,(i,m)=>(e(),a("div",{key:m},[u("div",null,[u("div",null,[t[1]||(t[1]=u("b",null,"Host",-1)),n(": "+d(i.hostName)+":"+d(i.port)+" ("+d(i.protocol)+`)
            `,1)]),t[10]||(t[10]=n()),i.routeEntries.length>0?(e(),a(r,{key:0},[t[8]||(t[8]=u("h4",{class:"mt-2 mb-2"},`
                Routes
              `,-1)),t[9]||(t[9]=n()),o(T,{"initially-open":[],"multiple-open":""},{default:s(()=>[(e(!0),a(r,null,g(i.routeEntries,(p,b)=>(e(),y(j,{key:b},H({"accordion-header":s(()=>{var l;return[u("div",ne,[u("div",null,[u("div",null,[t[2]||(t[2]=u("b",null,"Route",-1)),t[3]||(t[3]=n(": ")),o(X,{to:{name:"policy-detail-view",params:{mesh:p.route.mesh,policyPath:((l=$.types[p.route.type])==null?void 0:l.path)??"",policy:p.route.name}}},{default:s(()=>[n(d(p.route.name),1)]),_:2},1032,["to"])]),t[5]||(t[5]=n()),u("div",null,[t[4]||(t[4]=u("b",null,"Service",-1)),n(": "+d(p.service),1)])]),t[6]||(t[6]=n()),p.origins.length>0?(e(),a("div",le,[(e(!0),a(r,null,g(p.origins,(_,C)=>(e(),y(w,{key:`${m}-${C}`},{default:s(()=>[n(d(_.type),1)]),_:2},1024))),128))])):N("",!0)])]}),_:2},[p.origins.length>0?{name:"accordion-content",fn:s(()=>[u("ul",oe,[(e(!0),a(r,null,g(p.origins,(l,_)=>{var C;return e(),a("li",{key:`${m}-${_}`},[n(d(l.type)+`:

                        `,1),o(X,{to:{name:"policy-detail-view",params:{mesh:l.mesh,policyPath:((C=$.types[l.type])==null?void 0:C.path)??"",policy:l.name}}},{default:s(()=>[n(d(l.name),1)]),_:2},1032,["to"])])}),128))])]),key:"0"}:void 0]),1024))),128))]),_:2},1024)],64)):N("",!0)])]))),128))])])])}}}),re=G(ie,[["__scopeId","data-v-c6fda729"]]),pe={class:"policy-type-heading"},ue={class:"policy-list"},de={key:0},me=S({__name:"PolicyTypeEntryList",props:{items:{},types:{}},setup(R){const $=R;function P({headerKey:t}){return{class:`cell-${t}`}}return(t,X)=>{const w=h("XAction"),i=h("XCodeBlock"),m=h("KTable");return e(),y(T,{"initially-open":[],"multiple-open":""},{default:s(()=>[(e(!0),a(r,null,g(t.items,(p,b)=>(e(),y(j,{key:b},{"accordion-header":s(()=>[u("h3",pe,[o(M,{"policy-type":p.type},{default:s(()=>[n(d(p.type)+" ("+d(p.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":s(()=>[u("div",ue,[o(m,{class:"policy-type-table",fetcher:()=>({data:p.connections,total:p.connections.length}),headers:[{label:"From",key:"sourceTags"},{label:"To",key:"destinationTags"},{label:"On",key:"name"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}],"cell-attrs":P,"disable-pagination":"","is-clickable":""},{sourceTags:s(({row:l})=>[l.sourceTags.length>0?(e(),y(D,{key:0,class:"tag-list","should-truncate":"",tags:l.sourceTags},null,8,["tags"])):(e(),a(r,{key:1},[n(`
                —
              `)],64))]),destinationTags:s(({row:l})=>[l.destinationTags.length>0?(e(),y(D,{key:0,class:"tag-list","should-truncate":"",tags:l.destinationTags},null,8,["tags"])):(e(),a(r,{key:1},[n(`
                —
              `)],64))]),name:s(({row:l})=>[l.name!==null?(e(),a(r,{key:0},[n(d(l.name),1)],64)):(e(),a(r,{key:1},[n(`
                —
              `)],64))]),origins:s(({row:l})=>[l.origins.length>0?(e(),a("ul",de,[(e(!0),a(r,null,g(l.origins,(_,C)=>{var c;return e(),a("li",{key:`${b}-${C}`},[o(w,{to:{name:"policy-detail-view",params:{mesh:_.mesh,policyPath:((c=$.types[_.type])==null?void 0:c.path)??"",policy:_.name}}},{default:s(()=>[n(d(_.name),1)]),_:2},1032,["to"])])}),128))])):(e(),a(r,{key:1},[n(`
                —
              `)],64))]),config:s(({row:l})=>[l.config?(e(),y(i,{key:0,code:L(K).stringify(l.config),language:"yaml","show-copy-button":!1},null,8,["code"])):(e(),a(r,{key:1},[n(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}}),ce=G(me,[["__scopeId","data-v-53f1d6e8"]]),ye={class:"policy-type-heading"},_e={key:0,class:"matcher"},fe={key:1},ge={key:0},ke=S({__name:"RuleList",props:{rules:{},types:{}},setup(R){const{t:$}=Y(),P=R;return(t,X)=>{const w=h("XCodeBlock");return e(),y(T,{"initially-open":[],"multiple-open":""},{default:s(()=>[(e(!0),a(r,null,g([P.rules.reduce((i,m)=>(typeof i[m.type]>"u"&&(i[m.type]=[]),i[m.type].push(m),i),{})],i=>(e(),a(r,{key:i},[(e(!0),a(r,null,g(i,(m,p)=>(e(),y(j,{key:p},{"accordion-header":s(()=>[u("h3",ye,[o(M,{"policy-type":p},{default:s(()=>[n(d(p),1)]),_:2},1032,["policy-type"])])]),"accordion-content":s(()=>[(e(!0),a(r,null,g([m.some(b=>b.matchers.length>0)],b=>(e(),a("div",{key:b,class:"policy-list"},[o(J,{class:Q(["policy-type-table",{"has-matchers":b}]),items:m,headers:[...b?[{label:"Matchers",key:"matchers"}]:[],{label:"Origin policies",key:"origins"},{label:"Conf",key:"config"}]},{matchers:s(({row:l})=>[l.matchers.length>0?(e(),a("span",_e,[o(ee,{items:l.matchers},null,8,["items"])])):(e(),a("i",fe,d(L($)("data-planes.routes.item.matches_everything")),1))]),origins:s(({row:l})=>[l.origins.length>0?(e(),a("ul",ge,[(e(!0),a(r,null,g(l.origins,(_,C)=>{var c;return e(),a("li",{key:`${p}-${C}`},[o(U,{to:{name:"data-plane-policy-summary-view",params:{mesh:_.mesh,policyPath:((c=P.types[_.type])==null?void 0:c.path)??"",policy:_.name}}},{default:s(()=>[n(d(_.name),1)]),_:2},1032,["to"])])}),128))])):(e(),a(r,{key:1},[n(d(L($)("common.collection.none")),1)],64))]),config:s(({row:l})=>[Object.keys(l.raw).length>0?(e(),y(w,{key:0,code:L(K).stringify(l.raw),language:"yaml","show-copy-button":!1},null,8,["code"])):(e(),a(r,{key:1},[n(d(L($)("common.collection.none")),1)],64))]),_:2},1032,["class","items","headers"])]))),128))]),_:2},1024))),128))],64))),128))]),_:1})}}}),F=G(ke,[["__scopeId","data-v-39d2dd11"]]),he={class:"stack"},be={class:"mb-2"},ve={key:0},Re=S({__name:"DataPlanePoliciesView",props:{data:{}},setup(R){const $=R;return(P,t)=>{const X=h("RouteTitle"),w=h("XCard"),i=h("DataCollection"),m=h("DataLoader"),p=h("DataSource"),b=h("RouterView"),l=h("AppView"),_=h("RouteView");return e(),y(_,{name:"data-plane-policies-view",params:{mesh:"",proxy:""}},{default:s(({can:C,route:c,t:V,uri:q})=>[o(X,{render:!1,title:V("data-planes.routes.item.navigation.data-plane-policies-view")},null,8,["title"]),t[8]||(t[8]=n()),o(l,null,{default:s(()=>[u("div",he,[o(p,{src:"/policy-types"},{default:s(({data:B,error:O})=>[(e(!0),a(r,null,g([((B==null?void 0:B.policies)??[]).reduce((A,v)=>Object.assign(A,{[v.name]:v}),{})],A=>(e(),a(r,{key:typeof A},[o(m,{src:q(L(W),"/meshes/:mesh/rules/for/:dataplane",{mesh:c.params.mesh,dataplane:c.params.proxy}),data:[B],errors:[O]},{default:s(({data:v})=>[o(i,{items:v.rules},{default:s(()=>[(e(),a(r,null,g(["proxy","to"],k=>o(i,{key:k,items:v.rules,predicate:f=>f.ruleType===k,comparator:(f,x)=>f.type.localeCompare(x.type),empty:!1},{default:s(({items:f})=>[o(w,null,{default:s(()=>[u("h3",null,d(V(`data-planes.routes.item.rules.${k}`)),1),t[0]||(t[0]=n()),o(F,{class:"mt-2",rules:f,types:A,"data-testid":`${k}-rule-list`},null,8,["rules","types","data-testid"])]),_:2},1024)]),_:2},1032,["items","predicate","comparator"])),64)),t[3]||(t[3]=n()),o(i,{items:v.rules,predicate:k=>k.ruleType==="from",comparator:(k,f)=>k.type.localeCompare(f.type),empty:!1},{default:s(({items:k})=>[o(w,null,{default:s(()=>[u("h3",be,d(V("data-planes.routes.item.rules.from")),1),t[2]||(t[2]=n()),(e(!0),a(r,null,g([Object.groupBy(k,f=>f.inbound.port)],f=>(e(),a(r,{key:f},[(e(!0),a(r,null,g(Object.entries(f).sort(([x],[I])=>Number(I)-Number(x)),([x,I],z)=>(e(),a("div",{key:z},[u("h4",null,d(V("data-planes.routes.item.port",{port:x})),1),t[1]||(t[1]=n()),o(F,{class:"mt-2",rules:I,types:A,"data-testid":`from-rule-list-${z}`},null,8,["rules","types","data-testid"])]))),128))],64))),128))]),_:2},1024)]),_:2},1032,["items","predicate","comparator"])]),_:2},1032,["items"])]),_:2},1032,["src","data","errors"]),t[6]||(t[6]=n()),C("use zones")?N("",!0):(e(),a("div",ve,[$.data.dataplaneType==="builtin"?(e(),y(m,{key:0,src:`/meshes/${c.params.mesh}/dataplanes/${c.params.proxy}/gateway-dataplane-policies`,data:[B],errors:[O]},{default:s(({data:v})=>[v?(e(),y(i,{key:0,items:v.routePolicies,empty:!1},{default:s(()=>[u("h3",null,d(V("data-planes.routes.item.legacy_policies")),1),t[4]||(t[4]=n()),o(w,{class:"mt-4"},{default:s(()=>[o(re,{types:A,"gateway-dataplane":v,"data-testid":"builtin-gateway-dataplane-policies"},null,8,["types","gateway-dataplane"])]),_:2},1024)]),_:2},1032,["items"])):N("",!0)]),_:2},1032,["src","data","errors"])):(e(),y(m,{key:1,src:`/meshes/${c.params.mesh}/dataplanes/${c.params.proxy}/sidecar-dataplane-policies`,data:[B],errors:[O]},{default:s(({data:v})=>[o(i,{predicate:k=>{var f;return((f=A[k.type])==null?void 0:f.isTargetRefBased)===!1},items:v.policyTypeEntries,empty:!1},{default:s(({items:k})=>[u("h3",null,d(V("data-planes.routes.item.legacy_policies")),1),t[5]||(t[5]=n()),o(w,{class:"mt-4"},{default:s(()=>[o(ce,{items:k,types:A,"data-testid":"sidecar-dataplane-policies"},null,8,["items","types"])]),_:2},1024)]),_:2},1032,["predicate","items"])]),_:2},1032,["src","data","errors"]))]))],64))),128))]),_:2},1024)]),t[7]||(t[7]=n()),o(b,null,{default:s(({Component:B})=>[c.child()?(e(),y(E,{key:0,onClose:O=>c.replace({name:"data-plane-policies-view",params:{mesh:c.params.mesh,proxy:c.params.proxy}})},{default:s(()=>[(e(),y(Z(B)))]),_:2},1032,["onClose"])):N("",!0)]),_:2},1024)]),_:2},1024)]),_:1})}}});export{Re as default};
